import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi, beforeEach, type Mock } from 'vitest';
import { login } from '../api/auth';
import { setAccessToken } from '../api/client';
import { ApiError } from '../api/errors';
import { Login } from './Login';
import type { LoginResponse } from '../lib/types';

vi.mock('../api/auth', () => ({
  login: vi.fn(),
}));

vi.mock('../api/client', () => ({
  setAccessToken: vi.fn(),
}));

const mockLogin = vi.mocked(login) as Mock;

const loginResponse: LoginResponse = {
  access_token: 'tok-123',
  user: { id: 'u1', email: 'a@b.c', name: 'A', created_at: '2024-01-01T00:00:00Z' },
};

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(window, 'location', {
    value: { href: window.location.href },
    writable: true,
    configurable: true,
  });
});

describe('Login', () => {
  it('валидная форма вызывает login и setAccessToken, редирект на /', async () => {
    mockLogin.mockResolvedValue(loginResponse);
    render(<Login />);
    const inputs = screen.getAllByRole('textbox');
    inputs[0].value = 'a@b.c';
    inputs[1].value = 'password123';
    screen.getByRole('button', { name: 'Войти' }).click();
    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ email: 'a@b.c', password: 'password123' });
    });
    expect(setAccessToken).toHaveBeenCalledWith('tok-123');
    expect(window.location.href).toBe('/');
  });

  it('пустые поля: ошибки zod, запрос не уходит', async () => {
    render(<Login />);
    screen.getByRole('button', { name: 'Войти' }).click();
    await waitFor(() => {
      expect(screen.getByText(/Invalid email/i)).toBeInTheDocument();
      expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    });
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('ApiError unauthorized: баннер «Неверный email или пароль»', async () => {
    mockLogin.mockRejectedValue(new ApiError('unauthorized', 'bad credentials'));
    render(<Login />);
    const inputs = screen.getAllByRole('textbox');
    inputs[0].value = 'a@b.c';
    inputs[1].value = 'password123';
    screen.getByRole('button', { name: 'Войти' }).click();
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });
});
