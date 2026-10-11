import { api } from './client';
import { User, RegisterPayload, LoginPayload, LoginResponse } from '../lib/types';

export async function register(p: RegisterPayload): Promise<User> {
  const response = await api.post('/auth/register', p);
  return response.data.user;
}

export async function login(p: LoginPayload): Promise<LoginResponse> {
  const response = await api.post('/auth/login', p);
  return response.data;
}

export async function refresh(): Promise<{ access_token: string }> {
  const response = await api.post('/auth/refresh');
  return response.data;
}

export async function logout(): Promise<void> {
  await api.post('/auth/logout');
}

export async function me(): Promise<User> {
  const response = await api.get('/users/me');
  return response.data;
}
