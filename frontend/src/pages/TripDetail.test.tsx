import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TripDetail } from './TripDetail';
import { getTrip } from '../api/trips';
import { listStops, createStop, updateStop, deleteStop } from '../api/stops';
import { listPhotos } from '../api/photos';
import { MapView } from '../components/MapView';
import type { Trip, Stop, Photo } from '../lib/types';

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
  start_date: '2025-06-01',
  end_date: '2025-06-10',
  cover_photo_url: null,
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

beforeEach(() => {
  vi.clearAllMocks();
  mockGetTrip.mockResolvedValue(trip);
  mockListStops.mockResolvedValue([stop]);
  mockListPhotos.mockResolvedValue([photo]);
  mockUpdateStop.mockResolvedValue(stop);
  mockDeleteStop.mockResolvedValue(undefined);
  vi.spyOn(window, 'confirm').mockReturnValue(true);
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
