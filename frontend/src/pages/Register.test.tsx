import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { register } from '../api/auth';
import { ApiError } from '../api/errors';
import { Register } from './Register';
import type { User } from '../lib/types';

vi.mock('react-router-dom', () => ({
  useNavigate: vi.fn(),
  useSearchParams: vi.fn(() => [new URLSearchParams(), vi.fn()]),
  useParams: vi.fn(() => ({})),
  Link: ({ children }: { children?: React.ReactNode }) => <a>{children}</a>,
  useLocation: vi.fn(() => ({ pathname: '/' })),
}));

vi.mock('../api/auth', () => ({
  register: vi.fn(),
}));

// vi.mocked применяется к УЖЕ ИМПОРТИРОВАННОМУ объекту
const mockUseNavigate = vi.mocked(useNavigate);
const mockRegister = vi.mocked(register);

function renderWithProvider(ui: ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn());
  mockRegister.mockResolvedValue({ id: 'u2', email: 'new@b.c', name: 'New', created_at: '2024-01-01T00:00:00Z' });
});

describe('Register', () => {
  it('валидная форма вызывает register с полным payload', async () => {
    const user: User = { id: 'u2', email: 'new@b.c', name: 'New', created_at: '2024-01-01T00:00:00Z' };
    mockRegister.mockResolvedValue(user);
    const navigate = vi.fn();
    mockUseNavigate.mockReturnValue(navigate);

    renderWithProvider(<Register />);
    const emailInput = screen.getByLabelText('Email');
    const nameInput = screen.getByLabelText('Имя');
    const passInput = screen.getByLabelText('Пароль');
    fireEvent.change(emailInput, { target: { value: 'new@b.c' } });
    fireEvent.change(nameInput, { target: { value: 'New User' } });
    fireEvent.change(passInput, { target: { value: 'password1' } });
    fireEvent.click(screen.getByRole('button', { name: /зарегистрироваться/i }));

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'new@b.c',
        name: 'New User',
        password: 'password1',
      });
    });
  });

  it('ApiError conflict: баннер о существующем пользователе', async () => {
    const err = new ApiError('conflict', 'exists');
    mockRegister.mockRejectedValue(err);
    mockUseNavigate.mockReturnValue(vi.fn());

    renderWithProvider(<Register />);
    const emailInput = screen.getByLabelText('Email');
    const nameInput = screen.getByLabelText('Имя');
    const passInput = screen.getByLabelText('Пароль');
    fireEvent.change(emailInput, { target: { value: 'dup@b.c' } });
    fireEvent.change(nameInput, { target: { value: 'Dup' } });
    fireEvent.change(passInput, { target: { value: 'password1' } });
    fireEvent.click(screen.getByRole('button', { name: /зарегистрироваться/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Пользователь с таким email уже существует');
    });
  });

  it('ApiError validation: баннер «Проверьте правильность заполнения полей»', async () => {
    const err = new ApiError('validation', 'bad');
    mockRegister.mockRejectedValue(err);
    mockUseNavigate.mockReturnValue(vi.fn());

    renderWithProvider(<Register />);
    const emailInput = screen.getByLabelText('Email');
    const nameInput = screen.getByLabelText('Имя');
    const passInput = screen.getByLabelText('Пароль');
    fireEvent.change(emailInput, { target: { value: 'v@b.c' } });
    fireEvent.change(nameInput, { target: { value: 'V' } });
    fireEvent.change(passInput, { target: { value: 'password1' } });
    fireEvent.click(screen.getByRole('button', { name: /зарегистрироваться/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Проверьте правильность заполнения полей');
    });
  });
});
