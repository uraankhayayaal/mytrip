import { describe, it, expect } from 'vitest';
import { toApiError } from '../api/errors';
import { ApiErrorBody } from '../lib/types';

describe('toApiError', () => {
  it('should parse ApiErrorBody correctly', () => {
    const errorBody: ApiErrorBody = {
      error: {
        code: 'unauthorized',
        message: 'Unauthorized access',
        details: { field: 'password' },
      },
    };

    const error = toApiError({ response: { data: errorBody } });
    expect(error.code).toBe('unauthorized');
    expect(error.message).toBe('Unauthorized access');
    expect(error.details).toEqual({ field: 'password' });
  });

  it('should handle non-axios errors', () => {
    const error = toApiError(new Error('Something went wrong'));
    expect(error.code).toBe('internal');
    expect(error.message).toBe('Something went wrong');
  });

  it('should handle unknown error types', () => {
    const error = toApiError('unknown error');
    expect(error.code).toBe('internal');
    expect(error.message).toBe('Unknown error');
  });
});
