import { api } from './client';
import type { LoginPayload, LoginResponse, RegisterPayload, User } from '../lib/types';

// POST /auth/register -> 201 {user}
export async function register(p: RegisterPayload): Promise<User> {
  const r = await api.post<{ user: User }>('/auth/register', p);
  return r.data.user;
}

// POST /auth/login -> 200
export async function login(p: LoginPayload): Promise<LoginResponse> {
  const r = await api.post<LoginResponse>('/auth/login', p);
  return r.data;
}

// POST /auth/refresh -> 200 (httpOnly cookie 'rt', withCredentials)
export async function refresh(): Promise<{ access_token: string }> {
  const r = await api.post<{ access_token: string }>('/auth/refresh');
  return r.data;
}

// POST /auth/logout -> 204
export async function logout(): Promise<void> {
  await api.post('/auth/logout');
}

// GET /users/me
export async function me(): Promise<User> {
  const r = await api.get<User>('/users/me');
  return r.data;
}
