# Разработка

## Быстрый старт

### Предварительные требования

- **Go 1.24+** — для бэкенда
- **Node.js 18+** — для фронтенда
- **Docker + Docker Compose** — для инфраструктуры (PostgreSQL, MinIO)

### Установка и запуск

```bash
# 1. Клонировать репозиторий
git clone <repo-url> && cd mytrip

# 2. Создать .env
cp .env.example .env   # отредактируйте JWT_SECRET

# 3. Установить зависимости
make deps

# 4. Поднять инфраструктуру
make up

# 5. Запустить dev-серверы (backend + frontend параллельно)
make run
```

После запуска:
- **Фронтенд**: http://localhost:5173
- **Бэкенд**: http://localhost:8080
- **MinIO Console**: http://localhost:9001

## Makefile — справочник команд

Показать все команды: `make help`

### Зависимости и сборка

| Команда | Описание |
|---|---|
| `make deps` | `go mod download` + `npm --prefix frontend ci` |
| `make build` | Собрать backend (→ `bin/server`) + frontend (→ `frontend/dist/`) |
| `make backend-build` | `go build -o bin/server ./cmd/server` |
| `make frontend-build` | `tsc -b && vite build` |

### Запуск

| Команда | Описание |
|---|---|
| `make run` | Параллельный dev-запуск backend + frontend (Ctrl+C останавливает оба) |
| `make backend-run` | `go run ./cmd/server` на :8080 |
| `make frontend-run` | `vite` на :5173 (прокси /api → :8080) |

### Тесты

| Команда | Описание |
|---|---|
| `make test` | Тесты backend + frontend |
| `make backend-test` | `go test ./... -race -cover` |
| `make frontend-test` | `vitest --run` (unit-тесты, jsdom) |
| `make e2e` | Full E2E: `up` → healthcheck → Playwright → `down` |

### Линтинг

| Команда | Описание |
|---|---|
| `make lint` | Линт backend + frontend |
| `make backend-lint` | `gofmt -l .` + `go vet ./...` + `golangci-lint run` |
| `make frontend-lint` | `prettier --check .` + `eslint .` |

### Docker Compose

| Команда | Описание |
|---|---|
| `make up` | `docker compose up -d --build` |
| `make down` | `docker compose down` |
| `make logs` | `docker compose logs -f` |
| `make ps` | `docker compose ps` |
| `make infra.server-shell` | Интерактивный shell в server |
| `make infra.postgres-shell` | `psql` в postgres |
| `make infra.minio-mc` | mc-клиент minio |

## Структура разработки

### Бэкенд

```bash
# Тесты конкретного пакета
go test ./internal/service/... -v
go test ./internal/api/handlers/... -v

# Тест с покрытием
go test ./internal/service/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Race detection
go test ./... -race

# Форматирование
gofmt -w .

# Валидация типов
go vet ./...
```

### Фронтенд

```bash
# Unit-тесты (watch mode)
npm --prefix frontend test

# Unit-тесты (один прогон)
npm --prefix frontend test -- --run

# E2E-тесты
npm --prefix frontend run e2e

# TypeScript-проверка
cd frontend && npx tsc --noEmit
```

## Горячая перезагрузка (HMR)

При dev-запуске:
- **Vite** (фронтенд) — HMR для TS/TSX/CSS изменений.
- **Бэкенд** — перезапуск вручную (или используйте `air` / `entr` для auto-reload).

### Dev-прокси

В dev-режиме Vite проксирует `/api` → `localhost:8080`. Это позволяет обходить CORS и использовать относительные пути:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': { target: 'http://localhost:8080' },
  },
}
```

## Миграции БД

Миграции хранятся в `db/migrations/`. Сервер применяет их автоматически при запуске. Для ручного управления:

```bash
# Ручной запуск миграций (через golang-migrate CLI)
migrate -path db/migrations -database "postgres://..." up
migrate -path db/migrations -database "postgres://..." down
```

### Создание новой миграции

1. Создать пару файлов в `db/migrations/`:
   ```
   0005_<name>.up.sql
   0005_<name>.down.sql
   ```
2. `.up.sql` — SQL для применения миграции.
3. `.down.sql` — SQL для отката миграции.

Пример:
```sql
-- 0005_add_field.up.sql
ALTER TABLE trips ADD COLUMN new_field TEXT NOT NULL DEFAULT '';

-- 0005_add_field.down.sql
ALTER TABLE trips DROP COLUMN new_field;
```

## Отладка

### Логи сервера

```bash
make logs | grep server
```

### Debug-запуск сервера

```bash
dlv debug ./cmd/server
dlv attach <pid>
```

### Проверка БД

```bash
make infra.postgres-shell
-- внутри psql:
SELECT * FROM users;
SELECT * FROM trips;
SELECT count(*) FROM trip_stops;
```

### Проверка S3 (MinIO)

```bash
make infra.minio-mc
# или через консоль: http://localhost:9001
```

### API-тестирование (curl)

```bash
# Health check
curl http://localhost:8080/healthz

# Регистрация
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@test.com","password":"pass123","name":"Test"}'

# Логин
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@test.com","password":"pass123"}'

# Список поездок (с токеном)
curl http://localhost:8080/api/v1/trips \
  -H 'Authorization: Bearer <access_token>'
```

## Код-стайл

### Go

- `gofmt` — стандартное форматирование
- `go vet` — статический анализ
- `golangci-lint` — полный линтинг (в CI)
- Имена пакетов — строчные, без подчёркиваний
- Интерфейсы — в слое, где они используются (не рядом с реализацией)

### TypeScript

- `prettier` — форматирование
- `eslint` + `@typescript-eslint` — линтинг
- Строгие типы: нет `any` без комментария
- Экспорты — named exports (не default)
