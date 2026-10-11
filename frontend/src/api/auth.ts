import api from './client'
import { RegisterPayload, LoginPayload, LoginResponse, User } from '../lib/types'

export const register = (p: RegisterPayload): Promise<User> => {
  return api.post('/auth/register', p).then(res => res.data.user)
}

export const login = (p: LoginPayload): Promise<LoginResponse> => {
  return api.post('/auth/login', p).then(res => res.data)
}

export const refresh = (): Promise<{ access_token: string }> => {
  return api.post('/auth/refresh').then(res => res.data)
}

export const logout = (): Promise<void> => {
  return api.post('/auth/logout').then(() => undefined)
}

export const me = (): Promise<User> => {
  return api.get('/users/me').then(res => res.data)
}