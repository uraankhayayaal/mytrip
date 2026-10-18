import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { login, ApiError } from '../api/auth';
import { TripsList } from './TripsList';
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
  ApiError: class ApiError extends Error {},
}));

const mockUseQuery = vi.mocked(useQuery);
const mockUseNavigate = vi.mocked(useNavigate);
const mockLogin = vi.mocked(login);

function renderWithProvider(ui: ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn());
  mockUseQuery.mockReturnValue({ data: undefined, isLoading: false });
});

describe('TripsList', () => {
  it('renders a list of trips', () => {
    const trips = [
      { id: '1', title: 'Trip 1', description: 'd', start_date: '2024-01-01', end_date: '2024-01-02', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      { id: '2', title: 'Trip 2', description: 'd', start_date: '2024-01-01', end_date: '2024-01-02', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
    ];
    mockUseQuery.mockReturnValue({ data: trips, isLoading: false });
    renderWithProvider(<TripsList />);
    expect(screen.getByText('Trip 1')).toBeInTheDocument();
    expect(screen.getByText('Trip 2')).toBeInTheDocument();
  });

  it('renders placeholder when empty', () => {
    mockUseQuery.mockReturnValue({ data: [], isLoading: false });
    renderWithProvider(<TripsList />);
    expect(screen.getByText('Пока нет поездок')).toBeInTheDocument();
  });

  it('renders loading spinner', () => {
    mockUseQuery.mockReturnValue({ data: undefined, isLoading: true });
    const { container } = renderWithProvider(<TripsList />);
    expect(container.querySelector('.animate-spin')).not.toBeNull();
  });

  it('Login page submits credentials and redirects', async () => {
    const token = { access_token: 'tok', user: { id: 'u1', email: 'a@b.c', name: 'U', created_at: '2024-01-01T00:00:00Z' } };
    mockLogin.mockResolvedValue(token);
    const navigate = vi.fn();
    mockUseNavigate.mockReturnValue(navigate);

    renderWithProvider(<Login />);
    fireEvent.change(screen.getByPlaceholderText('email'), { target: { value: 'a@b.c' } });
    fireEvent.change(screen.getByPlaceholderText('password'), { target: { value: 'pass123' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ email: 'a@b.c', password: 'pass123' });
    });
    expect(navigate).toHaveBeenCalledWith('/');
  });

  it('Login shows banner on unauthorized', async () => {
    const err = new (ApiError as any)({ code: 'unauthorized', message: 'bad' }) as ApiError;
    mockLogin.mockRejectedValue(err);
    mockUseNavigate.mockReturnValue(vi.fn());

    renderWithProvider(<Login />);
    fireEvent.change(screen.getByPlaceholderText('email'), { target: { value: 'a@b.c' } });
    fireEvent.change(screen.getByPlaceholderText('password'), { target: { value: 'pass123' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });
});
