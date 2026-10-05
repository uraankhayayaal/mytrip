import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { Stop, Location } from '../lib/types';

// Мок react-leaflet ОБЯЗАТЕЛЕН: jsdom не поддерживает canvas/SVG-карту.
vi.mock('react-leaflet', () => ({
  MapContainer: ({ children }: { children?: React.ReactNode }) => <div data-testid="map-container">{children}</div>,
  TileLayer: () => <div data-testid="tile-layer" />,
  Marker: ({
    position,
    eventHandlers,
  }: {
    position: [number, number];
    eventHandlers?: { click?: () => void; dragend?: (e: { target: { getLatLng: () => { lat: number; lng: number } } }) => void };
  }) => (
    <div
      data-testid="marker"
      data-lat={position[0]}
      data-lng={position[1]}
      onClick={() => eventHandlers?.click?.()}
      onDragEnd={() =>
        eventHandlers?.dragend?.({ target: { getLatLng: () => ({ lat: position[0] + 0.1, lng: position[1] + 0.1 }) } })
      }
    />
  ),
  Polyline: ({ positions }: { positions: [number, number][] }) => (
    <div data-testid="polyline" data-positions={JSON.stringify(positions)} />
  ),
}));

import { MapView } from './MapView';

function makeStop(id: string, order: number, lat: number, lng: number): Stop {
  return {
    id,
    trip_id: 'trip-1',
    title: `Stop ${id}`,
    description: '',
    location: { lat, lng },
    visit_date: '2024-01-01',
    order,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };
}

const stops: Stop[] = [
  makeStop('s1', 0, 55.75, 37.6),
  makeStop('s2', 1, 55.8, 37.7),
  makeStop('s3', 2, 55.9, 37.8),
];

describe('MapView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders N markers for N stops', () => {
    render(<MapView stops={stops} />);
    expect(screen.getAllByTestId('marker')).toHaveLength(3);
    expect(screen.getByTestId('map-container')).toBeInTheDocument();
    expect(screen.getByTestId('tile-layer')).toBeInTheDocument();
  });

  it('renders empty map without markers/polyline when stops is empty', () => {
    render(<MapView stops={[]} />);
    expect(screen.queryAllByTestId('marker')).toHaveLength(0);
    expect(screen.queryByTestId('polyline')).toBeNull();
    expect(screen.getByTestId('map-container')).toBeInTheDocument();
  });

  it('calls onMoveStop with new location on dragend', () => {
    const onMoveStop = vi.fn();
    render(<MapView stops={stops} onMoveStop={onMoveStop} />);
    const markers = screen.getAllByTestId('marker');
    fireEvent.dragEnd(markers[0]);
    expect(onMoveStop).toHaveBeenCalledTimes(1);
    const [stopId, location] = onMoveStop.mock.calls[0] as [string, Location];
    expect(stopId).toBe('s1');
    expect(location).toEqual({ lat: 55.75 + 0.1, lng: 37.6 + 0.1 });
  });

  it('calls onSelectStop on marker click', () => {
    const onSelectStop = vi.fn();
    render(<MapView stops={stops} onSelectStop={onSelectStop} />);
    const markers = screen.getAllByTestId('marker');
    fireEvent.click(markers[1]);
    expect(onSelectStop).toHaveBeenCalledTimes(1);
    expect(onSelectStop).toHaveBeenCalledWith('s2');
  });

  it('passes polyline positions in order-sorted sequence', () => {
    const unsorted: Stop[] = [stops[2], stops[0], stops[1]];
    render(<MapView stops={unsorted} />);
    const polyline = screen.getByTestId('polyline');
    const positions = JSON.parse(polyline.getAttribute('data-positions') ?? '[]') as [number, number][];
    expect(positions).toEqual([
      [55.75, 37.6],
      [55.8, 37.7],
      [55.9, 37.8],
    ]);
  });
});
