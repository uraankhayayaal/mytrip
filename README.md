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
- `e2e` — самозавершающийся прогон: `up` → healthcheck `http://localhost:8080/healthz` (до 60s) → `npm --prefix e2e test` → `down` → exit code прогона. Цели `up`/`down` добавит DOL-03 (ARCH-02).

## План работ

- [x] BEL-01 — корневой Makefile: help/deps/build/test/lint/run/e2e + backend-*/frontend-* (завершено).
- [x] BEL-03 — SQL-миграции golang-migrate в db/migrations/: 0001_init (users, refresh_tokens, pgcrypto), 0002_trips, 0003_stops, 0004_photos (up/down). Проверено: все 5 таблиц создаются без ошибок, down-миграции откатывают схему.
- [x] FEL-01 — Каркас SPA (frontend/): Vite+TS+Tailwind, React Router v6, Layout/ProtectedLayout, axios-клиент /api/v1 с JWT-интерцептором (refresh на 401), ApiError, клиенты auth/trips/stops/photos, QueryClient, страницы-заглушки, unit-тесты (vitest+jsdom). Проверено: make frontend-build / frontend-test / frontend-lint зелёные.
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
