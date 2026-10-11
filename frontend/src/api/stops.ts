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

// PUT /stops/{stopId}
export async function updateStop(stopId: string, p: StopPayload): Promise<Stop> {
  const r = await api.put<Stop>(`/stops/${stopId}`, p);
  return r.data;
}

// DELETE /stops/{stopId} -> 204
export async function deleteStop(stopId: string): Promise<void> {
  await api.delete(`/stops/${stopId}`);
}
