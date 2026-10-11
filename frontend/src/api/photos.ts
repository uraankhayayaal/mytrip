import { api } from './client';
import type { Photo } from '../lib/types';

// GET /trips/{tripId}/photos
export async function listPhotos(tripId: string): Promise<Photo[]> {
  const r = await api.get<Photo[]>(`/trips/${tripId}/photos`);
  return r.data;
}

// POST /trips/{tripId}/photos (multipart 'file') → 201 Photo
export async function uploadPhoto(tripId: string, file: File): Promise<Photo> {
  const form = new FormData();
  form.append('file', file);
  const r = await api.post<Photo>(`/trips/${tripId}/photos`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return r.data;
}

// DELETE /photos/{photoId} → 204
export async function deletePhoto(photoId: string): Promise<void> {
  await api.delete(`/photos/${photoId}`);
}
