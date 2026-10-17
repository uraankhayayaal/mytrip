import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { login } from '../api/auth';
import { setAccessToken } from '../api/client';
import { ApiError } from '../api/errors';
import type { LoginResponse } from '../lib/types';
import { Login } from './Login';

vi.mock('../api/auth', () => ({
  login: vi.fn(),
}));

vi.mock('../api/client', () => ({
  setAccessToken: vi.fn(),
}));

// Мок только useNavigate; MemoryRouter/Link остаются реальными,
// чтобы роутер работал как в проде.
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: vi.fn() };
});

// Корректный паттерн: vi.mocked требует уже импортированный объект (не require).
const mockLogin = vi.mocked(login);
const mockSetAccessToken = vi.mocked(setAccessToken);

const loginResponse: LoginResponse = {
  access_token: 'tok-123',
  user: { id: 'u1', email: 'a@b.c', name: 'A', created_at: '2024-01-01T00:00:00Z' },
};

// QueryClientProvider обязателен: QueryClient используется в приложении (App.tsx),
// рендерим под провайдером как в проде.
function renderLogin() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

// react-hook-form читает значения из onChange-обработчиков, поэтому заполняем
// поля fireEvent.change (не dispatchEvent(new Event('input'))).
function fillAndSubmit() {
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'a@b.c' } });
  fireEvent.change(screen.getByLabelText('Пароль'), { target: { value: 'password123' } });
  return screen.getByRole('button', { name: 'Войти' });
}

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
    renderLogin();

    const emailInput = screen.getByLabelText('Email');
    const passwordInput = screen.getByLabelText('Пароль');
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });

    // DEBUG: что видит React после fireEvent.change
    console.log('EMAIL VALUE AFTER CHANGE:', (emailInput as HTMLInputElement).value);
    console.log('EMAIL ATTR:', (emailInput as HTMLInputElement).getAttribute('value'));

    screen.getByRole('button', { name: 'Войти' }).click();

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ email: 'a@b.c', password: 'password123' });
    });
    expect(mockSetAccessToken).toHaveBeenCalledWith('tok-123');
    expect(window.location.href).toBe('/');
  });

  it('пустые поля: ошибки zod, запрос не уходит', async () => {
    renderLogin();
    screen.getByRole('button', { name: 'Войти' }).click();

    await waitFor(() => {
      expect(screen.getByText(/Invalid email/i)).toBeInTheDocument();
      expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    });
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('ApiError unauthorized: баннер «Неверный email или пароль»', async () => {
    mockLogin.mockRejectedValue(new ApiError('unauthorized', 'bad credentials'));
    renderLogin();

    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'a@b.c' } });
    fireEvent.change(screen.getByLabelText('Пароль'), { target: { value: 'password123' } });
    screen.getByRole('button', { name: 'Войти' }).click();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });
});
