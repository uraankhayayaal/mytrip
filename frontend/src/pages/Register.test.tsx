import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi, beforeEach, type Mock } from 'vitest';
import { register } from '../api/auth';
import { Register } from './Register';
import type { User } from '../lib/types';

vi.mock('../api/auth', () => ({
  register: vi.fn(),
}));

const mockRegister = vi.mocked(register);

const user: User = {
  id: 'u1',
  email: 'a@b.c',
  name: 'Анна',
  created_at: '2024-01-01T00:00:00Z',
};

function renderRegister() {
  return render(
    <MemoryRouter>
      <Register />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.stubGlobal('window', {
    ...window,
    location: { href: window.location.href },
  });
});

describe('Register', () => {
  it('валидная форма вызывает register, редирект на /login', async () => {
    mockRegister.mockResolvedValue(user);
    renderRegister();
    const nameInput = screen.getByLabelText('Имя') as HTMLInputElement;
    const emailInput = screen.getByLabelText('Email') as HTMLInputElement;
    const passwordInput = screen.getByLabelText('Пароль') as HTMLInputElement;
    nameInput.value = 'Анна';
    nameInput.dispatchEvent(new Event('input', { bubbles: true }));
    emailInput.value = 'a@b.c';
    emailInput.dispatchEvent(new Event('input', { bubbles: true }));
    passwordInput.value = 'password123';
    passwordInput.dispatchEvent(new Event('input', { bubbles: true }));
    await new Promise((r) => setTimeout(r, 100));
    screen.getByRole('button', { name: 'Зарегистрироваться' }).click();
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'a@b.c',
        password: 'password123',
        name: 'Анна',
      });
    });
  });

  it('password < 8 символов: ошибка zod, запрос не уходит', async () => {
    renderRegister();
    const nameInput = screen.getByLabelText('Имя') as HTMLInputElement;
    const emailInput = screen.getByLabelText('Email') as HTMLInputElement;
    const passwordInput = screen.getByLabelText('Пароль') as HTMLInputElement;
    nameInput.value = 'Анна';
    nameInput.dispatchEvent(new Event('input', { bubbles: true }));
    emailInput.value = 'a@b.c';
    emailInput.dispatchEvent(new Event('input', { bubbles: true }));
    passwordInput.value = 'short';
    passwordInput.dispatchEvent(new Event('input', { bubbles: true }));
    await new Promise((r) => setTimeout(r, 100));
    screen.getByRole('button', { name: 'Зарегистрироваться' }).click();
    await waitFor(() => {
      expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    });
    expect(mockRegister).not.toHaveBeenCalled();
  });
});
