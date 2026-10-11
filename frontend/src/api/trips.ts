import api from './client'
import { TripPayload, Trip } from '../lib/types'

export const listTrips = (): Promise<Trip[]> => {
  return api.get('/trips').then(res => res.data)
}

export const createTrip = (p: TripPayload): Promise<Trip> => {
  return api.post('/trips', p).then(res => res.data)
}

export const getTrip = (id: string): Promise<Trip> => {
  return api.get(`/trips/${id}`).then(res => res.data)
}

export const updateTrip = (id: string, p: TripPayload): Promise<Trip> => {
  return api.put(`/trips/${id}`, p).then(res => res.data)
}

export const deleteTrip = (id: string): Promise<void> => {
  return api.delete(`/trips/${id}`).then(() => undefined)
}