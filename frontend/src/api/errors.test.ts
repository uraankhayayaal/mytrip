import { describe, it, expect } from 'vitest'
import { toApiError, ApiError } from './errors'

describe('toApiError', () => {
  it('should parse ApiErrorBody correctly', () => {
    const errorData = {
      error: {
        code: 'unauthorized',
        message: 'Unauthorized access',
        details: { field: 'email' }
      }
    }
    
    const error = toApiError({ response: { data: errorData } })
    
    expect(error).toBeInstanceOf(ApiError)
    expect(error.code).toBe('unauthorized')
    expect(error.message).toBe('Unauthorized access')
    expect(error.details).toEqual({ field: 'email' })
  })

  it('should return internal error for non-axios errors', () => {
    const error = toApiError(new Error('Network error'))
    
    expect(error).toBeInstanceOf(ApiError)
    expect(error.code).toBe('internal')
    expect(error.message).toBe('Internal server error')
  })

  it('should return internal error for invalid error data', () => {
    const error = toApiError({ response: { data: { invalid: 'data' } } })
    
    expect(error).toBeInstanceOf(ApiError)
    expect(error.code).toBe('internal')
    expect(error.message).toBe('Internal server error')
  })
})