import { describe, it, expect } from 'vitest';
import { tripsToEvents, stopsToEvents } from './calendar';
import { Trip, Stop } from '../lib/types';

describe('calendar helpers', () => {
  describe('tripsToEvents', () => {
    it('should convert trips to events', () => {
      const trips: Trip[] = [
        {
          id: '1',
          title: 'Trip 1',
          description: 'Description 1',
          start_date: '2023-01-01',
          end_date: '2023-01-05',
          created_at: '2023-01-01T00:00:00Z',
          updated_at: '2023-01-01T00:00:00Z',
        },
        {
          id: '2',
          title: 'Trip 2',
          description: 'Description 2',
          start_date: '2023-02-01',
          end_date: '2023-02-05',
          created_at: '2023-02-01T00:00:00Z',
          updated_at: '2023-02-01T00:00:00Z',
        },
      ];

      const result = tripsToEvents(trips);

      expect(result).toEqual([
        {
          id: 'trip-1',
          title: 'Trip 1',
          start: new Date('2023-01-01'),
          end: new Date('2023-01-05'),
          tripId: '1',
          kind: 'trip',
        },
        {
          id: 'trip-2',
          title: 'Trip 2',
          start: new Date('2023-02-01'),
          end: new Date('2023-02-05'),
          tripId: '2',
          kind: 'trip',
        },
      ]);
    });

    it('should return empty array for empty input', () => {
      expect(tripsToEvents([])).toEqual([]);
    });
  });

  describe('stopsToEvents', () => {
    it('should convert stops to events', () => {
      const stops: Stop[] = [
        {
          id: '1',
          trip_id: 't1',
          title: 'Stop 1',
          description: 'Description 1',
          location: { lat: 1, lng: 1 },
          visit_date: '2023-01-02',
          order: 1,
          created_at: '2023-01-01T00:00:00Z',
          updated_at: '2023-01-01T00:00:00Z',
        },
        {
          id: '2',
          trip_id: 't2',
          title: 'Stop 2',
          description: 'Description 2',
          location: { lat: 2, lng: 2 },
          visit_date: '2023-01-03',
          order: 2,
          created_at: '2023-01-01T00:00:00Z',
          updated_at: '2023-01-01T00:00:00Z',
        },
      ];

      const result = stopsToEvents(stops);

      expect(result).toEqual([
        {
          id: 'stop-1',
          title: 'Stop 1',
          start: new Date('2023-01-02'),
          end: new Date('2023-01-02'),
          tripId: 't1',
          kind: 'stop',
        },
        {
          id: 'stop-2',
          title: 'Stop 2',
          start: new Date('2023-01-03'),
          end: new Date('2023-01-03'),
          tripId: 't2',
          kind: 'stop',
        },
      ]);
    });

    it('should return empty array for empty input', () => {
      expect(stopsToEvents([])).toEqual([]);
    });
  });
});