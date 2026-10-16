import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import type { QueryObserverResult } from '@tanstack/react-query';
import { TripsList } from '../pages/TripsList';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { createTrip, listTrips } from '../api/trips';
import type { Trip } from '../lib/types';

type QueryResult<T> = Partial<QueryObserverResult<T, unknown>>;

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useQueryClient: vi.fn(() => ({
    invalidateQueries: vi.fn(),
  })),
}));

vi.mock('react-router-dom', () => ({
  useNavigate: vi.fn(),
}));

vi.mock('../api/trips', () => ({
  createTrip: vi.fn(),
  listTrips: vi.fn(),
}));

const mockUseQuery = vi.mocked(useQuery);
const mockUseNavigate = vi.mocked(useNavigate);
const mockCreateTrip = vi.mocked(createTrip);

function makeTrip(id: string, title: string, description: string): Trip {
  return {
    id,
    title,
    description,
    start_date: '2024-01-01',
    end_date: '2024-01-02',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn());
  mockUseQuery.mockReturnValue({
    data: undefined,
    isLoading: false,
    isError: false,
    error: null,
  } satisfies QueryResult<Trip[]>);
});

describe('TripsList', () => {
  it('should render a list of trips', async () => {
    const mockTrips: Trip[] = [makeTrip('1', 'Trip 1', 'Desc 1'), makeTrip('2', 'Trip 2', 'Desc 2')];
    mockUseQuery.mockReturnValue({
      data: mockTrips,
      isLoading: false,
      isError: false,
      error: null,
    } satisfies QueryResult<Trip[]>);

    render(<TripsList />);
    expect(screen.getByText('Trip 1')).toBeInTheDocument();
    expect(screen.getByText('Trip 2')).toBeInTheDocument();
  });

  it('should render placeholder when list is empty', async () => {
    mockUseQuery.mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      error: null,
    } satisfies QueryResult<Trip[]>);

    render(<TripsList />);
    expect(screen.getByText('Пока нет поездок')).toBeInTheDocument();
  });

  it('should create a draft trip and navigate to edit when clicking + New Trip', async () => {
    const navigate = mockUseNavigate() as ReturnType<typeof useNavigate>;
    const newTrip: Trip = makeTrip('new-id', 'Новая поездка', '');
    mockCreateTrip.mockResolvedValue(newTrip);

    render(<TripsList />);
    fireEvent.click(screen.getByText('+ Новая поездка'));

    await waitFor(() => {
      expect(mockCreateTrip).toHaveBeenCalledWith(
        expect.objectContaining({ title: 'Новая поездка' }),
      );
      expect(navigate).toHaveBeenCalledWith('/trips/new-id/edit');
    });
  });
});
