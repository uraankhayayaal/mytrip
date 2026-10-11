import api from './client'
import { Photo, PhotoUrlResponse } from '../lib/types'

export const uploadPhoto = (tripId: string, file: File): Promise<Photo> => {
  const formData = new FormData()
  formData.append('file', file)
  return api.post(`/trips/${tripId}/photos`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  }).then(res => res.data)
}

export const deletePhoto = (photoId: string): Promise<void> => {
  return api.delete(`/photos/${photoId}`).then(() => undefined)
}

export const getPhotoUrl = (photoId: string): Promise<PhotoUrlResponse> => {
  return api.get(`/photos/${photoId}/url`).then(res => res.data)
}