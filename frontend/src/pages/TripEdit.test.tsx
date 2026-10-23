import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import type { QueryObserverResult, UseMutationResult } from '@tanstack/react-query';
import { QueryClient, QueryClientProvider, useQuery, useMutation } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { getTrip, updateTrip } from '../api/trips';
import type { Trip } from '../lib/types';
import { TripEdit } from './TripEdit';

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/react-query')>(
    '@tanstack/react-query',
  );
  return {
    ...actual,
    useQuery: vi.fn(),
    useMutation: vi.fn(),
  };
});

vi.mock('react-router-dom', () => ({
  useNavigate: vi.fn(),
  useParams: vi.fn(() => ({ id: 'trip-1' })),
}));

vi.mock('../api/trips', () => ({
  getTrip: vi.fn(),
  updateTrip: vi.fn(),
  listTrips: vi.fn(),
  createTrip: vi.fn(),
  deleteTrip: vi.fn(),
}));

const mockUseQuery = vi.mocked(useQuery);
const mockUseMutation = vi.mocked(useMutation);
const mockUseParams = vi.mocked(useParams);
const mockUseNavigate = vi.mocked(useNavigate);
const mockGetTrip = vi.mocked(getTrip);
const mockUpdateTrip = vi.mocked(updateTrip);

const fullTrip: Trip = {
  id: 'trip-1',
  title: 'Original Title',
  description: 'Original desc',
  start_date: '2024-05-01',
  end_date: '2024-05-05',
  created_at: '2024-05-01T00:00:00Z',
  updated_at: '2024-05-01T00:00:00Z',
};

// Возврат useQuery типизируется под реальный QueryObserverResult<Trip, unknown>
function tripQuery(
  overrides: Partial<QueryObserverResult<Trip, unknown>>,
): QueryObserverResult<Trip, unknown> {
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
  mockUseParams.mockReturnValue({ id: 'trip-1' });
  mockUseNavigate.mockReturnValue(vi.fn());
  mockGetTrip.mockResolvedValue(fullTrip);
  mockUseQuery.mockReturnValue(tripQuery({ data: fullTrip, isSuccess: true }));
  mockUseMutation.mockReturnValue({
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    reset: vi.fn(),
    status: 'idle',
    data: undefined,
    error: null,
    variables: undefined,
  } as unknown as UseMutationResult);
});

describe('TripEdit', () => {
  it('should prefill form with trip data', () => {
    mockUseQuery.mockReturnValue(tripQuery({ data: fullTrip, isSuccess: true }));
    renderWithProvider(<TripEdit />);

    expect(screen.getByDisplayValue('Original Title')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Original desc')).toBeInTheDocument();
  });

  it('should update trip and navigate on submit', async () => {
    const updated: Trip = { ...fullTrip, title: 'Updated Title', updated_at: '2024-05-02T00:00:00Z' };
    const mutateAsync = vi.fn().mockResolvedValue(updated);
    mockUpdateTrip.mockResolvedValue(updated);
    mockUseMutation.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync,
      reset: vi.fn(),
      status: 'idle',
      data: undefined,
      error: null,
      variables: undefined,
    } as unknown as UseMutationResult);
    const navigate = vi.fn();
    mockUseNavigate.mockReturnValue(navigate);
    mockUseQuery.mockReturnValue(tripQuery({ data: fullTrip, isSuccess: true }));

    renderWithProvider(<TripEdit />);

    const titleInput = screen.getByDisplayValue('Original Title') as HTMLInputElement;
    fireEvent.change(titleInput, { target: { value: 'Updated Title' } });
    fireEvent.click(screen.getByRole('button', { name: /сохранить/i }));

    await waitFor(() => {
      expect(mockUpdateTrip).toHaveBeenCalledWith('trip-1', {
        title: 'Updated Title',
        description: 'Original desc',
        start_date: '2024-05-01',
        end_date: '2024-05-05',
      });
    });
    expect(navigate).toHaveBeenCalledWith('/trips/trip-1');
  });
});
