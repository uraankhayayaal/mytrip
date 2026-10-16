import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import type { QueryObserverResult, UseMutationResult } from '@tanstack/react-query';
import { TripDetail } from './TripDetail';
import { getTrip } from '../api/trips';
import { listStops, createStop, updateStop, deleteStop } from '../api/stops';
import { listPhotos } from '../api/photos';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { MapView } from '../components/MapView';
import type { Location, Photo, Stop, Trip } from '../lib/types';

vi.mock('../api/trips', () => ({ getTrip: vi.fn() }));
vi.mock('../api/stops', () => ({
  listStops: vi.fn(),
  createStop: vi.fn(),
  updateStop: vi.fn(),
  deleteStop: vi.fn(),
}));
vi.mock('../api/photos', () => ({ listPhotos: vi.fn() }));
vi.mock('react-router-dom', () => ({
  useParams: vi.fn(),
  useNavigate: vi.fn(),
}));
vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useMutation: vi.fn(),
  useQueryClient: vi.fn(() => ({
    invalidateQueries: vi.fn(),
  })),
}));
vi.mock('../components/MapView', () => ({
  MapView: vi.fn(
    ({
      stops,
      onMoveStop,
      selectedStopId,
      onSelectStop,
    }: {
      stops: Stop[];
      onMoveStop?: (id: string, loc: Location) => void;
      selectedStopId?: string | null;
      onSelectStop?: (id: string) => void;
    }) => (
      <div data-testid="map-view">
        <span data-testid="stops-count">{stops.length}</span>
        <span data-testid="selected">{selectedStopId ?? 'none'}</span>
        <button type="button" onClick={() => onSelectStop?.('stop-1')}>
          select
        </button>
        <button type="button" onClick={() => onMoveStop?.('stop-1', { lat: 10, lng: 20 })}>
          move
        </button>
      </div>
    ),
  ),
}));

const mockUseQuery = vi.mocked(useQuery) as Mock;
const mockUseMutation = vi.mocked(useMutation) as Mock;
const mockUseParams = vi.mocked(useParams) as Mock;
const mockUseNavigate = vi.mocked(useNavigate) as Mock;

const trip: Trip = {
  id: 'trip-1',
  title: 'Поездка в Питер',
  description: 'Поездка по Петербургу',
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

function mockQueries() {
  mockUseQuery.mockImplementation((options: { queryKey: readonly unknown[] }) => {
    const key = options.queryKey[0] as string;
    if (key === 'trip') {
      return {
        data: trip,
        isLoading: false,
        isError: false,
        error: null,
      } satisfies Partial<QueryObserverResult<Trip, unknown>>;
    }
    if (key === 'stops') {
      return {
        data: [stop],
        isLoading: false,
        isError: false,
        error: null,
      } satisfies Partial<QueryObserverResult<Stop[], unknown>>;
    }
    return {
      data: [photo],
      isLoading: false,
      isError: false,
      error: null,
    } satisfies Partial<QueryObserverResult<Photo[], unknown>>;
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseParams.mockReturnValue({ id: 'trip-1' });
  mockUseNavigate.mockReturnValue(vi.fn());
  mockQueries();
  mockUseMutation.mockReturnValue({
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    reset: vi.fn(),
  } as unknown as UseMutationResult);
  vi.spyOn(window, 'confirm').mockReturnValue(true);
  Object.defineProperty(window, 'location', {
    value: { href: window.location.href },
    writable: true,
    configurable: true,
  });
});

describe('TripDetail', () => {
  it('рендерит MapView со стопами', async () => {
    render(<TripDetail />);
    await waitFor(() => expect(screen.getByTestId('stops-count')).toHaveTextContent('1'));
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Поездка в Питер');
  });

  it('drag маркера вызывает updateStop', async () => {
    const mutate = vi.fn();
    mockUseMutation.mockReturnValue({
      mutate,
      mutateAsync: vi.fn(),
      reset: vi.fn(),
    } as unknown as UseMutationResult);
    render(<TripDetail />);
    await waitFor(() => expect(screen.getByTestId('stops-count')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'move' }));

    await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));
    expect(mutate).toHaveBeenCalledWith({
      stopId: 'stop-1',
      payload: expect.objectContaining({ location: { lat: 10, lng: 20 } }),
    });
  });

  it('экспорт PDF/CSV устанавливает window.location.href', () => {
    render(<TripDetail />);
    fireEvent.click(screen.getByRole('button', { name: /Экспорт PDF/ }));
    expect(window.location.href).toBe('/api/v1/trips/trip-1/export?format=pdf');

    fireEvent.click(screen.getByRole('button', { name: /Экспорт CSV/ }));
    expect(window.location.href).toBe('/api/v1/trips/trip-1/export?format=csv');
  });

  it('«+ Стоп» открывает StopForm', async () => {
    render(<TripDetail />);
    await waitFor(() => expect(screen.getByTestId('stops-count')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /\+ Стоп/ }));

    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Сохранить/ })).toBeInTheDocument(),
    );
  });
});
