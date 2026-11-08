import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TripDetail } from './TripDetail';
import { getTrip } from '../api/trips';
import { listStops, updateStop, deleteStop } from '../api/stops';
import { listPhotos } from '../api/photos';
import type { Trip, Stop, Photo } from '../lib/types';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { useQuery, useMutation } from '@tanstack/react-query';

vi.mock('react-router-dom', () => ({
  useParams: () => ({ id: 'trip-1' }),
  useNavigate: vi.fn(() => vi.fn()),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useMutation: vi.fn(),
  useQueryClient: vi.fn(() => ({
    invalidateQueries: vi.fn(),
    refetchQueries: vi.fn(),
  })),
}));

vi.mock('../api/trips', () => ({ getTrip: vi.fn() }));
vi.mock('../api/stops', () => ({
  listStops: vi.fn(),
  createStop: vi.fn(),
  updateStop: vi.fn(),
  deleteStop: vi.fn(),
}));
vi.mock('../api/photos', () => ({ listPhotos: vi.fn() }));
vi.mock('../components/MapView', () => ({
  MapView: vi.fn(({ stops, onMoveStop, selectedStopId, onSelectStop }: any) => (
    <div data-testid="map-view">
      <span data-testid="stops-count">{stops.length}</span>
      <span data-testid="selected">{selectedStopId ?? 'none'}</span>
      <button type="button" onClick={() => onSelectStop('stop-1')}>select</button>
      <button
        type="button"
        onClick={() => onMoveStop('stop-1', { lat: 10, lng: 20 })}
      >
        move
      </button>
    </div>
  )),
}));

const mockGetTrip = vi.mocked(getTrip);
const mockListStops = vi.mocked(listStops);
const mockUpdateStop = vi.mocked(updateStop);
const mockDeleteStop = vi.mocked(deleteStop);
const mockListPhotos = vi.mocked(listPhotos);

const trip: Trip = {
  id: 'trip-1',
  title: 'Поездка в Питер',
  description: '',
  start_date: '2025-06-01',
  end_date: '2025-06-10',
  created_at: '2025-05-01T00:00:00Z',
  updated_at: '2025-05-01T00:00:00Z',
};

const stop: Stop = {
  id: 'stop-1',
  trip_id: 'trip-1',
  title: 'Эрмитаж',
  description: 'Музей',
  location: { lat: 59.9398, lng: 30.3146 },
  visit_date: '2025-06-05',
  order: 0,
  created_at: '2025-05-01T00:00:00Z',
  updated_at: '2025-05-01T00:00:00Z',
};

const photo: Photo = {
  id: 'photo-1',
  trip_id: 'trip-1',
  url: 'https://cdn.example.com/photo-1.jpg',
  created_at: '2025-05-01T00:00:00Z',
};

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
    refetch: vi.fn().mockResolvedValue({ data } as any),
    remove: vi.fn(),
    promise: Promise.resolve(data),
    hasNextPage: false,
    hasPreviousPage: false,
    dataUpdatedAt: Date.now(),
    errorUpdatedAt: 0,
    observersCount: 1,
  } as any;
}

function makeMutation(): UseMutationResult<any, any, any, any> {
  return {
    mutate: vi.fn(),
    mutateAsync: vi.fn().mockResolvedValue({}),
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
  } as any;
}

beforeEach(() => {
  vi.clearAllMocks();
  // Mock window.location for export tests
  Object.defineProperty(window, 'location', {
    value: { href: 'http://localhost:3000/' },
    writable: true,
    configurable: true,
  });
  mockGetTrip.mockResolvedValue(trip);
  mockListStops.mockResolvedValue([stop]);
  mockListPhotos.mockResolvedValue([photo]);
  mockUpdateStop.mockResolvedValue(stop);
  mockDeleteStop.mockResolvedValue(undefined);
  vi.spyOn(window, 'confirm').mockReturnValue(true);
  
  // Mock all queries used in TripDetail
  vi.mocked(useQuery).mockImplementation((opts: any) => {
    if (opts?.queryKey?.includes('trip')) {
      return makeQueryResult(trip);
    }
    if (opts?.queryKey?.includes('stops')) {
      return makeQueryResult([stop]);
    }
    if (opts?.queryKey?.includes('photos')) {
      return makeQueryResult([photo]);
    }
    return makeQueryResult(undefined);
  });

  // Mock all mutations used in TripDetail — mutate calls the actual API mock
  const mutState = makeMutation();
  vi.mocked(mutState.mutate as any).mockImplementation((params: { stopId: string; payload: any }) => {
    mockUpdateStop(params.stopId, params.payload);
  });
  vi.mocked(useMutation).mockReturnValue(mutState);
});

describe('TripDetail', () => {
  it('рендерит MapView со стопами', async () => {
    render(<TripDetail />);
    await waitFor(() => expect(screen.getByTestId('stops-count')).toHaveTextContent('1'));
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Поездка в Питер');
  });

  it('drag маркера вызывает updateStop', async () => {
    render(<TripDetail />);
    await waitFor(() => expect(screen.getByTestId('stops-count')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'move' }));

    await waitFor(() => expect(mockUpdateStop).toHaveBeenCalledTimes(1));
    expect(mockUpdateStop).toHaveBeenCalledWith('stop-1', expect.objectContaining({
      location: { lat: 10, lng: 20 },
    }));
  });

  it('экспорт PDF/CSV устанавливает window.location.href', async () => {
    render(<TripDetail />);
    await waitFor(() => expect(screen.getByTestId('stops-count')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /Экспорт PDF/ }));
    expect(window.location.href).toBe('/api/v1/trips/trip-1/export?format=pdf');

    fireEvent.click(screen.getByRole('button', { name: /Экспорт CSV/ }));
    expect(window.location.href).toBe('/api/v1/trips/trip-1/export?format=csv');
  });

  it('«+ Стоп» открывает StopForm', async () => {
    render(<TripDetail />);
    await waitFor(() => expect(screen.getByTestId('stops-count')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /\+ Стоп/ }));

    await waitFor(() => expect(screen.getByRole('button', { name: /Сохранить/ })).toBeInTheDocument());
  });
});
