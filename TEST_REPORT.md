# TEST_REPORT.md — FIX-A7-03-01

## Статус: PASS (приёмка повторного раунда)

Задача: привести клиентские функции updateStop/deleteStop в frontend к бэкенд-контракту A7-03 (PUT/DELETE `/trips/{tripId}/stops/{stopId}`), устранить гарантированный 404. Бэкенд не трогать.

## Верификация по объёму (сверено с кодом)

1. `frontend/src/api/stops.ts` — PASS
   - `updateStop(tripId, stopId, p)`: `api.put<Stop>(`/trips/${tripId}/stops/${stopId}`, p)` → `Promise<Stop>` (200). Соответствует контракту.
   - `deleteStop(tripId, stopId)`: `api.delete(`/trips/${tripId}/stops/${stopId}`)` → `Promise<void>` (204). Соответствует контракту.
   - Комментарии актуальны: `// PUT /trips/{tripId}/stops/{stopId} -> 200 Stop (контракт A7-03)` / `// DELETE ... -> 204`.
   - `listStops`/`createStop` не тронуты.
2. `frontend/src/pages/TripDetail.tsx` — PASS
   - `updateStopMut.mutationFn`: `updateStop(id!, stopId, payload)`.
   - `deleteStopMut.mutationFn`: `deleteStop(id!, stopId)`.
   - `onSuccess`: invalidateStops по ключу `['stops', id]` — сохранён.
3. `frontend/src/components/StopForm.tsx` — PASS
   - Режим редактирования: `await updateStop(tripId, initial.id, payload)`; `tripId` приходит через `StopFormProps`.
   - `createStop` не тронут.
4. `frontend/src/lib/types.ts` — `StopPayload` не тронут, совпадает со `StopJSON`.

## Результаты приёмки (повторный раунд)

| Проверка | Команда | Результат |
|---|---|---|
| Сборка | `make frontend-build` | PASS (exit 0, vite build ✓, 211 modules) |
| Стиль | `make frontend-lint` (prettier+eslint) | PASS (exit 0) |
| Контрактные тесты A7-03 | `npx vitest run src/api/stops.test.ts` | PASS 5/5 (PUT/DELETE пути, 200/204, StopPayload-тело, сигнатуры) |

Git: рабочее дерево чистое (`git status --porcelain` пусто), правки закоммичены (HEAD fba398a).

## Ожидаемые падения (не блокируют приёмку задачи, по ТЗ не трогать)

- `TripDetail.test.tsx` / `StopForm.test.tsx` (тест «updateStop при редактировании») — сигнатуры `updateStop/deleteStop` изменились; тесты перепишет задача FIX-A7-03-02. В финале эпика весь пакет должен быть зелёным.
- Деградации окружения вне объёма задачи (зафиксированы ранее): e2e-спеки собираются vitest'ом вместо playwright-runner (покрывается `make e2e` в CI); backend `go mod tidy` (go.mod degrade, бэкенд не в объёме задачи).

## Вывод

Код по объёму задачи соответствует контракту A7-03 посимвольно (пути, статус-коды, типы, комментарии). make frontend-build && make frontend-lint — зелёно. Задача переводится в testing.
