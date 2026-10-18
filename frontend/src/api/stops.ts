import { api } from './client';
import type { Stop, StopPayload } from '../lib/types';

// GET /trips/{tripId}/stops
export async function listStops(tripId: string): Promise<Stop[]> {
  const r = await api.get<Stop[]>(`/trips/${tripId}/stops`);
  return r.data;
}

// POST /trips/{tripId}/stops -> 201
export async function createStop(tripId: string, p: StopPayload): Promise<Stop> {
  const r = await api.post<Stop>(`/trips/${tripId}/stops`, p);
  return r.data;
}

// PUT /trips/{tripId}/stops/{stopId} -> 200
export async function updateStop(tripId: string, stopId: string, p: StopPayload): Promise<Stop> {
  const r = await api.put<Stop>(`/trips/${tripId}/stops/${stopId}`, p);
  return r.data;
}

// DELETE /trips/{tripId}/stops/{stopId} -> 204
export async function deleteStop(tripId: string, stopId: string): Promise<void> {
  await api.delete(`/trips/${tripId}/stops/${stopId}`);
}
