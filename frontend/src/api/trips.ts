import { api } from './client';
import type { Trip, TripPayload } from '../lib/types';

// GET /trips
export async function listTrips(): Promise<Trip[]> {
  const r = await api.get<Trip[]>('/trips');
  return r.data;
}

// POST /trips -> 201
export async function createTrip(p: TripPayload): Promise<Trip> {
  const r = await api.post<Trip>('/trips', p);
  return r.data;
}

// GET /trips/{id}
export async function getTrip(id: string): Promise<Trip> {
  const r = await api.get<Trip>(`/trips/${id}`);
  return r.data;
}

// PUT /trips/{id}
export async function updateTrip(id: string, p: TripPayload): Promise<Trip> {
  const r = await api.put<Trip>(`/trips/${id}`, p);
  return r.data;
}

// DELETE /trips/{id} -> 204
export async function deleteTrip(id: string): Promise<void> {
  await api.delete(`/trips/${id}`);
}
