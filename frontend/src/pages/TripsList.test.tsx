import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import type { QueryObserverResult } from '@tanstack/react-query';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { login } from '../api/auth';
import { ApiError } from '../api/errors';
import { listTrips, createTrip } from '../api/trips';
import { TripsList } from './TripsList';
import { Login } from './Login';
import type { Trip } from '../lib/types';

// useQuery — мок с сохранением остальных экспортов реального пакета (QueryClient и т.д.)
vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/react-query')>(
    '@tanstack/react-query',
  );
  return { ...actual, useQuery: vi.fn() };
});

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

vi.mock('../api/trips', () => ({
  listTrips: vi.fn(),
  createTrip: vi.fn(),
  getTrip: vi.fn(),
  updateTrip: vi.fn(),
  deleteTrip: vi.fn(),
}));

// vi.mocked применяется к УЖЕ ИМПОРТИРОВАННОМУ объекту, а не к require/import()
const useQueryMock = vi.mocked(useQuery);
const navMock = vi.mocked(useNavigate);
const listTripsMock = vi.mocked(listTrips);
const createTripMock = vi.mocked(createTrip);

function makeTrip(id: string, title: string): Trip {
  return {
    id,
    title,
    description: 'desc',
    start_date: '2024-01-01',
    end_date: '2024-01-10',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };
}

// Возврат useQuery типизируется под реальный QueryObserverResult<Trip[], unknown>
function tripsQuery(
  overrides: Partial<QueryObserverResult<Trip[], unknown>>,
): QueryObserverResult<Trip[], unknown> {
  return {
    data: undefined,
    dataUpdatedAt: 0,
    error: null,
    errorUpdatedAt: 0,
    fetchStatus: 'idle',
    fetchFailureCount: 0,
    isFetching: false,
    isInitialLoading: false,
    isLoading: false,
    isError: false,
    isLoadingError: false,
    isPaused: false,
    isPlaceholderData: false,
    isPreviousData: false,
    isRefetchError: false,
    isRefetching: false,
    isStale: false,
    isSuccess: false,
    refetch: vi.fn(),
    remove: vi.fn(),
    status: 'idle',
    suspense: false,
    ...overrides,
  };
}

function renderWithProvider(ui: ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

beforeEach(() => {
  vi.clearAllMocks();
  navMock.mockReturnValue(vi.fn());
  listTripsMock.mockResolvedValue([]);
  createTripMock.mockResolvedValue(makeTrip('new', 'Новая поездка'));
  // Default: loading state
  useQueryMock.mockReturnValue(tripsQuery({ isLoading: true }));
});

describe('TripsList', () => {
  it('renders a list of trips', () => {
    const trips = [makeTrip('1', 'Trip 1'), makeTrip('2', 'Trip 2')];
    useQueryMock.mockReturnValue(tripsQuery({ data: trips, isSuccess: true }));

    renderWithProvider(<TripsList />);
    expect(screen.getByText('Trip 1')).toBeInTheDocument();
    expect(screen.getByText('Trip 2')).toBeInTheDocument();
  });

  it('renders placeholder when empty', () => {
    useQueryMock.mockReturnValue(tripsQuery({ data: [], isSuccess: true }));

    renderWithProvider(<TripsList />);
    expect(screen.getByText('Пока нет поездок')).toBeInTheDocument();
  });

  it('renders loading spinner', () => {
    useQueryMock.mockReturnValue(tripsQuery({ isLoading: true }));

    const { container } = renderWithProvider(<TripsList />);
    expect(container.querySelector('.animate-spin')).not.toBeNull();
  });

  it('renders error banner on failure', () => {
    const err = new ApiError('not_found', 'Не найдено');
    useQueryMock.mockReturnValue(
      tripsQuery({ isLoading: false, isError: true, error: err }),
    );

    renderWithProvider(<TripsList />);
    expect(screen.getByRole('alert')).toHaveTextContent('Не найдено');
  });

  it('createTrip: navigates to edit page on success', async () => {
    const newTrip = makeTrip('new-1', 'Новая поездка');
    createTripMock.mockResolvedValue(newTrip);
    const navigate = vi.fn();
    navMock.mockReturnValue(navigate);
    useQueryMock.mockReturnValue(tripsQuery({ data: [], isSuccess: true }));

    renderWithProvider(<TripsList />);
    fireEvent.click(screen.getByRole('button', { name: /Новая поездка/i }));

    await waitFor(() => {
      expect(createTripMock).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith('/trips/new-1/edit');
    });
  });

  it('createTrip: shows alert on failure', async () => {
    createTripMock.mockRejectedValue(new Error('fail'));
    const navigate = vi.fn();
    navMock.mockReturnValue(navigate);
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});
    useQueryMock.mockReturnValue(tripsQuery({ data: [], isSuccess: true }));

    renderWithProvider(<TripsList />);
    fireEvent.click(screen.getByRole('button', { name: /Новая поездка/i }));

    await waitFor(() => {
      expect(alertSpy).toHaveBeenCalled();
    });
  });
});

describe('Login (embedded in TripsList test file for shared mocks)', () => {
  it('submits credentials and calls login', async () => {
    const token = {
      access_token: 'tok',
      user: { id: 'u1', email: 'a@b.c', name: 'U', created_at: '2024-01-01T00:00:00Z' },
    };
    vi.mocked(login).mockResolvedValue(token);
    const navigate = vi.fn();
    navMock.mockReturnValue(navigate);
    useQueryMock.mockReturnValue(tripsQuery({}));

    renderWithProvider(<Login />);
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'a@b.c' } });
    fireEvent.change(screen.getByLabelText('Пароль'), { target: { value: 'pass1234' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(vi.mocked(login)).toHaveBeenCalledWith({ email: 'a@b.c', password: 'pass1234' });
    });
  });

  it('shows banner on unauthorized', async () => {
    const err = new ApiError('unauthorized', 'bad');
    vi.mocked(login).mockRejectedValue(err);
    navMock.mockReturnValue(vi.fn());
    useQueryMock.mockReturnValue(tripsQuery({}));

    renderWithProvider(<Login />);
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'a@b.c' } });
    fireEvent.change(screen.getByLabelText('Пароль'), { target: { value: 'pass1234' } });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });
});
