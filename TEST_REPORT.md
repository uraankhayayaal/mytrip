# TEST_REPORT — FIX-A7-03-01

**Задача:** обновить updateStop/deleteStop в api/stops.ts + callers (TripDetail, StopForm) к контракту A7-03.

## Статус: PASSED

## Верификация по объёму задачи (сверено с кодом в worktree, HEAD 0daa19b)

1. **frontend/src/api/stops.ts**
   - `updateStop(tripId: string, stopId: string, p: StopPayload): Promise<Stop>` — `api.put<Stop>(`/trips/${tripId}/stops/${stopId}`, p)` (200) — соответствует A7-03.
   - `deleteStop(tripId: string, stopId: string): Promise<void>` — `api.delete(`/trips/${tripId}/stops/${stopId}`)` (204) — соответствует A7-03.
   - Комментарии: `// PUT /trips/{tripId}/stops/{stopId} -> 200 Stop (контракт A7-03)`, `// DELETE /trips/{tripId}/stops/{stopId} -> 204 (контракт A7-03)` — актуальны, артефактный путь `/stops/{stopId}` в коде отсутствует.
   - `listStops`/`createStop` не изменены.
2. **frontend/src/pages/TripDetail.tsx**
   - `updateStopMut.mutationFn`: `updateStop(id!, stopId, payload)` — tripId из `useParams`.
   - `deleteStopMut.mutationFn`: `deleteStop(id!, stopId)`.
   - `onSuccess: invalidateStops` по ключу `['stops', id]` — сохранён.
3. **frontend/src/components/StopForm.tsx**
   - Режим редактирования: `await updateStop(tripId, initial.id, payload)`; tripId приходит через `StopFormProps`.
   - `createStop` не тронут.
4. **frontend/src/lib/types.ts** — `StopPayload` не тронут, совпадает со StopJSON.

Дефект гарантированного 404 (клиент ходил на `/stops/{stopId}`) устранён: все мутации идут на `/trips/{tripId}/stops/{stopId}`.

## Команды и результаты

| Команда | Результат |
|---|---|
| `make frontend-build` | PASS (exit 0, tsc -b + vite build, 211 модулей) |
| `make frontend-lint` (prettier + eslint) | PASS (exit 0) |
| `npx vitest run src/api/stops.test.ts` | PASS 5/5 (контрактные тесты A7-03: PUT/DELETE пути, 200/204, StopPayload-тело, сигнатуры) |

Git: правки закоммичены (HEAD 0daa19b); TEST_REPORT.md закоммичен в этом же раунде.

## Ожидаемые падения (по ТЗ не блокируют приёмку, тесты перепишет FIX-A7-03-02)

- `TripDetail.test.tsx` (4 теста) и `StopForm.test.tsx` «updateStop при редактировании» (1 тест) — написаны на старые сигнатуры updateStop/deleteStop; по ТЗ задачи в рамках FIX-A7-03-01 тесты не править.

## Падения вне объёма задачи (зафиксированы, не вызваны A7-03)

- e2e/*.spec.ts собираются vitest'ом вместо playwright-runner — покрывается `make e2e` в CI.
- unit-тесты TripCard/TripsList/Login/TripEdit/Register падают с `useNavigate() may be used only in the context of a <Router> component` — рендер без MemoryRouter, предсуществующий дефект тестовой обвязки.
- backend: `go: updates to go.mod needed` — degrade go.mod, не объём задачи (бэкенд запрещён к изменениям).
