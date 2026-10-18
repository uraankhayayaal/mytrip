import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { login } from '../api/auth';
import { ApiError } from '../api/errors';
import { Login } from './Login';
import type { LoginResponse } from '../lib/types';

vi.mock('react-router-dom', () => ({
  useNavigate: vi.fn(),
  useSearchParams: vi.fn(() => [new URLSearchParams(), vi.fn()]),
  useParams: vi.fn(() => ({})),
  Link: ({ children }: { children?: React.ReactNode }) => <a>{children}</a>,
  useLocation: vi.fn(() => ({ pathname: '/' })),
}));

vi.mock('../api/auth', () => ({
  login: vi.fn(),
}));

// vi.mocked применяется к УЖЕ ИМПОРТИРОВАННОМУ объекту
const mockUseNavigate = vi.mocked(useNavigate);
const mockLogin = vi.mocked(login);

function renderWithProvider(ui: ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn());
  mockLogin.mockResolvedValue({
    access_token: 'tok123',
    user: { id: 'u1', email: 'a@b.c', name: 'User', created_at: '2024-01-01T00:00:00Z' },
  });
});

describe('Login', () => {
  it('валидная форма вызывает login и setAccessToken', async () => {
    const token: LoginResponse = {
      access_token: 'tok123',
      user: { id: 'u1', email: 'a@b.c', name: 'User', created_at: '2024-01-01T00:00:00Z' },
    };
    mockLogin.mockResolvedValue(token);
    const navigate = vi.fn();
    mockUseNavigate.mockReturnValue(navigate);

    renderWithProvider(<Login />);
    const emailInput = screen.getByLabelText('Email');
    const passInput = screen.getByLabelText('Пароль');
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(passInput, { target: { value: 'password1' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ email: 'a@b.c', password: 'password1' });
    });
    expect(localStorage.getItem('access_token')).toBe('tok123');
  });

  it('ApiError unauthorized: баннер «Неверный email или пароль»', async () => {
    const err = new ApiError('unauthorized', 'bad creds');
    mockLogin.mockRejectedValue(err);
    mockUseNavigate.mockReturnValue(vi.fn());

    renderWithProvider(<Login />);
    const emailInput = screen.getByLabelText('Email');
    const passInput = screen.getByLabelText('Пароль');
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(passInput, { target: { value: 'wrong' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });

  it('ApiError validation: баннер «Проверьте правильность заполнения полей»', async () => {
    const err = new ApiError('validation', 'invalid');
    mockLogin.mockRejectedValue(err);
    mockUseNavigate.mockReturnValue(vi.fn());

    renderWithProvider(<Login />);
    const emailInput = screen.getByLabelText('Email');
    const passInput = screen.getByLabelText('Пароль');
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(passInput, { target: { value: 'password1' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Проверьте правильность заполнения полей');
    });
  });

  it('не-ApiError: баннер «Не удалось войти. Попробуйте позже.»', async () => {
    mockLogin.mockRejectedValue(new Error('network'));
    mockUseNavigate.mockReturnValue(vi.fn());

    renderWithProvider(<Login />);
    const emailInput = screen.getByLabelText('Email');
    const passInput = screen.getByLabelText('Пароль');
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(passInput, { target: { value: 'password1' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Не удалось войти. Попробуйте позже.');
    });
  });
});
