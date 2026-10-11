export class ApiError extends Error {
  constructor(public code: string, message: string, public details?: unknown) { super(message); }
}

export const toApiError = (err: unknown): ApiError => {
  if (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object') {
    const response = err.response
    if ('data' in response && response.data && typeof response.data === 'object' && 'error' in response.data) {
      const errorData = response.data.error
      if (typeof errorData === 'object' && errorData !== null && 'code' in errorData && 'message' in errorData) {
        return new ApiError(
          errorData.code,
          errorData.message,
          errorData.details
        )
      }
    }
  }
  
  // Default error
  return new ApiError('internal', 'Internal server error')
}