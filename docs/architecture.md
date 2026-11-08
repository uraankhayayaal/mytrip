# Архитектура MyTrip

## Обзор

MyTrip — монолитное веб-приложение для организации путешествий. Пользователи регистрируются, создают поездки (trips), добавляют точки маршрута (stops), прикрепляют фото и экспортируют маршруты.

## Высокоуровневая диаграмма

```
┌──────────────────────────────────────────────────────────┐
│                    Client (Browser)                      │
│  React SPA (Vite + TS + Tailwind)                       │
│  ┌─────────┐ ┌──────────┐ ┌────────┐ ┌───────────────┐  │
│  │  Pages   │ │ Components│ │ API    │ │  Query Client │  │
│  │  Routing │ │ Leaflet  │ │ axios  │ │  React Query  │  │
│  └─────────┘ └──────────┘ └────────┘ └───────────────┘  │
└──────────────────────────────────────────────────────────┘
                            │ HTTP (JSON)
                            ▼
┌──────────────────────────────────────────────────────────┐
│              Nginx (reverse proxy)                       │
│  - static files (frontend/dist/)                         │
│  - /api → upstream server:8080                           │
└──────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────┐
│         Go Server (chi router, :8080)                    │
│                                                          │
│  ┌───────────────────────────────────────────────────┐   │
│  │  Layer: HTTP Handlers (internal/api/handlers)     │   │
│  │  - AuthHandler  TripHandler  StopHandler  Photo   │   │
│  └───────────────────────┬───────────────────────────┘   │
│                          │                                │
│  ┌───────────────────────────────────────────────────┐   │
│  │  Layer: Services (internal/service)               │   │
│  │  - AuthService  TripService  StopService  Photo   │   │
│  └───────────────────────┬───────────────────────────┘   │
│                          │                                │
│  ┌───────────────────────────────────────────────────┐   │
│  │  Layer: Repositories (internal/repo)              │   │
│  │  - PGUserRepo  PGTripRepo  PGStopRepo  PGPhoto   │   │
│  └───────────────────────┬───────────────────────────┘   │
│                          │                                │
│  ┌──────────┐  ┌──────────────────┐                      │
│  │ pgxpool  │  │ S3 client        │                      │
│  │ (Postgres│  │ (aws-sdk-go-v2)  │                      │
│  └──────────┘  └──────────────────┘                      │
└──────────────────────────────────────────────────────────┘
         │                        │
         ▼                        ▼
┌──────────────┐        ┌────────────────┐
│ PostgreSQL 16 │        │ MinIO (S3 API) │
│              │        │  (фото, карты)  │
└──────────────┘        └────────────────┘
```

## Архитектурный стиль: Layered (Clean-ish)

Проект следует **layered architecture** — каждый слой знает только о слое ниже себя:

```
HTTP Handlers → Services → Repositories → Database / External APIs
```

### Слои

| Слой | Путь | Ответственность |
|---|---|---|
| **Domain** | `internal/domain/` | Модели (struct), ошибки (AppError) — «ядро» |
| **Repo** | `internal/repo/` | Интерфейсы репозиториев + PostgreSQL-реализации (pgx) |
| **Service** | `internal/service/` | Бизнес-логика: валидация, транзакции, координация репозиториев |
| **Handlers** | `internal/api/handlers/` | HTTP-маршруты, парсинг запросов, JSON-ответы |
| **Middleware** | `internal/api/middleware/` | Auth, error handling, context enrichment |
| **JWT** | `internal/jwtutil/` | Issuer/Validator для JWT HS256 |
| **Config** | `internal/config/` | Загрузка конфигурации из env |
| **Server** | `cmd/server/main.go` | DI, маршрутизация, запуск HTTP-сервера |

### Правила слоёв

1. **Domain не зависит от других слоёв** — только стандартная библиотека Go.
2. **Repositories реализуют интерфейсы, но не импортируют Services.**
3. **Services знают Domain и Repo-интерфейсы, но не HTTP.**
4. **Handlers знают Services, но не Repositories напрямую.**
5. **`cmd/server/main.go` — единственная точка DI (wire-up).**

## Dependency Injection

Весь граф зависимостей собирается в `cmd/server/main.go`:

```go
// 1. Конфиг
 cfg, err := config.Load()

// 2. БД
 pool, err := pgxpool.New(ctx, cfg.DatabaseURL)

// 3. Миграции
 runMigrations(cfg.DatabaseURL)

// 4. Auth
 userRepo := repo.NewPGUserRepo(pool)
 tokenRepo := repo.NewPGRefreshTokenRepo(pool)
 issuer, validator := jwtutil.NewHS256(cfg.JWTSecret, ...)
 authService := service.NewAuthService(userRepo, tokenRepo, ...)

// 5. Trips
 tripRepo := repo.NewPGTripRepo(pool)
 tripService := service.NewTripService(tripRepo)
 tripHandler := handlers.NewTripHandler(tripService)

// 6. Маршрутизация
 r := chi.NewRouter()
 handlers.NewAuthHandler(authService).Mount(r)
 r.Use(authMiddleware)
 tripHandler.Mount(r)
```

## Потоки данных

### Регистрация пользователя

```
POST /api/v1/auth/register
  → AuthHandler.register()
    → AuthService.Register()
      → UserRepo.Insert() → PostgreSQL
    → TokenIssuer.IssueAccess() → JWT access token
    → TokenIssuer.IssueRefresh() → JWT refresh token
      → RefreshTokenRepo.Insert(hash) → PostgreSQL
    → Set-Cookie: rt (httpOnly, secure, sameSite=strict)
    → 201 {user, access_token}
```

### Авторизованный запрос к поездкам

```
GET /api/v1/trips  Authorization: Bearer <access_token>
  → authMiddleware
    → TokenValidator.Validate(token, "access")
    → context.WithValue(ctx, UserIDKey, claims.Subject)
  → TripHandler.list()
    → TripService.List(ctx, userID)
      → TripRepo.ListByUser(userID) → PostgreSQL
  → 200 [{trip}, ...]
```

### Refresh токена

```
POST /api/v1/auth/refresh  Cookie: rt=<refresh_token>
  → AuthHandler.refresh()
    → AuthService.Refresh(refreshToken)
      → SHA-256(token) → RefreshTokenRepo.DeleteByHash()
      → TokenIssuer.IssueAccess() → новый access token
      → TokenIssuer.IssueRefresh() → новый refresh token
      → RefreshTokenRepo.Insert(newHash)
    → Set-Cookie: rt (new token)
    → 200 {access_token}
```

## Ключевые решения

| Решение | Обоснование |
|---|---|
| **chi/v5** вместо gin/echo | Минимальные зависимости, чистая stdlib-совместимость, простой middleware |
| **pgx/v5** вместо database/sql напрямую | Нативная PostgreSQL оптимизация, scanning, connection pool |
| **golang-migrate** | Версионированные SQL-миграции с откатом (up/down) |
| **JWT HS256** | Симвметричное шифрование, подходит для монолита без отдельных auth-сервисов |
| **Refresh cookies (httpOnly)** | CSRF-безопасное хранение refresh-токенов; access-token в header |
| **aws-sdk-go-v2 S3** | S3-совместимый API для хранения фото; MinIO в dev, S3 в prod |
| **Vite + React Query** | Быстрый HMR, кэширование и автоматический refetch на клиенте |
| **Leaflet** | Лёгкая карта без API-ключа (OpenStreetMap), drag-маркеры для точек маршрута |
