# Тестовый отчёт — A7-06 DI и монтаж маршрутов в cmd/server/main.go (включая /trips/{id}/export)

## Статус: **PASSED** (сборка, тесты, gofmt/go vet); golangci-lint — **DEGRADED** (бинарь отсутствует в песочнице), runtime/e2e — **DEGRADED** (Docker/Postgres/MinIO недоступны в песочнице)

Задача: A7-06, epic CHAT-07, ветка ai/task/A7-06, слой composition root (cmd/server).

## Проверки (запущены в этом раунде)

| Проверка | Команда | Результат |
|---|---|---|
| Сборка бэкенда | `make backend-build` | **Passed** (exit 0) |
| Unit-тесты (race + coverage) | `make backend-test` (`go test ./... -race -cover`) | **Passed** (exit 0, все пакеты ok) |
| gofmt | `gofmt -l .` | **Passed** (пустой вывод) |
| go vet | `go vet ./...` | **Passed** (exit 0) |
| golangci-lint | `golangci-lint run` | **DEGRADED**: бинарь отсутствует в песочнице (`command not found`) — не дефект кода |
| Runtime (`make run` → /healthz 200) | `ReadAppLogs` (go run ./cmd/server) | **DEGRADED**: в песочнице нет Postgres и Docker (`docker: not found`, apt-get без прав root); сервер падает на `migrate: dial tcp 127.0.0.1:5432: connect: connection refused` — миграции запускаются до старта HTTP-сервера. Не дефект кода; покрывается `make e2e` в CI |
| E2E (`make e2e`) | — | **DEGRADED**: Docker недоступен в песочнице — compose-стек (Postgres/MinIO/nginx) поднять нельзя |

## Сверка с контрактом (комментарии-контракты в коде)

1. **DI-композиция в cmd/server/main.go** — выполнена:
   - `pgxpool.New(ctx, cfg.DatabaseURL)` → `pool`;
   - репозитории: `repo.NewPGUserRepo/NewPGRefreshTokenRepo/NewPGTripRepo/NewPGStopRepo/NewPGPhotoRepo`;
   - `s3.NewFromConfig` (endpoint `cfg.S3Endpoint`, static credentials) → `storage.NewS3Storage` (StoragePort);
   - `jwtutil.NewHS256(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)` → issuer/validator;
   - сервисы: `NewAuthService`, `NewTripService`, `NewStopService`, `NewPhotoService`;
   - хендлеры: `NewAuthHandler`, `NewTripHandler`, `NewStopHandler`, `NewPhotoHandler`, `NewExportHandler`.
   - TODO(BEL-03..06) отсутствуют, символ `authMiddleware` отсутствует.
2. **Auth-группа**: `r.With(func(next http.Handler) http.Handler { return middleware.Auth(cfg.JWTSecret, next) }).Group(...)` — эквивалент контрактовому `r.With(middleware.Auth(cfg.JWTSecret))` (chi `With` принимает http.HandlerFunc-обёртку; `middleware.Auth(secret, next)` — сигнатура из A7-05). Внутри группы: `tripHandler.Mount`, `stopHandler.Mount`, `photoHandler.Mount`, `exportHandler.Mount`.
3. **Маршруты**: `authHandler.Mount` — POST /auth/register, /auth/login, /auth/refresh, /auth/logout, GET /users/me (публичные). Trip: GET/POST /trips, GET/PUT/DELETE /trips/{id}. Stop: GET/POST /trips/{id}/stops, GET/PUT/DELETE /trips/{id}/stops/{stopId}. Photo: GET/POST /trips/{id}/photos, DELETE /trips/{id}/photos/{photoId}, POST /trips/{id}/photos/presign. Export: GET /trips/{id}/export (под auth). Соответствие контрактам в handlers.go — полное.
4. **ExportHandler** (`internal/api/handlers/export.go`): CSV header `order,title,description,lat,lng,visit_date`, разделитель `,`, `Content-Disposition: attachment; filename="<slug>.csv"` (slug из title, пусто → trip.csv), visit_date RFC3339, координаты 6 знаков, CSV-экранирование (`,` `\n` `\r` `"`), 404/500 через `middleware.WriteError`.

## Unit-тесты

`internal/api/handlers/export_test.go` — 6 тестов, все **Passed**:
- `TestExport_Success` — 200, text/csv, `attachment; filename="trip-title.csv"`, полный CSV-тело (header + 2 строки, RFC3339, 6 знаков); пробел в description → поле в кавычках (поведение csvQuote, зафиксировано в тесте);
- `TestExport_EmptyTitle_DefaultFilename` — пустой title → `trip.csv`, пустой список остановок → только header;
- `TestExport_NotFound` — чужая/нет поездка → 404;
- `TestExport_StopListError_500` — ошибка списка остановок → 500;
- `TestExport_CsvQuoting` — экранирование `,` `"` `\n` (удвоение кавычек, оборачивание поля);
- `TestExport_RouteRegistered` — GET /trips/{id}/export зарегистрирован: DELETE → 405, другой путь → 404.

Мок TripService — переиспользует `mockTripService` из trip_handler_test.go (один пакет), мок StopService — локальный `mockStopSvc` с assert-компиляцией `var _ service.StopService`. Без сети и БД.

## Покрытие (go test -cover)

- internal/api/handlers — 55.2% (handlers +export-тесты)
- остальные пакеты: middleware 69.8%, config 96.0%, domain 100.0%, jwtutil 89.7%, service 46.5%, repo 3.2% — ok

## Runtime

- `go run ./cmd/server` без env: `config: missing required env: [DATABASE_URL JWT_SECRET]` (exit 1) — ожидаемо (контракт config.Load).
- `go run ./cmd/server` с DATABASE_URL → 127.0.0.1:5432: падение на `migrate: dial tcp 127.0.0.1:5432: connect: connection refused` (exit 1) — **не дефект кода**: в песочнице нет Postgres и Docker (`docker: not found`; установка postgres через apt невозможна — нет прав root), миграции запускаются до старта HTTP-сервера. Поведение в CI покрывается `make e2e` (compose up → healthcheck → e2e → down).

## Локальный запуск

- `.env` (по образцу `.env.example`): POSTGRES_USER/PASSWORD/DB, JWT_SECRET, S3_ENDPOINT/BUCKET/keys, PORT=8080. DATABASE_URL формируется из этих значений (postgres://mytrip:mytrip@localhost:5432/mytrip) — файл используется при `docker compose up` и локальном `make run`; сам файл в git не коммитится (.gitignore).

## Дефекты по контракту

Не найдено. Расхождений между реализацией и контрактами (handlers.go, service/interfaces.go, repo/interfaces.go) нет.

## Итог

Все применимые в песочнице проверки зелёные (сборка, тесты, gofmt, go vet). golangci-lint, runtime с живой БД и e2e — degradable окружения, покрываются CI (`make e2e`). Задача переведена в статус **testing**.
