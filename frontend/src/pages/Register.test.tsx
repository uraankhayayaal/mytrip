import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { register } from '../api/auth';
import { Register } from './Register';

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
  register: vi.fn(),
  ApiError: class ApiError extends Error {
    readonly body: { error: { code: string; message: string } };
    constructor(body: { error: { code: string; message: string } }) {
      super(body.error.message);
      this.body = body;
    }
  },
}));

const mockUseNavigate = vi.mocked(useNavigate);
const mockRegister = vi.mocked(register);

function renderWithProvider(ui: ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn());
});

describe('Register', () => {
  it('валидная форма вызывает register с полным payload', async () => {
    const token = {
      access_token: 'tok456',
      user: { id: 'u2', email: 'new@b.c', name: 'New', created_at: '2024-01-01T00:00:00Z' },
    };
    mockRegister.mockResolvedValue(token);
    const navigate = vi.fn();
    mockUseNavigate.mockReturnValue(navigate);

    renderWithProvider(<Register />);
    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    const nameInput = screen.getByLabelText(/имя/i) as HTMLInputElement;
    const passInput = screen.getByLabelText(/пароль/i) as HTMLInputElement;
    const pass2 = screen.getByLabelText(/повторите пароль/i) as HTMLInputElement;
    fireEvent.change(emailInput, { target: { value: 'new@b.c' } });
    fireEvent.change(nameInput, { target: { value: 'New User' } });
    fireEvent.change(passInput, { target: { value: 'password123' } });
    fireEvent.change(pass2, { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: /зарегистрироваться/i }));

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'new@b.c',
        name: 'New User',
        password: 'password123',
      });
    });
    expect(navigate).toHaveBeenCalledWith('/');
  });

  it('ApiError conflict: баннер о существующем пользователе', async () => {
    const { ApiError: AErr } = await import('../api/auth');
    const err = new AErr({ code: 'conflict', message: 'exists' });
    mockRegister.mockRejectedValue(err);
    mockUseNavigate.mockReturnValue(vi.fn());

    renderWithProvider(<Register />);
    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    const nameInput = screen.getByLabelText(/имя/i) as HTMLInputElement;
    const passInput = screen.getByLabelText(/пароль/i) as HTMLInputElement;
    const pass2 = screen.getByLabelText(/повторите пароль/i) as HTMLInputElement;
    fireEvent.change(emailInput, { target: { value: 'dup@b.c' } });
    fireEvent.change(nameInput, { target: { value: 'Dup' } });
    fireEvent.change(passInput, { target: { value: 'password123' } });
    fireEvent.change(pass2, { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: /зарегистрироваться/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Пользователь с таким email уже существует');
    });
  });
});
