import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'
import { api, setAccessToken, getAccessToken } from './client'

// Mock axios
vi.mock('axios')

const mockedAxios = axios as vi.Mocked<typeof axios>

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setAccessToken(null)
  })

  it('should set and get access token', () => {
    setAccessToken('test-token')
    expect(getAccessToken()).toBe('test-token')
    
    setAccessToken(null)
    expect(getAccessToken()).toBeNull()
  })

  it('should add Authorization header when token exists', () => {
    setAccessToken('test-token')
    
    // Mock a request config
    const config = { headers: {} }
    
    // This would be called by the interceptor
    const result = api.interceptors.request.handlers[0].fulfilled(config)
    
    expect(result.headers.Authorization).toBe('Bearer test-token')
  })

  it('should not add Authorization header when no token', () => {
    const config = { headers: {} }
    
    // This would be called by the interceptor
    const result = api.interceptors.request.handlers[0].fulfilled(config)
    
    expect(result.headers.Authorization).toBeUndefined()
  })

  it('should handle 401 error with refresh', async () => {
    // Mock the API with a 401 response
    mockedAxios.post.mockRejectedValueOnce({
      response: { status: 401 },
      config: { url: '/api/v1/test' }
    })
    
    // Mock refresh response
    mockedAxios.post.mockResolvedValueOnce({
      data: { access_token: 'new-token' }
    })
    
    // Mock the request function
    mockedAxios.request.mockResolvedValueOnce({ data: 'success' })
    
    // Mock window.location
    vi.spyOn(window.location, 'href', 'set')
    
    try {
      await api.get('/test')
    } catch (error) {
      // We expect this to throw because we're mocking a 401
      expect(error).toBeDefined()
    }
    
    // Check that refresh was called
    expect(mockedAxios.post).toHaveBeenCalledWith('/auth/refresh')
  })

  it('should redirect to login on failed refresh', async () => {
    // Mock the API with a 401 response
    mockedAxios.post.mockRejectedValueOnce({
      response: { status: 401 },
      config: { url: '/api/v1/test' }
    })
    
    // Mock refresh to fail
    mockedAxios.post.mockRejectedValueOnce({
      response: { status: 401 }
    })
    
    // Mock window.location
    const locationSpy = vi.spyOn(window.location, 'href', 'set')
    
    try {
      await api.get('/test')
    } catch (error) {
      // We expect this to throw because refresh fails
      expect(error).toBeDefined()
    }
    
    // Check that we redirected to login
    expect(locationSpy).toHaveBeenCalledWith('/login')
  })
})