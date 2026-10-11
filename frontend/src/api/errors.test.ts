import { describe, expect, it } from 'vitest';
import { AxiosError, AxiosResponse } from 'axios';
import { ApiError, toApiError } from './errors';

function axiosErrorWithData(data: unknown, status: number, url = '/trips/1'): AxiosError {
  const config: unknown = { url };
  const response: AxiosResponse = {
    data,
    status,
    statusText: String(status),
    headers: {},
    config: config as AxiosResponse['config'],
  } as unknown as AxiosResponse;
  return new AxiosError(
    'Request failed',
    String(status),
    config as AxiosError['config'],
    undefined,
    response,
  );
}

describe('toApiError', () => {
  it('разбирает ApiErrorBody из axios-ответа', () => {
    const e = toApiError(
      axiosErrorWithData(
        { error: { code: 'validation', message: 'bad payload', details: { field: 'title' } } },
        400,
      ),
    );
    expect(e).toBeInstanceOf(ApiError);
    expect(e.code).toBe('validation');
    expect(e.message).toBe('bad payload');
    expect(e.details).toEqual({ field: 'title' });
  });

  it('not_found из тела сохраняется', () => {
    const e = toApiError(
      axiosErrorWithData({ error: { code: 'not_found', message: 'no trip' } }, 404),
    );
    expect(e.code).toBe('not_found');
    expect(e.message).toBe('no trip');
  });

  it('не-axios Error -> code=internal с сообщением', () => {
    const e = toApiError(new Error('boom'));
    expect(e).toBeInstanceOf(ApiError);
    expect(e.code).toBe('internal');
    expect(e.message).toBe('boom');
  });

  it('не-Error значения -> code=internal', () => {
    expect(toApiError('string failure').code).toBe('internal');
    expect(toApiError(42).code).toBe('internal');
    expect(toApiError(undefined).code).toBe('internal');
    expect(toApiError({ weird: true }).code).toBe('internal');
  });

  it('axios-ошибка без тела -> code=internal', () => {
    const e = toApiError(axiosErrorWithData(undefined, 502));
    expect(e.code).toBe('internal');
  });

  it('ApiErrorBody без code в error -> internal', () => {
    const e = toApiError(axiosErrorWithData({ error: { message: 'x' } } as never, 400));
    expect(e.code).toBe('internal');
  });
});
