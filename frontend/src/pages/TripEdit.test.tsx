import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, type Mock, beforeEach } from 'vitest';
import { TripEdit } from '../pages/TripEdit';
import type { UseQueryResult } from '@tanstack/react-query';
import { useQuery, useMutation } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useParams: vi.fn(() => ({ id: '1' })),
  useNavigate: vi.fn(() => mockNavigate),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useMutation: vi.fn(),
  useQueryClient: vi.fn(() => ({
    invalidateQueries: vi.fn(),
  })),
}));

vi.mock('../api/trips', () => ({
  getTrip: vi.fn(),
  updateTrip: vi.fn(),
}));

// Mock TripForm to avoid react-hook-form / jsdom interaction issues
vi.mock('../components/TripForm', () => ({
  TripForm: ({ onSubmit }: { onSubmit: (p: any) => Promise<void> }) => (
    <form data-testid="trip-form">
      <label htmlFor="title">Название поездки</label>
      <input id="title" defaultValue="Trip 1" />
      <label htmlFor="description">Описание</label>
      <textarea id="description" defaultValue="Desc 1" />
      <button type="submit" onClick={async () => {
        await onSubmit({
          title: 'Trip 1',
          description: 'Desc 1',
          start_date: '2024-01-01',
          end_date: '2024-01-02',
        });
      }}>
        Сохранить
      </button>
    </form>
  ),
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

function makeMutation(mutateAsyncFn?: Mock, onSuccess?: () => void): any {
  const ma = mutateAsyncFn || vi.fn().mockResolvedValue({});
  return {
    mutate: vi.fn(),
    mutateAsync: async (...args: any[]) => {
      const result = await ma(...args);
      if (onSuccess) onSuccess();
      return result;
    },
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

describe('TripEdit', () => {
  const mockTrip = {
    id: '1',
    title: 'Trip 1',
    description: 'Desc 1',
    start_date: '2024-01-01',
    end_date: '2024-01-02',
    created_at: '',
    updated_at: '',
  };

  it('should prefill form with trip data', async () => {
    vi.mocked(useQuery).mockReturnValue(makeQueryResult(mockTrip));
    vi.mocked(useMutation).mockReturnValue(makeMutation());

    render(<TripEdit />);
    expect(screen.getByLabelText(/Название поездки/i)).toHaveValue('Trip 1');
    expect(screen.getByLabelText(/Описание/i)).toHaveValue('Desc 1');
  });

  it('should update trip and navigate on submit', async () => {
    vi.mocked(useParams).mockReturnValue({ id: '1' });
    vi.mocked(useQuery).mockReturnValue(makeQueryResult(mockTrip));
    const mutateAsync = vi.fn().mockResolvedValue({});
    
    // Capture the onSuccess callback from useMutation options
    vi.mocked(useMutation).mockImplementation((opts: any) => {
      return makeMutation(mutateAsync, opts?.onSuccess);
    });

    render(<TripEdit />);
    
    // Click the submit button in the mocked TripForm
    const submitBtn = screen.getByText('Сохранить');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalled();
    });
    // TripEdit's mutation.onSuccess calls navigate — it should be called after mutateAsync resolves
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/trips/1');
    });
  });
});
