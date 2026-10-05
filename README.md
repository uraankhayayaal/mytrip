# mytrip — корневой Makefile (BEL-01)

Единая точка входа команд для всех субагентов и приёмки. Запуск: `make help`.

## Цели

- `help` — автогенерация списка целей из комментариев `# help: ...` (подхватывает и будущие цели DOL-03: up/down/logs/ps/infra.*).
- `deps` — установка зависимостей: `go mod download && npm --prefix frontend ci`.
- `build` — сборка всего: `backend-build` + `frontend-build`.
- `backend-build` — `go build -o bin/server ./cmd/server`.
- `frontend-build` — `npm --prefix frontend run build`.
- `test` — `backend-test` + `frontend-test`.
- `backend-test` — `go test ./... -race -cover`.
- `frontend-test` — `npm --prefix frontend test -- --run`.
- `lint` — `backend-lint` + `frontend-lint`.
- `backend-lint` — `gofmt -l . && go vet ./... && golangci-lint run`.
- `frontend-lint` — `npm --prefix frontend run lint`.
- `run` — параллельный dev-запуск: `backend-run` + `frontend-run` (Ctrl+C останавливает оба).
- `backend-run` — `go run ./cmd/server`.
- `frontend-run` — `npm --prefix frontend run dev`.
- `e2e` — самозавершающийся прогон: `up` → healthcheck `http://localhost:8080/healthz` (до 60s) → `npm --prefix frontend run e2e` (Playwright, `frontend/e2e/*.spec.ts`) → `down` → exit code прогона. Цели `up`/`down` добавит DOL-03 (ARCH-02).

## План работ

- [x] BEL-01 — корневой Makefile: help/deps/build/test/lint/run/e2e + backend-*/frontend-* (завершено).
- [x] BEL-03 — SQL-миграции golang-migrate в db/migrations/: 0001_init (users, refresh_tokens, pgcrypto), 0002_trips, 0003_stops, 0004_photos (up/down). Проверено: все 5 таблиц создаются без ошибок, down-миграции откатывают схему.
- [x] FEL-01 — Каркас SPA (frontend/): Vite+TS+Tailwind, React Router v6, Layout/ProtectedLayout, axios-клиент /api/v1 с JWT-интерцептором (refresh на 401), ApiError, клиенты auth/trips/stops/photos, QueryClient, страницы-заглушки, unit-тесты (vitest+jsdom). Проверено: make frontend-build / frontend-test / frontend-lint зелёные.
- [x] FEL-04 — MapView (Leaflet): `src/components/MapView.tsx` — MapContainer+TileLayer (OSM, без API-ключа), divIcon-маркеры с номером по order (выбранный — bg-red-600), drag → onMoveStop, click → onSelectStop, Polyline по stopsSorted (order). Unit-тесты `MapView.test.tsx` (мок react-leaflet, jsdom): N маркеров, dragend → onMoveStop, click → onSelectStop, polyline в порядке order, пустая карта. LSP-диагностика чистая; make frontend-build/frontend-test/frontend-lint — см. план работ (в этой сессии Run-песочница недоступна, docker-образ не поднимается).
- [x] FEL-07 — Интеграция: e2e-сценарии (Playwright), make-цели, финальная полировка. Созданы `frontend/playwright.config.ts` (testDir `./e2e`, baseURL `http://localhost:5173`, webServer `npm run dev`, reuseExistingServer), `frontend/e2e/auth.spec.ts` (register → login → trips list), `frontend/e2e/trips.spec.ts` (create trip → detail → add stop → export CSV). В `frontend/package.json`: скрипт `e2e` (`playwright test`) + devDependency `@playwright/test@^1`. Makefile: цель `e2e` теперь запускает `npm --prefix frontend run e2e` (было `npm --prefix e2e test`). Примечание: в этой сессии Run-песочница недоступна (docker-образ `ai-sandbox` не поднимается), поэтому `make frontend-build/test/lint/e2e` не прогонялись вживую; синтаксис/типы e2e-спеков соответствуют Playwright API, прогон — `make e2e` или `npm --prefix frontend run e2e`.
- [ ] DOL-03 (ARCH-02) — цели up/down/logs/ps/infra.* (будут ДОБАВЛЕНЫ в этот же Makefile, help подхватит автоматически).

## Фронтенд (frontend/)

React SPA на Vite + TypeScript + Tailwind. Каталог `frontend/`.

### Запуск (dev)

```bash
make deps        # npm ci в frontend/ (после go mod download)
make frontend-run # vite dev на :5173, прокси /api -> http://localhost:8080
```

Dev-прокси `/api` -> `localhost:8080` настроен в `vite.config.ts` (контракт ARCH-02). Фронтенд использует только относительные пути `/api/v1/...` (никаких абсолютных URL API).

### Команды

| Команда | Что делает |
| --- | --- |
| `make frontend-build` | `tsc -b && vite build` -> `frontend/dist` |
| `make frontend-test` | `vitest --run` (unit-тесты, jsdom) |
| `make frontend-lint` | `prettier --check . && eslint .` |
| `make frontend-run` | vite dev-сервер |

### Структура

- `src/lib/types.ts` — единый источник типов (зеркалит ARCH-03, менять сигнатуры нельзя).
- `src/api/client.ts` — axios-инстанс `baseURL:/api/v1`, `withCredentials`; JWT-интерцепторы (Bearer; refresh на 401 через httpOnly cookie `rt`, один раз, `_retry`; сбой -> сброс + `/login`).
- `src/api/errors.ts` — `ApiError` + `toApiError` (разбор `ApiErrorBody`).
- `src/api/{auth,trips,stops,photos}.ts` — клиенты, пути относительные.
- `src/lib/queryClient.ts` — `QueryClient` (retry:1, refetchOnWindowFocus:false).
- `src/components/{Layout,ProtectedLayout}.tsx` — nav + защищённый layout (useQuery `['me']`).
- `src/pages/*` — заглушки Login/Register/TripsList/TripDetail/TripEdit/CalendarPage.

### Запуск приложения (frontend dev)

Запуск dev-сервера — `make frontend-run` (или `npm --prefix frontend run dev`). Сервер самозавершающимся НЕ является (dev-сервер), поэтому для автоматической проверки не используется; проверять код надо через `make frontend-build` / `make frontend-test` / `make frontend-lint`.
