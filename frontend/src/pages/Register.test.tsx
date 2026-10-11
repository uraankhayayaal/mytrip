import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { register } from '../api/auth';
import { Register } from './Register';

vi.mock('../api/auth', () => ({
  register: vi.fn(),
}));

function renderRegister() {
  return render(
    <MemoryRouter>
      <Register />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(window, 'location', {
    value: { href: window.location.href },
    writable: true,
  });
});

describe('Register', () => {
  it('валидная форма вызывает register, редирект на /login', async () => {
    vi.mocked(register).mockResolvedValue({
      id: 'u1',
      email: 'a@b.c',
      name: 'A',
      created_at: '',
    });
    renderRegister();
    const user = userEvent.setup();
    await user.type(screen.getByLabelText('Имя'), 'Анна');
    await user.type(screen.getByLabelText('Email'), 'a@b.c');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByRole('button', { name: 'Зарегистрироваться' }));
    await vi.waitFor(() => {
      expect(register).toHaveBeenCalledWith({
        email: 'a@b.c',
        password: 'password123',
        name: 'Анна',
      });
    });
    expect(window.location.href).toBe('/login');
  });

  it('password < 8 символов: ошибка zod, запрос не уходит', async () => {
    renderRegister();
    const user = userEvent.setup();
    await user.type(screen.getByLabelText('Имя'), 'Анна');
    await user.type(screen.getByLabelText('Email'), 'a@b.c');
    await user.type(screen.getByLabelText('Пароль'), 'short');
    await user.click(screen.getByRole('button', { name: 'Зарегистрироваться' }));
    expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    expect(register).not.toHaveBeenCalled();
  });
});
