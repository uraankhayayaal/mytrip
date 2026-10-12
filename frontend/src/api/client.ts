import axios, { AxiosInstance } from 'axios';

let accessToken: string | null = null;

export const api: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
});

export const setAccessToken = (t: string | null): void => {
  accessToken = t;
};

export const getAccessToken = (): string | null => {
  return accessToken;
};

// Request interceptor
api.interceptors.request.use(
  (config) => {
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshResponse = await api.post('/auth/refresh');
        const newAccessToken = refreshResponse.data.access_token;

        setAccessToken(newAccessToken);

        // Retry original request with new token
        return api.request(originalRequest);
      } catch (refreshError) {
        setAccessToken(null);
        window.location.href = '/login';
      }
    }

    return Promise.reject(error);
  },
);
