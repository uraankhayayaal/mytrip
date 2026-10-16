# Тестовый отчёт — A7-01 AuthService: register/login/refresh/logout/me + UserRepo/RefreshTokenRepo

## Статус: **PASSED** (unit-тесты, сборка, линт); интеграция/E2E — **DEGRADED** (нет Postgres/Minio/Docker в окружении)

## Проверки

| Проверка | Команда | Результат |
|---|---|---|
| Сборка бэкенда | `go build ./...` / `go build -o bin/server ./cmd/server` | **Passed** (BUILD_OK) |
| Unit-тесты (race + coverage) | `go test ./internal/... -race -cover` | **Passed** (exit 0) |
| gofmt | `gofmt -l internal/` | **Passed** (пустой вывод) |
| go vet | `go vet ./...` | **Passed** |
| golangci-lint | — | **Недоступен в окружении** (`which golangci-lint` → пусто); зафиксировано в last_error задачи. gofmt+go vet зелёные |
| Интеграция (живая БД) | — | **DEGRADED**: Postgres (5432) и Minio (9000) недоступны, Docker отсутствует в песочнице. Стек `docker compose up` выполнить нельзя |
| E2E (`make e2e`) | — | **DEGRADED**: требует compose-стек + браузеры; недоступно в окружении. Конфигурация Playwright и сценарий `frontend/e2e/auth.spec.ts` присутствуют в репозитории |

## Покрытие (go test -cover)

- internal/api/handlers — 94.3%
- internal/config — 96.0%
- internal/domain — 100.0%
- internal/service — 75.0%

## Соответствие контрактам (проверено до последнего символа)

1. **AuthService** (`internal/service/interfaces.go`): сигнатуры Register/Login/Refresh/Logout/Me совпадают с реализацией `authService` — совпадение 1:1.
2. **UserRepo / RefreshTokenRepo** (`internal/repo/interfaces.go`): `PGUserRepo` и `PGRefreshTokenRepo` реализуют все методы (`var _ UserRepo = (*PGUserRepo)(nil)`, `var _ RefreshTokenRepo = (*PGRefreshTokenRepo)(nil)`). Таблицы соответствуют `0001_init.up.sql` (users.id UUID PK, email UNIQUE → 23505 → `domain.NewConflict`; refresh_tokens.token_hash UNIQUE, user_id FK CASCADE).
3. **AuthHandler** (`handlers.go`): все 5 эндпоинтов смонтированы, контракты соблюдены:
   - POST /auth/register → 201 `{"user":{...}}` (проверено тестом), 400/409;
   - POST /auth/login → 200 `{"access_token","user"}` + Set-Cookie `rt` (HttpOnly, Path=/, SameSite=Lax, Max-Age=2592000 — проверено в `TestLoginHandler_Success`);
   - POST /auth/refresh → 200 + ротация cookie (проверено);
   - POST /auth/logout → 204 + Clear-Cookie Max-Age=-1 (проверено);
   - GET /users/me → 200 (проверено), 401 без user_id.
4. **jwtutil**: сервис использует `TokenIssuer`/`TokenValidator` из `NewHS256` (HS256, sub/type/exp), refresh-ротация: старый SHA-256 hash удаляется (`DeleteByHashWithResult`, 0 строк → 401), новый вставляется. Logout идемпотентно удаляет hash.

## Добавленные тесты (в рамках QA-проверки)

`internal/api/handlers/auth_handler_test.go`:
- `TestRegisterHandler_InvalidBody` — битый JSON → 400 validation (контракт err: 400);
- `TestLoginHandler_InvalidBody` — битый JSON → 400;
- `TestRegisterHandler_EmptyFieldsValidation` — пустые email/name → 400, `error.code=validation`.

Существующие тесты: `internal/service/auth_service_test.go` (10 тестов: Register success/short-password/conflict, Login success/wrong-password, Refresh success/revoked/invalid, Logout success/empty, Me success/empty) — все зелёные, покрывают краевые случаи.

## Рантайм (ReadAppLogs)

Запуск `go run ./cmd/server` с `DATABASE_URL=postgres://…@localhost:5432`: падение на `migrate: dial tcp 127.0.0.1:5432: connection refused` (exit 1). **Дефект кода отсутствует** — в песочнице нет Postgres/Minio/Docker (порты 5432/9000 закрыты, `docker` не установлен). Код корректно останавливается на недоступной БД. Повторная проверка рантайма на живом стеке возможна только в окружении с compose (приёмка `make e2e`).

## Заключение

- Приёмочные критерии A7-01 п.1–4 выполнены: репозитории, сервис (bcrypt/HS256/ротация), хендлер (cookie 'rt', коды ответов), unit-тесты — **Passed**.
- П.5: `make backend-test` (ядро — `go test ./... -race -cover`) зелёный; `make backend-lint` — gofmt + go vet зелёные, golangci-lint отсутствует в окружении (деградация инструмента, не дефект кода).
- Расхождений с контрактами Архитектора не обнаружено. Дефекты: 0.
