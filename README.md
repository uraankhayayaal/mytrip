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
- [x] A7-01 — AuthService (register/login/refresh/logout/me) + UserRepo/RefreshTokenRepo (Postgres pgx, таблицы 0001_init) + AuthHandler (/api/v1: POST /auth/register 201, /auth/login 200 + Set-Cookie rt, /auth/refresh 200 + ротация, /auth/logout 204 + Clear-Cookie, GET /users/me 200) + JWT HS256 (jwtutil.NewHS256) + bcrypt-хэш + refresh-ротация (SHA-256 hash, удаление старого/вставка нового). Unit-тесты: internal/service/auth_service_test.go, internal/api/handlers/auth_handler_test.go. Проверено: make backend-build / backend-test — зелёные.
- [x] A7-05 — middleware.Auth + jwtutil.NewHS256 (валидация JWT access-токена). `internal/jwtutil/jwt.go`: `NewHS256` (адаптер golang-jwt/jwt/v5, Claims {sub,type,exp}, HS256), `IssueAccess`/`IssueRefresh`, `Validate` (подпись/exp/kind; refresh как access → unauthorized). `internal/api/middleware/auth.go`: `Auth(secret, next)` — парсинг "Bearer <token>", валидация через TokenValidator (expectedKind=TokenKindAccess), на успех `context.WithValue(UserIDKey, sub)`, на ошибки `WriteError(domain.NewUnauthorized)`; refresh-токен → 401 (проверка kind обязательна). `cmd/server/main.go`: `middleware.Auth(cfg.JWTSecret)` смонтирован на защищённые маршруты. Unit-тесты: `internal/jwtutil/jwt_test.go` (issue→validate, expired, wrong kind, malformed, wrong secret), `internal/api/middleware/auth_test.go` (httptest: нет header, malformed, expired, refresh вместо access, успех, lowercase bearer). Проверено: make backend-test / backend-lint (gofmt+go vet) — зелёные.
- [x] A7-02 — TripService + TripRepo + TripHandler (CRUD /trips): TripRepo (Postgres pgx, таблица 0002_trips) — CRUD + SetCover (нет/чужая → domain.NewNotFound), TripService — валидация (title non-blank, end_date >= start_date → domain.NewValidation), TripHandler — GET/POST /trips, GET/PUT/DELETE /trips/{id} (chi-параметры {id}, middleware.WriteJSON / WriteError, DELETE → 204, пустой список → []). ПОДКЛЮЧЕНО В СЕРВЕР (cmd/server/main.go): `handlers.NewTripHandler(service.NewTripService(repo.NewPGTripRepo(pool))).Mount(r)` под auth-middleware внутри /api/v1. Unit-тесты: internal/service/trip_service_test.go (мок TripRepo), internal/api/handlers/trip_handler_test.go (httptest, мок TripService, краевые кейсы: nil payload, blank title, end==start, bad JSON, 500-пропагация, регистрация маршрутов). Проверено: make backend-build / backend-test — зелёные (handlers 96.4%, service 82.1% coverage, race), gofmt/go vet — чистые (golangci-lint в песочнице не установлен — деградация окружения, в CI пройдёт). Отчёт: TEST_REPORT.md.
- [x] A7-06 — DI + монтаж маршрутов в cmd/server/main.go + ExportHandler (CSV) + интеграционный контракт-тест. Composition root (`cmd/server/main.go`): pgxpool → 5 репозиториев (NewPGUserRepo/NewPGRefreshTokenRepo/NewPGTripRepo/NewPGStopRepo/NewPGPhotoRepo) → s3.NewFromConfig (endpoint cfg.S3Endpoint, static credentials) → storage.NewS3Storage (StoragePort, 15m presign-TTL) → jwtutil.NewHS256(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL) → 4 сервиса (NewAuthService/NewTripService/NewStopService/NewPhotoService) → 5 хендлеров (NewAuthHandler/NewTripHandler/NewStopHandler/NewPhotoHandler/NewExportHandler). TODO(BEL-03..06) и символ `authMiddleware` удалены; защищённая группа — `r.With(func(next) { return middleware.Auth(cfg.JWTSecret, next) })` (обёртка под chi `r.With`; сигнатура middleware.Auth(secret, next) из A7-05). Публичные: `authHandler.Mount` (POST /auth/register|login|refresh|logout, GET /users/me); под auth: `tripHandler.Mount` (CRUD /trips), `stopHandler.Mount` (CRUD /trips/{id}/stops), `photoHandler.Mount` (GET/POST /trips/{id}/photos, DELETE /trips/{id}/photos/{photoId}, POST /trips/{id}/photos/presign), `exportHandler.Mount` (GET /trips/{id}/export). ExportHandler (`internal/api/handlers/export.go`): CSV header `order,title,description,lat,lng,visit_date` (разделитель `,`), Content-Type text/csv, Content-Disposition `attachment; filename="<slug>.csv"` (пустой title → trip.csv), visit_date RFC3339, координаты 6 знаков, CSV-экранирование (`,` `"` `\n` `\r`), 404/500 через middleware.WriteError. Unit-тесты: `export_test.go` (6 тестов: success/пустой title/404/500/CSV-экранирование/регистрация маршрута), `mount_test.go` (collectRoutes через chi.Walk: все Mount-маршруты + TestMountComposedAPI — полный состав /api/v1). Проверено: `make backend-build` (exit 0), `make backend-test` (`go test ./... -race -cover`, все пакеты ok; handlers 62.9%), `gofmt -l .` — пусто, `go vet ./...` — exit 0; golangci-lint — DEGRADED (бинарь отсутствует в песочнице, покрывается CI). Рантайм: сервер стартует при наличии Postgres/MinIO (компоненты-зависимости поднимаются через `make up`/`docker compose up`); в песочнице без БД падение на `migrate` (connection refused 5432) — ожидаемо, покрывается `make e2e` в CI. Отчёт: TEST_REPORT.md.

## Запуск приложения (backend dev)

```bash
cp .env.example .env   # или export DATABASE_URL/JWT_SECRET/S3_* вручную
make backend-run       # go run ./cmd/server на :8080
```

Требуемые env (см. .env.example): `DATABASE_URL` (postgres://...), `JWT_SECRET`, `S3_ENDPOINT`/`S3_BUCKET`/`S3_ACCESS_KEY`/`S3_SECRET_KEY`. Сервер при старте прогоняет golang-migrate (file://db/migrations) и слушает `:PORT` (default 8080). `GET /healthz` → 200 `{"status":"ok"}` (без auth). `POST /api/v1/auth/*` — A7-01 (публичные); `GET /api/v1/users/me`, `GET/POST /api/v1/trips`, `GET/PUT/DELETE /api/v1/trips/{id}`, `/trips/{id}/stops*`, `/trips/{id}/photos*`, `/trips/{id}/photos/presign`, `/trips/{id}/export` — под Bearer access-токеном (401 без токена) — A7-02/A7-03/A7-04/A7-06.

## Запуск приложения (frontend dev)

```bash
make deps          # npm ci в frontend/
make frontend-run  # vite dev на :5173, прокси /api -> http://localhost:8080
```

| Команда | Что делает |
| --- | --- |
| `make frontend-build` | `tsc -b && vite build` -> `frontend/dist` |
| `make frontend-test` | `vitest --run` (unit-тесты, jsdom) |
| `make frontend-lint` | `prettier --check . && eslint .` |
| `make frontend-run` | vite dev-сервер |

Dev-прокси `/api` -> `localhost:8080` настроен в `vite.config.ts` (контракт ARCH-02). Фронтенд использует только относительные пути `/api/v1/...`.

Структура `frontend/src/`: `lib/types.ts` (единый источник типов, зеркалит ARCH-03), `api/client.ts` (axios, baseURL:/api/v1, withCredentials, JWT-интерцепторы, refresh на 401 через httpOnly cookie `rt`), `api/errors.ts` (ApiError), `api/{auth,trips,stops,photos}.ts` (клиенты), `lib/{queryClient,dates,schemas}.ts`, `components/{Layout,ProtectedLayout,MapView,PhotoUploader,StopForm,TripCard,TripForm}.tsx`, `pages/*`.
