import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { login } from '../api/auth';
import { Login } from './Login';

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/react-query')>(
    '@tanstack/react-query',
  );
  return {
    ...actual,
    useQuery: vi.fn(),
    useQueryClient: vi.fn(() => new QueryClient()),
  };
});

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: vi.fn(),
    useSearchParams: vi.fn(() => [new URLSearchParams(), vi.fn()]),
  };
});

vi.mock('../api/auth', () => ({
  login: vi.fn(),
  ApiError: class ApiError extends Error {
    readonly body: { error: { code: string; message: string } };
    constructor(body: { error: { code: string; message: string } }) {
      super(body.error.message);
      this.body = body;
    }
  },
}));

const mockUseNavigate = vi.mocked(useNavigate);
const mockLogin = vi.mocked(login);

function renderWithProvider(ui: ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn());
});

describe('Login', () => {
  it('валидная форма вызывает login и setAccessToken, редирект на /', async () => {
    const token = {
      access_token: 'tok123',
      user: { id: 'u1', email: 'a@b.c', name: 'User', created_at: '2024-01-01T00:00:00Z' },
    };
    mockLogin.mockResolvedValue(token);
    const navigate = vi.fn();
    mockUseNavigate.mockReturnValue(navigate);

    renderWithProvider(<Login />);
    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    const passInput = screen.getByLabelText(/пароль/i) as HTMLInputElement;
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(passInput, { target: { value: 'pass123' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ email: 'a@b.c', password: 'pass123' });
    });
    expect(localStorage.getItem('access_token')).toBe('tok123');
    expect(navigate).toHaveBeenCalledWith('/');
  });

  it('ApiError unauthorized: баннер «Неверный email или пароль»', async () => {
    const { ApiError: AErr } = await import('../api/auth');
    const err = new AErr({ code: 'unauthorized', message: 'bad creds' });
    mockLogin.mockRejectedValue(err);
    mockUseNavigate.mockReturnValue(vi.fn());

    renderWithProvider(<Login />);
    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    const passInput = screen.getByLabelText(/пароль/i) as HTMLInputElement;
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(passInput, { target: { value: 'wrong' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });
});
