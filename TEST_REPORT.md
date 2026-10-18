# Тест-отчёт — FIX-A7-03-01 (контракт A7-03: updateStop/deleteStop + callers)

## Вердикт: PASS

Все 4 пункта объёма задачи выполнены в рабочей ветке `ai/task/FIX-A7-03-01`
и зафиксированы коммитами: `updateStop(tripId, stopId, p)` →
`PUT /trips/{tripId}/stops/{stopId}`, `deleteStop(tripId, stopId)` →
`DELETE /trips/{tripId}/stops/{stopId}`, callers TripDetail/StopForm передают
tripId, комментарии над функциями актуальны. Бэкенд не тронут.

Контрактные автотесты QA: `frontend/src/api/stops.test.ts` (5 тестов, реальный
axios с подменённым адаптером — проверяются метод/URL/тело запроса, уходящего в сеть).

## Пройденные проверки (перепрогнаны в раунде 46, 2026-10-18)

| Проверка | Команда | Результат |
|---|---|---|
| Сборка frontend | `make frontend-build` | **PASS** (exit 0, vite build ~6.5s) |
| Линт frontend | `make frontend-lint` (prettier+eslint) | **PASS** (exit 0) |
| Контрактные тесты A7-03 | `npm --prefix frontend test -- --run src/api/stops.test.ts` | **PASS** 5/5 |

### Детали контрактных тестов (все 5 зелёные)
- `listStops` → `GET /trips/{tripId}/stops` 200 Stop[] (не тронута)
- `createStop` → `POST /trips/{tripId}/stops` 201 Stop (не тронута)
- `updateStop` → `PUT /trips/{tripId}/stops/{stopId}` 200 Stop, тело = StopPayload — **НЕ** `/stops/{stopId}` (старый путь → гарантированный 404)
- `updateStop('t-2','s-2',p)` → URL `/trips/t-2/stops/s-2` — фиксация сигнатуры (tripId первым)
- `deleteStop` → `DELETE /trips/{tripId}/stops/{stopId}` 204, возвращает void

## Ожидаемые падения (по постановке задачи — не блокируют)

`TripDetail.test.tsx` (4) и `StopForm.test.tsx` (1, «вызывает updateStop при
редактировании») упали на изменении сигнатур — **ожидаемо**, эти тесты
перепишет задача FIX-A7-03-02; в рамках FIX-A7-03-01 тесты не правились (запрещено ТЗ).

## Падения вне объёма задачи (не вызваны FIX-A7-03-01)

1. `e2e/auth.spec.ts`, `e2e/trips.spec.ts` — vitest пытается собрать Playwright-e2e
   (нужен отдельный runner/браузеры + поднятый стек `make e2e`); в sandbox degrades.
2. Остальные unit-тесты (TripCard/TripsList/Login/TripEdit/Register) падают
   с `useNavigate() may be used only in the context of a <Router> component` —
   рендер без MemoryRouter: предсуществующий дефект тестовой обвязки, не зависит от
   правок A7-03. Не входит в объём FIX-A7-03-01.
3. backend (`make run`): `go: updates to go.mod needed; go mod tidy` — degrade
   окружения, не объём задачи (бэкенд не трогать).

## Итог

Критерии приёмки задачи выполнены: `make frontend-build && make frontend-lint` — зелёно;
клиентские функции updateStop/deleteStop соответствуют контракту A7-03 до последнего
символа (проверено автотестами). Дефект 404 устранён. Задача переведена в `testing`.
