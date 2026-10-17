import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
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

// Мок только того, что тестируем; MemoryRouter остаётся реальным,
// чтобы Link/роутер работали как в проде.
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: vi.fn() };
});

const mockLogin = vi.mocked(login);

const loginResponse: LoginResponse = {
  access_token: 'tok-123',
  user: { id: 'u1', email: 'a@b.c', name: 'A', created_at: '2024-01-01T00:00:00Z' },
};

// Рендерим внутри QueryClientProvider как в App.tsx — иначе useQueryClient()
// в Login упадёт с "No QueryClient set".
function renderLogin() {
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function fillAndSubmit() {
  const emailInput = screen.getByLabelText('Email') as HTMLInputElement;
  const passwordInput = screen.getByLabelText('Пароль') as HTMLInputElement;
  emailInput.value = 'a@b.c';
  emailInput.dispatchEvent(new Event('input', { bubbles: true }));
  passwordInput.value = 'password123';
  passwordInput.dispatchEvent(new Event('input', { bubbles: true }));
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
    fillAndSubmit().click();

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ email: 'a@b.c', password: 'password123' });
    });
    expect(setAccessToken).toHaveBeenCalledWith('tok-123');
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
    fillAndSubmit().click();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });
});
