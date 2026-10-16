import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi, beforeEach, type Mock } from 'vitest';
import { register } from '../api/auth';
import { Register } from './Register';
import type { User } from '../lib/types';

vi.mock('../api/auth', () => ({
  register: vi.fn(),
}));

const mockRegister = vi.mocked(register) as Mock;

const user: User = {
  id: 'u1',
  email: 'a@b.c',
  name: 'Анна',
  created_at: '2024-01-01T00:00:00Z',
};

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(window, 'location', {
    value: { href: window.location.href },
    writable: true,
    configurable: true,
  });
});

describe('Register', () => {
  it('валидная форма вызывает register, редирект на /login', async () => {
    mockRegister.mockResolvedValue(user);
    render(<Register />);
    const user = userEvent.setup();
    await user.type(screen.getByLabelText('Имя'), 'Анна');
    await user.type(screen.getByLabelText('Email'), 'a@b.c');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByRole('button', { name: 'Зарегистрироваться' }));
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'a@b.c',
        password: 'password123',
        name: 'Анна',
      });
    });
    expect(window.location.href).toBe('/login');
  });

  it('password < 8 символов: ошибка zod, запрос не уходит', async () => {
    render(<Register />);
    const user = userEvent.setup();
    await user.type(screen.getByLabelText('Имя'), 'Анна');
    await user.type(screen.getByLabelText('Email'), 'a@b.c');
    await user.type(screen.getByLabelText('Пароль'), 'short');
    await user.click(screen.getByRole('button', { name: 'Зарегистрироваться' }));
    await waitFor(() => {
      expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    });
    expect(mockRegister).not.toHaveBeenCalled();
  });
});
