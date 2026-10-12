import { api } from './client';
import { Stop, StopPayload } from '../lib/types';

export async function listStops(tripId: string): Promise<Stop[]> {
  const response = await api.get(`/trips/${tripId}/stops`);
  return response.data;
}

export async function createStop(tripId: string, p: StopPayload): Promise<Stop> {
  const response = await api.post(`/trips/${tripId}/stops`, p);
  return response.data;
}

export async function updateStop(stopId: string, p: StopPayload): Promise<Stop> {
  const response = await api.put(`/stops/${stopId}`, p);
  return response.data;
}

export async function deleteStop(stopId: string): Promise<void> {
  await api.delete(`/stops/${stopId}`);
}
