import { api } from './client';
import { Trip, TripPayload } from '../lib/types';

export async function listTrips(): Promise<Trip[]> {
  const response = await api.get('/trips');
  return response.data;
}

export async function createTrip(p: TripPayload): Promise<Trip> {
  const response = await api.post('/trips', p);
  return response.data;
}

export async function getTrip(id: string): Promise<Trip> {
  const response = await api.get(`/trips/${id}`);
  return response.data;
}

export async function updateTrip(id: string, p: TripPayload): Promise<Trip> {
  const response = await api.put(`/trips/${id}`, p);
  return response.data;
}

export async function deleteTrip(id: string): Promise<void> {
  await api.delete(`/trips/${id}`);
}
