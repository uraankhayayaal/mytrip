# Бэкенд

## Обзор

Бэкенд MyTrip — Go-сервер на `chi/v5` маршрутизаторе, слушающий на порту `8080`. Обслуживает REST API для аутентификации, CRUD поездок, точек маршрута и фотографий.

## Структура проекта

```
cmd/server/main.go          — точка входа, DI, маршрутизация
internal/
  config/config.go          — загрузка конфигурации из env
  domain/
    models.go               — domain-модели (User, Trip, Stop, Photo)
    errors.go               — единый AppError-контракт
  jwtutil/jwt.go            — JWT issuer/validator (HS256)
  repo/
    interfaces.go           — интерфейсы репозиториев
    postgres.go             — UserRepo + RefreshTokenRepo (pgx)
    trip_repo.go            — TripRepo (pgx)
    stop_repo.go            — StopRepo (pgx)
  service/
    interfaces.go           — интерфейсы сервисов + payloads
    auth_service.go         — регистрация, логин, refresh, logout
    trip_service.go         — CRUD поездок с валидацией
    stop_service.go         — CRUD точек маршрута
  api/
    middleware/
      auth.go               — context key, AuthContext()
      error_handler.go      — panic recovery, WriteJSON, WriteError
    handlers/
      handlers.go           — JSON-структуры ответов
      auth_handler.go       — POST /auth/register, /auth/login, /auth/refresh, /auth/logout, GET /users/me
      trip_handler.go       — GET/POST /trips, GET/PUT/DELETE /trips/{id}
      stop_handler.go       — GET/POST /trips/{id}/stops, GET/PUT/DELETE /trips/{id}/stops/{stopId}
```

## Запуск

```bash
# Установить зависимости
make deps

# Собрать
make backend-build

# Запустить (dev)
make backend-run

# Тесты
make backend-test

# Линт
make backend-lint
```

### Требуемые переменные окружения

| Переменная | Обязательно | Описание | Пример |
|---|---|---|---|
| `DATABASE_URL` | ✅ | Подключение к PostgreSQL | `postgres://user:pass@localhost:5432/mytrip?sslmode=disable` |
| `JWT_SECRET` | ✅ | Секрет для подписи JWT | произвольная строка |
| `PORT` | ❌ | Порт сервера (по умолч. `8080`) | `8080` |
| `JWT_ACCESS_TTL` | ❌ | Живучесть access-токена (по умолч. `15m`) | `15m` |
| `JWT_REFRESH_TTL` | ❌ | Живучесть refresh-токена (по умолч. `30d`) | `720h` |
| `S3_ENDPOINT` | ✅ | Endpoint S3-совместимого хранилища | `http://localhost:9000` |
| `S3_BUCKET` | ❌ | Имя бакета (по умолч. `mytrip`) | `mytrip` |
| `S3_ACCESS_KEY` | ✅ | Access key S3 | `minioadmin` |
| `S3_SECRET_KEY` | ✅ | Secret key S3 | `minioadmin` |

## Конфигурация (`config.Config`)

Загрузка из переменных окружения через `config.Load()`. Если `DATABASE_URL` или `JWT_SECRET` не заданы — паника на старте.

```go
type Config struct {
    Port          string
    DatabaseURL   string
    JWTSecret     string
    JWTAccessTTL  time.Duration  // default 15m
    JWTRefreshTTL time.Duration  // default 30d
    S3Endpoint    string
    S3Bucket      string         // default "mytrip"
    S3AccessKey   string
    S3SecretKey   string
}
```

## Middleware

### ErrorHandler (глобальный)

Первый middleware на роутере. Перехватывает panic и преобразует их в JSON-ответы:

```go
r.Use(middleware.ErrorHandler)
```

- Если panic — `*domain.AppError` → пишет соответствующий JSON с HTTP-кодом.
- Если panic — любой другой `error` → 500 `internal server error`.

### AuthMiddleware (для защищённых маршрутов)

Валидирует `Authorization: Bearer <token>` заголовок:

1. Проверяет наличие префикса `Bearer `.
2. Валидирует JWT через `deps.validator.Validate(token, TokenKindAccess)`.
3. Извлекает `sub` (user_id) и кладёт в context.
4. При ошибке — 401 `unauthorized`.

Применяется ко всем маршрутам под `/api/v1` за исключением `/auth/*`:

```go
r.Route("/api/v1", func(r chi.Router) {
    handlers.NewAuthHandler(deps.auth).Mount(r)  // без auth
    r.Use(authMiddleware)                        // с auth
    handlers.NewTripHandler(trips.trips).Mount(r)
})
```

## Domain Models

### User
```go
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Trip
```go
type Trip struct {
    ID            string    `json:"id"`
    UserID        string    `json:"-"`           // скрыт в JSON
    Title         string    `json:"title"`
    Description   string    `json:"description"`
    StartDate     time.Time `json:"start_date"`
    EndDate       time.Time `json:"end_date"`
    CoverPhotoURL *string   `json:"cover_photo_url,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

### Stop
```go
type Stop struct {
    ID          string    `json:"id"`
    TripID      string    `json:"trip_id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Lat         float64   `json:"lat"`
    Lng         float64   `json:"lng"`
    VisitDate   time.Time `json:"visit_date"`
    Order       int       `json:"order"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### Photo
```go
type Photo struct {
    ID        string    `json:"id"`
    TripID    string    `json:"trip_id"`
    URL       string    `json:"url"`
    CreatedAt time.Time `json:"created_at"`
}
```

## AppError — единый контракт ошибок

```go
type AppError struct {
    Code       string         `json:"code"`
    Message    string         `json:"message"`
    Details    map[string]any `json:"details,omitempty"`
    HTTPStatus int            `json:"-"`
}
```

| Код | HTTP | Фабрика |
|---|---|---|
| `validation` | 400 | `domain.NewValidation(msg, details)` |
| `not_found` | 404 | `domain.NewNotFound(msg)` |
| `unauthorized` | 401 | `domain.NewUnauthorized(msg)` |
| `conflict` | 409 | `domain.NewConflict(msg)` |
| `internal` | 500 | `domain.NewInternal()` |

JSON-ответ на ошибку:
```json
{
  "error": {
    "code": "not_found",
    "message": "trip not found",
    "details": {"id": "abc123"}
  }
}
```

## JWT (jwtutil)

### Схема токенов

```json
{
  "sub": "<user-uuid>",
  "type": "access" | "refresh",
  "exp": 1234567890,
  "iat": 1234567890
}
```

- **Access token**: JWT, передаётся в `Authorization: Bearer` header, TTL по умолчанию 15 минут.
- **Refresh token**: JWT, хранится в httpOnly cookie `rt`, TTL по умолчанию 30 дней.

### Ротация refresh-токенов

При каждом refresh:
1. SHA-256(refreshToken) → поиск в `refresh_tokens` таблице.
2. Старый hash удаляется (`RefreshTokenRepo.DeleteByHash`).
3. Выдаётся новый access + refresh token.
4. Hash нового refresh токена вставляется в БД.

## Graceful shutdown

Сервер слушает `SIGINT` и `SIGTERM`:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
// ...
go func() {
    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = srv.Shutdown(shutdownCtx)
}()
```

## Health check

```http
GET /healthz
HTTP/1.1 200 OK
{"status": "ok"}
```

Без авторизации. Используется для healthcheck Docker Compose.

## Миграции

При запуске сервер автоматически применяет миграции из `db/migrations/` через `golang-migrate`:

```go
func runMigrations(databaseURL string) error {
    source, _ := (&file.File{}).Open("file://db/migrations")
    db, _ := (&postgres.Postgres{}).Open(databaseURL)
    m, _ := migrate.NewWithInstance("file", source, "postgres", db)
    return m.Up() // применяет все неприменённые миграции
}
```

## Тестирование

```
internal/service/*_test.go     — unit-тесты сервисов с моками репозиториев
internal/api/handlers/*_test.go — unit-тесты хендлеров с httptest + моками сервисов
internal/repo/*_test.go         — интеграционные тесты репозиториев
internal/domain/errors_test.go   — тесты контракта ошибок
internal/config/config_test.go   — тесты загрузки конфигурации
```

Запуск: `make backend-test` → `go test ./... -race -cover`
