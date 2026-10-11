import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import axios, { AxiosAdapter, AxiosError, AxiosResponse } from 'axios';
import { api, getAccessToken, setAccessToken } from './client';
import { ApiError } from './errors';

// Тестируем через РЕАЛЬНЫЙ axios: заменяем адаптер (axios.defaults.adapter + api.defaults.adapter),
// без msw/внешней сети.

type AdapterHandler = (config: any) => Promise<AxiosResponse>;

function installAdapter(handler: AdapterHandler): void {
  const adapter: AxiosAdapter = (config) => handler(config);
  axios.defaults.adapter = adapter;
  api.defaults.adapter = adapter;
}

// Мок window.location (jsdom): перехватываем window.location.href = '/login'.
function fakeLocation(): void {
  const fake = {
    href: 'http://localhost/',
    origin: 'http://localhost',
    protocol: 'http:',
    host: 'localhost',
    hostname: 'localhost',
    port: '',
    pathname: '/',
    search: '',
    hash: '',
    assign: vi.fn(),
    reload: vi.fn(),
    replace: vi.fn(),
  } as unknown as Location;
  Object.defineProperty(window, 'location', {
    value: fake,
    writable: true,
    configurable: true,
  });
}

// AxiosHeaders хранит ключи в нижнем регистре; проверяем и метод get, и свойства.
function readHeader(headers: unknown, name: string): string | undefined {
  if (!headers || typeof headers !== 'object') return undefined;
  const record = headers as Record<string, unknown>;
  for (const candidate of [record[name], record[name.toLowerCase()]]) {
    if (typeof candidate === 'string') return candidate;
  }
  const get = (headers as { get?: unknown }).get;
  if (typeof get === 'function') {
    const value = (get as (h: string) => unknown).call(headers, name);
    if (typeof value === 'string') return value;
  }
  return undefined;
}

function unauthorizedError(config: any, message: string): AxiosError {
  return new AxiosError('Request failed', '401', config, undefined, {
    data: { error: { code: 'unauthorized', message } },
    status: 401,
    statusText: 'Unauthorized',
    headers: {},
    config,
  } as unknown as AxiosResponse);
}

describe('api client (JWT-интерцепторы)', () => {
  let savedLocation: PropertyDescriptor | undefined;

  beforeEach(() => {
    setAccessToken(null);
    savedLocation = Object.getOwnPropertyDescriptor(window, 'location');
    fakeLocation();
  });

  afterEach(() => {
    if (savedLocation) {
      Object.defineProperty(window, 'location', savedLocation);
    }
  });

  it('(a) request-интерцептор добавляет Authorization: Bearer при accessToken', async () => {
    let captured: any;
    installAdapter(async (config) => {
      captured = config;
      return {
        data: { ok: true },
        status: 200,
        statusText: 'OK',
        headers: {},
        config,
      } as unknown as AxiosResponse;
    });

    setAccessToken('test-jwt');
    const res = await api.get('/healthz');
    expect(res.status).toBe(200);
    expect(readHeader(captured.headers, 'Authorization')).toBe('Bearer test-jwt');
  });

  it('без accessToken — заголовок Authorization не ставится', async () => {
    let captured: any;
    installAdapter(async (config) => {
      captured = config;
      return {
        data: {},
        status: 200,
        statusText: 'OK',
        headers: {},
        config,
      } as unknown as AxiosResponse;
    });

    setAccessToken(null);
    await api.get('/healthz');
    expect(readHeader(captured.headers, 'Authorization')).toBeUndefined();
  });

  it('(b) 401 -> успешный refresh -> повторный запрос возвращает данные', async () => {
    let refreshCalls = 0;
    let tripsCalls = 0;
    let lastAuth: string | undefined;
    let lastUrl = '';
    let lastBase: string | undefined;

    installAdapter(async (config) => {
      const url = String(config.url ?? '');
      if (url.includes('/auth/refresh')) {
        refreshCalls += 1;
        return {
          data: { access_token: 'new-jwt' },
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        } as unknown as AxiosResponse;
      }
      if (url.includes('/trips')) {
        tripsCalls += 1;
        lastUrl = url;
        lastBase = config.baseURL;
        lastAuth = readHeader(config.headers, 'Authorization');
        if (tripsCalls === 1) {
          throw unauthorizedError(config, 'invalid token');
        }
        return {
          data: [{ id: 't1', title: 'Trip' }],
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        } as unknown as AxiosResponse;
      }
      return {
        data: {},
        status: 404,
        statusText: 'NF',
        headers: {},
        config,
      } as unknown as AxiosResponse;
    });

    setAccessToken('expired-jwt');
    const res = await api.get('/trips');

    expect(res.status).toBe(200);
    expect(res.data).toEqual([{ id: 't1', title: 'Trip' }]);
    expect(refreshCalls).toBe(1);
    expect(tripsCalls).toBe(2);
    expect(lastAuth).toBe('Bearer new-jwt');
    // Повторный запрос (retry) сохраняет относительный путь и baseURL без дублей:
    // реальный XHR-адаптер compose-ит baseURL + url, поэтому в адаптере url остаётся '/trips'.
    expect(lastUrl).toBe('/trips');
    expect(lastBase).toBe('/api/v1');
    expect(getAccessToken()).toBe('new-jwt');
  });

  it('(c) неудачный refresh -> сброс токена и редирект на /login', async () => {
    installAdapter(async (config) => {
      const url = String(config.url ?? '');
      if (url.includes('/auth/refresh')) {
        throw unauthorizedError(config, 'refresh failed');
      }
      throw unauthorizedError(config, 'invalid token');
    });

    setAccessToken('expired-jwt');
    let thrown: unknown;
    try {
      await api.get('/trips');
    } catch (e) {
      thrown = e;
    }

    expect(thrown).toBeInstanceOf(ApiError);
    expect((thrown as ApiError).code).toBe('unauthorized');
    expect(getAccessToken()).toBeNull();
    expect(window.location.href).toBe('/login');
  });

  it('ошибка не-401 нормализуется в ApiError с code из тела', async () => {
    installAdapter(async (config) => {
      return Promise.reject(
        new AxiosError('Request failed', '404', config, undefined, {
          data: { error: { code: 'not_found', message: 'trip not found' } },
          status: 404,
          statusText: 'Not Found',
          headers: {},
          config,
        } as unknown as AxiosResponse),
      );
    });

    setAccessToken('tok');
    let thrown: unknown;
    try {
      await api.get('/trips/unknown');
    } catch (e) {
      thrown = e;
    }

    expect(thrown).toBeInstanceOf(ApiError);
    expect((thrown as ApiError).code).toBe('not_found');
    expect((thrown as ApiError).message).toBe('trip not found');
  });
});
