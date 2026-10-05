import type { ApiErrorBody } from '../lib/types';

// Единая клиентская ошибка API: нормализует бэкенд-контракт {error:{code,message,details?}}.
export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public details?: unknown,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// toApiError разбирает ApiErrorBody из axios-ответа; для не-axios ошибок code='internal'.
export function toApiError(err: unknown): ApiError {
  const body: unknown = axiosErrorBody(err);
  if (body !== null && typeof body === 'object') {
    const b = body as ApiErrorBody;
    const errObj = b?.error;
    if (errObj && typeof errObj === 'object' && typeof errObj.code === 'string') {
      return new ApiError(errObj.code, errObj.message, errObj.details);
    }
  }
  const msg =
    err instanceof Error ? err.message : typeof err === 'string' && err ? err : 'internal error';
  return new ApiError('internal', msg);
}

// axiosErrorBody достаёт тело (response.data) из axios-ошибки (duck-typing, без импорта axios).
function axiosErrorBody(err: unknown): unknown {
  if (err && typeof err === 'object' && 'response' in err) {
    const response = (err as { response?: unknown }).response;
    if (response && typeof response === 'object' && 'data' in response) {
      return (response as { data?: unknown }).data;
    }
  }
  return null;
}
