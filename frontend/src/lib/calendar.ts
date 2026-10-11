import { Trip, Stop } from './types';

export interface CalendarEvent {
  id: string;
  title: string;
  start: Date;
  end: Date;
  tripId: string;
  kind: 'trip' | 'stop';
}

export function tripsToEvents(trips: Trip[]): CalendarEvent[] {
  return trips.map(trip => ({
    id: `trip-${trip.id}`,
    title: trip.title,
    start: new Date(trip.start_date),
    end: new Date(trip.end_date),
    tripId: trip.id,
    kind: 'trip',
  }));
}

export function stopsToEvents(stops: Stop[]): CalendarEvent[] {
  return stops.map(stop => ({
    id: `stop-${stop.id}`,
    title: stop.title,
    start: new Date(stop.visit_date),
    end: new Date(stop.visit_date),
    tripId: stop.trip_id,
    kind: 'stop',
  }));
}
