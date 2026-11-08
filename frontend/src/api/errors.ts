export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public details?: unknown,
  ) {
    super(message);
  }
}

export function toApiError(err: unknown): ApiError {
  if (err && typeof err === 'object' && err !== null) {
    const axiosError = err as any;
    if ('response' in axiosError && axiosError.response !== null) {
      const response = axiosError.response;
      if (
        'data' in response &&
        response.data &&
        typeof response.data === 'object' &&
        'error' in response.data &&
        response.data.error !== null
      ) {
        const errorData = response.data.error;
        return new ApiError(errorData.code, errorData.message, errorData.details);
      }
    }
  }

  // Fallback for non-axios errors
  if (err instanceof Error) {
    return new ApiError('internal', err.message);
  }

  return new ApiError('internal', 'Unknown error');
}
