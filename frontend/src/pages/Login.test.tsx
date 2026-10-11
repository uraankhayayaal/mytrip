import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { login } from '../api/auth';
import { setAccessToken } from '../api/client';
import { ApiError } from '../api/errors';
import { Login } from './Login';

vi.mock('../api/auth', () => ({
  login: vi.fn(),
}));

vi.mock('../api/client', () => ({
  setAccessToken: vi.fn(),
}));

function renderLogin() {
  return render(
    <MemoryRouter>
      <Login />
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

describe('Login', () => {
  it('валидная форма вызывает login и setAccessToken, редирект на /', async () => {
    vi.mocked(login).mockResolvedValue({
      access_token: 'tok-123',
      user: { id: 'u1', email: 'a@b.c', name: 'A', created_at: '' },
    });
    renderLogin();
    const user = userEvent.setup();
    await user.type(screen.getByLabelText('Email'), 'a@b.c');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByRole('button', { name: 'Войти' }));
    await vi.waitFor(() => {
      expect(login).toHaveBeenCalledWith({ email: 'a@b.c', password: 'password123' });
    });
    expect(setAccessToken).toHaveBeenCalledWith('tok-123');
    expect(window.location.href).toBe('/');
  });

  it('пустые поля: ошибки zod, запрос не уходит', async () => {
    renderLogin();
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: 'Войти' }));
    expect(screen.getByText(/Invalid email/i)).toBeInTheDocument();
    expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    expect(login).not.toHaveBeenCalled();
  });

  it('ApiError unauthorized: баннер «Неверный email или пароль»', async () => {
    vi.mocked(login).mockRejectedValue(new ApiError('unauthorized', 'bad credentials'));
    renderLogin();
    const user = userEvent.setup();
    await user.type(screen.getByLabelText('Email'), 'a@b.c');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByRole('button', { name: 'Войти' }));
    await vi.waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });
});
