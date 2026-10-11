import api from './client'
import { StopPayload, Stop } from '../lib/types'

export const listStops = (tripId: string): Promise<Stop[]> => {
  return api.get(`/trips/${tripId}/stops`).then(res => res.data)
}

export const createStop = (tripId: string, p: StopPayload): Promise<Stop> => {
  return api.post(`/trips/${tripId}/stops`, p).then(res => res.data)
}

export const updateStop = (stopId: string, p: StopPayload): Promise<Stop> => {
  return api.put(`/stops/${stopId}`, p).then(res => res.data)
}

export const deleteStop = (stopId: string): Promise<void> => {
  return api.delete(`/stops/${stopId}`).then(() => undefined)
}