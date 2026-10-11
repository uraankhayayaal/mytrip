import { api } from './client';
import { Photo, PhotoUrlResponse } from '../lib/types';

export async function uploadPhoto(tripId: string, file: File): Promise<Photo> {
  const formData = new FormData();
  formData.append('file', file);

  const response = await api.post(`/trips/${tripId}/photos`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });

  return response.data;
}

export async function deletePhoto(photoId: string): Promise<void> {
  await api.delete(`/photos/${photoId}`);
}

export async function getPhotoUrl(photoId: string): Promise<PhotoUrlResponse> {
  const response = await api.get(`/photos/${photoId}/url`);
  return response.data;
}
