import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { toApiError } from './errors';

// Модуль-состояние: JWT access token. Refresh идёт через httpOnly cookie 'rt' (withCredentials).
let accessToken: string | null = null;

export function setAccessToken(t: string | null): void {
  accessToken = t;
}

export function getAccessToken(): string | null {
  return accessToken;
}

// Единый axios-инстанс: ТОЛЬКО относительные пути (dev — Vite proxy /api -> localhost:8080,
// prod — nginx проксирует /api -> server:8080).
export const api: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
});

// Request-интерцептор: Authorization: Bearer <token>, если accessToken установлен.
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  if (accessToken) {
    config.headers.set('Authorization', `Bearer ${accessToken}`);
  }
  return config;
});

// Response-интерцептор: при 401 — один раз POST /auth/refresh (без Authorization, cookie rt
// httpOnly). Успех: обновляем token и повторяем запрос; неудача: сброс + редирект на /login.
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as
      (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined;
    const status = error.response?.status;
    const isRefreshCall = !!originalRequest?.url && originalRequest.url.includes('/auth/refresh');

    if (status === 401 && originalRequest && originalRequest._retry !== true && !isRefreshCall) {
      originalRequest._retry = true;
      try {
        // Refresh без заголовка Authorization (httpOnly cookie 'rt').
        const { data } = await axios.post<{ access_token: string }>(
          `${api.defaults.baseURL}/auth/refresh`,
          undefined,
          { withCredentials: true },
        );
        setAccessToken(data.access_token);
        return await api.request(originalRequest);
      } catch {
        setAccessToken(null);
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        // Исходную 401 нормализуем в ApiError.
        throw toApiError(error);
      }
    }

    // Ошибки нормализуются в ApiError.
    throw toApiError(error);
  },
);
