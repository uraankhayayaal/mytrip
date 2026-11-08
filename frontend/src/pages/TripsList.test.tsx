import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, type Mock, beforeEach } from 'vitest';
import { TripsList } from '../pages/TripsList';
import type { UseQueryResult } from '@tanstack/react-query';
import { useQuery, useMutation } from '@tanstack/react-query';
import { createTrip, listTrips } from '../api/trips';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useMutation: vi.fn(),
  useQueryClient: vi.fn(() => ({
    invalidateQueries: vi.fn(),
    refetchQueries: vi.fn(),
  })),
}));

vi.mock('../api/trips', () => ({
  createTrip: vi.fn(),
  listTrips: vi.fn(),
}));

function makeQueryResult(data: unknown): UseQueryResult<unknown, unknown> {
  return {
    data,
    isLoading: false,
    isFetching: false,
    isFetched: true,
    isPending: false,
    isSuccess: data !== undefined && data !== null,
    isError: false,
    isLoadingError: false,
    isRefetchError: false,
    isFetchNextPageError: false,
    isFetchPreviousPageError: false,
    error: null,
    fetchStatus: 'idle' as const,
    fetchNextPage: vi.fn(),
    fetchPreviousPage: vi.fn(),
    refetch: vi.fn().mockResolvedValue({ data, isLoading: false } as any),
    remove: vi.fn(),
    promise: Promise.resolve(data),
    hasNextPage: false,
    hasPreviousPage: false,
    dataUpdatedAt: Date.now(),
    errorUpdatedAt: 0,
    observersCount: 1,
  } as any;
}

function makeMutation(mutateAsyncFn?: Mock): any {
  const ma = mutateAsyncFn || vi.fn().mockResolvedValue({});
  return {
    mutate: vi.fn(),
    mutateAsync: ma,
    data: undefined,
    error: null,
    variables: undefined,
    isSuccess: false,
    isError: false,
    isIdle: true,
    isPending: false,
    isSettled: false,
    status: 'idle' as const,
    reset: vi.fn(),
    promise: undefined,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('TripsList', () => {
  it('should render a list of trips', async () => {
    const mockTrips = [
      { id: '1', title: 'Trip 1', description: 'Desc 1', start_date: '2024-01-01', end_date: '2024-01-02', created_at: '', updated_at: '' },
      { id: '2', title: 'Trip 2', description: 'Desc 2', start_date: '2024-01-03', end_date: '2024-01-04', created_at: '', updated_at: '' },
    ];
    vi.mocked(useQuery).mockReturnValue(makeQueryResult(mockTrips));
    vi.mocked(useMutation).mockReturnValue(makeMutation());
    vi.mocked(listTrips).mockResolvedValue(mockTrips);

    render(<TripsList />);
    expect(screen.getByText('Trip 1')).toBeInTheDocument();
    expect(screen.getByText('Trip 2')).toBeInTheDocument();
  });

  it('should render placeholder when list is empty', async () => {
    vi.mocked(useQuery).mockReturnValue(makeQueryResult([]));
    vi.mocked(useMutation).mockReturnValue(makeMutation());
    vi.mocked(listTrips).mockResolvedValue([]);

    render(<TripsList />);
    expect(screen.getByText('Пока нет поездок')).toBeInTheDocument();
  });

  it('should create a draft trip and navigate to edit when clicking + New Trip', async () => {
    vi.mocked(useQuery).mockReturnValue(makeQueryResult([]));
    vi.mocked(createTrip).mockResolvedValue({ id: 'new-id', title: 'Новая поездка' } as any);
    const mockMutateAsync: Mock = vi.fn().mockResolvedValue({ id: 'new-id' });
    vi.mocked(useMutation).mockReturnValue({ ...makeMutation(), mutateAsync: mockMutateAsync });
    vi.mocked(listTrips).mockResolvedValue([]);

    render(<TripsList />);
    fireEvent.click(screen.getByText('+ Новая поездка'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/trips/new-id/edit');
    });
  });
});
