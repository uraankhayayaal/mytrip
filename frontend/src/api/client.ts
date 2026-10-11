import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios'

export let accessToken: string | null = null

export const setAccessToken = (t: string | null): void => {
  accessToken = t
}

export const getAccessToken = (): string | null => {
  return accessToken
}

const api: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  withCredentials: true
})

// Request interceptor
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response
  },
  async (error) => {
    const originalRequest = error.config
    
    // If error is 401 and it's not a retry attempt
    if (error.response?.status === 401 && originalRequest && (originalRequest as any)._retry !== true) {
      // Mark this request as a retry attempt
      (originalRequest as any)._retry = true
      
      try {
        // Try to refresh the token
        const refreshResponse = await api.post('/auth/refresh')
        
        // Set the new access token
        setAccessToken(refreshResponse.data.access_token)
        
        // Retry the original request
        return api.request(originalRequest)
      } catch (refreshError) {
        // If refresh fails, redirect to login
        setAccessToken(null)
        window.location.href = '/login'
        return Promise.reject(error)
      }
    }
    
    // For all other errors, reject with normalized error
    return Promise.reject(error)
  }
)

export default api