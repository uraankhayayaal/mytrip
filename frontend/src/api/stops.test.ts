import { beforeEach, describe, expect, it, vi } from 'vitest';
import axios, { AxiosAdapter, AxiosResponse } from 'axios';
import { api } from './client';
import { createStop, deleteStop, listStops, updateStop } from './stops';
import type { Stop, StopPayload } from '../lib/types';

// Контракт A7-03 (source of truth): бэкенд отдаёт CRUD стопов ТОЛЬКО
// по вложенным маршрутам /trips/{tripId}/stops[/{stopId}]:
//   GET    /trips/{tripId}/stops        -> 200 Stop[]
//   POST   /trips/{tripId}/stops        -> 201 Stop
//   PUT    /trips/{tripId}/stops/{id}   -> 200 Stop
//   DELETE /trips/{tripId}/stops/{id}   -> 204
// Старый путь /stops/{stopId} (без tripId) гарантированно даёт 404 —
// эти тесты фиксируют, что клиент ходит строго по контракту.

type RecordedCall = { method: string; url: string; data?: unknown };

// Реальный axios с подменённым адаптером (паттерн client.test.ts): проверяем
// метод/URL/тело, которые реально уходят в сеть.
function installAdapter(handler: (config: any) => Promise<AxiosResponse>): void {
  const adapter: AxiosAdapter = (config) => handler(config);
  axios.defaults.adapter = adapter;
  api.defaults.adapter = adapter;
}

function jsonResponse(status: number, data: unknown): AxiosResponse {
  return { data, status, statusText: '', headers: {}, config: {} as any };
}

// Axios сериализует JSON-тело в строку в config.data — нормализуем к объекту.
function body(configData: unknown): unknown {
  if (typeof configData === 'string') return JSON.parse(configData);
  return configData;
}

const tripId = 'trip-1';
const stopId = 'stop-9';

const stop: Stop = {
  id: stopId,
  trip_id: tripId,
  title: 'Эрмитаж',
  description: 'Музей',
  location: { lat: 59.9398, lng: 30.3146 },
  visit_date: '2025-06-05',
  order: 0,
  created_at: '2025-05-01T00:00:00Z',
  updated_at: '2025-05-02T00:00:00Z',
};

const payload: StopPayload = {
  title: 'Эрмитаж',
  description: 'Музей',
  location: { lat: 59.9398, lng: 30.3146 },
  visit_date: '2025-06-05',
  order: 0,
};

describe('api/stops — контракт A7-03 (пути CRUD стопов)', () => {
  let calls: RecordedCall[];

  beforeEach(() => {
    vi.clearAllMocks();
    calls = [];
    installAdapter(async (config) => {
      calls.push({ method: config.method as string, url: config.url as string, data: config.data });
      if (config.method === 'delete') return jsonResponse(204, undefined);
      if (config.method === 'put') return jsonResponse(200, stop);
      if (config.method === 'post') return jsonResponse(201, stop);
      return jsonResponse(200, [stop]);
    });
  });

  it('listStops: GET /trips/{tripId}/stops -> 200 Stop[]', async () => {
    const result = await listStops(tripId);
    expect(calls).toHaveLength(1);
    expect(calls[0].method).toBe('get');
    expect(calls[0].url).toBe(`/trips/${tripId}/stops`);
    expect(result).toEqual([stop]);
  });

  it('createStop: POST /trips/{tripId}/stops -> 201 Stop', async () => {
    const result = await createStop(tripId, payload);
    expect(calls).toHaveLength(1);
    expect(calls[0].method).toBe('post');
    expect(calls[0].url).toBe(`/trips/${tripId}/stops`);
    expect(body(calls[0].data)).toEqual(payload);
    expect(result).toEqual(stop);
  });

  it('updateStop: PUT /trips/{tripId}/stops/{stopId} -> 200 Stop (НЕ /stops/{stopId})', async () => {
    const result = await updateStop(tripId, stopId, payload);
    expect(calls).toHaveLength(1);
    expect(calls[0].method).toBe('put');
    expect(calls[0].url).toBe(`/trips/${tripId}/stops/${stopId}`);
    expect(body(calls[0].data)).toEqual(payload);
    expect(result).toEqual(stop);
  });

  it('updateStop: сигнатура (tripId, stopId, payload) — tripId обязателен первым аргументом', async () => {
    // Защита от возврата к старому одноаргументному/неверному порядку аргументов.
    const p: StopPayload = {
      title: 't',
      description: 'd',
      location: { lat: 1, lng: 2 },
      visit_date: '2025-06-01',
      order: 1,
    };
    await updateStop('t-2', 's-2', p);
    expect(calls[0].url).toBe('/trips/t-2/stops/s-2');
  });

  it('deleteStop: DELETE /trips/{tripId}/stops/{stopId} -> 204 (НЕ /stops/{stopId})', async () => {
    const result = await deleteStop(tripId, stopId);
    expect(result).toBeUndefined();
    expect(calls).toHaveLength(1);
    expect(calls[0].method).toBe('delete');
    expect(calls[0].url).toBe(`/trips/${tripId}/stops/${stopId}`);
  });
});
