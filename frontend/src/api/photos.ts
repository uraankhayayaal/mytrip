import { api } from './client';
import type { Photo, PhotoUrlResponse } from '../lib/types';

// POST /trips/{tripId}/photos, multipart FormData поле 'file' -> 201
export async function uploadPhoto(tripId: string, file: File): Promise<Photo> {
  const fd = new FormData();
  fd.append('file', file);
  const r = await api.post<Photo>(`/trips/${tripId}/photos`, fd);
  return r.data;
}

// DELETE /photos/{photoId} -> 204
export async function deletePhoto(photoId: string): Promise<void> {
  await api.delete(`/photos/${photoId}`);
}

// GET /photos/{photoId}/url
export async function getPhotoUrl(photoId: string): Promise<PhotoUrlResponse> {
  const r = await api.get<PhotoUrlResponse>(`/photos/${photoId}/url`);
  return r.data;
}
