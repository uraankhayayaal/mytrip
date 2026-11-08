# Развёртывание

## Docker Compose

Проект разворачивается через `docker-compose.yml` — все сервисы поднимаются одной командой:

```bash
make up       # docker compose up -d --build
make down     # docker compose down
make logs     # docker compose logs -f
make ps       # docker compose ps
```

## Сервисы

```
┌─────────────────────────────────────────────────────────────────┐
│                        Интернет / Браузер                       │
│                    http://localhost:3000                        │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  Nginx (web)  —  :3000                                          │
│  - статические файлы frontend/dist/                             │
│  - /api → проксирование на server:8080                          │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTP
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
┌──────────────────┐ ┌──────────────────┐ ┌────────────────┐
│ server (:8080)   │ │ postgres (:5432) │ │ minio (:9000)  │
│ Go HTTP-сервер   │ │ PostgreSQL 16    │ │ MinIO (S3 API) │
│ chi router       │ │                  │ │ :9001 console  │
└──────────────────┘ └──────────────────┘ └────────────────┘
```

### server

Go-сервер. Собирается из `Dockerfile.server` (distroless/static — минимальный образ без shell).

- **Порт**: 8080
- **Зависимости**: postgres, minio (ожидает healthcheck)
- **Конфигурация**: через переменные окружения (`.env` файл + `environment`)

### postgres

PostgreSQL 16 — основная база данных.

- **Порт**: 5432
- **Хранилище**: именованный том `pgdata` (данные сохраняются между перезапусками)
- **Healthcheck**: `pg_isready` (интервал 5s, 10 попыток)

### minio

MinIO — S3-совместимое хранилище для фотографий.

- **Порт API**: 9000
- **Порт консоли**: 9001
- **Хранилище**: именованный том `miniodata`
- **Консоль**: http://localhost:9001

### web (Nginx)

Обратный прокси и раздача статики.

- **Порт**: 3000 → 80 (внутри контейнера)
- **Статика**: `frontend/dist/` → `/usr/share/nginx/html`
- **Конфиг**: `deploy/nginx.conf` → `/etc/nginx/conf.d/default.conf`

## Переменные окружения

Создайте `.env` файл в корне проекта:

```env
# PostgreSQL
POSTGRES_USER=mytrip
POSTGRES_PASSWORD=mytrip
POSTGRES_DB=mytrip

# JWT
JWT_SECRET=<произвольная-строка-минимум-32-символа>

# S3 (MinIO)
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=mytrip
```

Сервер подхватывает переменные автоматически:
- `DATABASE_URL` конструируется из `POSTGRES_*` переменных
- `S3_ENDPOINT` = `http://minio:9000`

## Production-развёртывание

Для production замените MinIO на AWS S3 (или другой S3-совместимый сервис):

```env
S3_ENDPOINT=https://s3.region.amazonaws.com
S3_ACCESS_KEY=<aws-access-key>
S3_SECRET_KEY=<aws-secret-key>
S3_BUCKET=mytrip-production
```

Полный `.env` для production:

```env
POSTGRES_USER=mytrip
POSTGRES_PASSWORD=<сильный-пароль>
POSTGRES_DB=mytrip_prod
JWT_SECRET=<длинный-случайный-секрет>
S3_ENDPOINT=https://s3.eu-west-1.amazonaws.com
S3_ACCESS_KEY=<aws-key>
S3_SECRET_KEY=<aws-secret>
S3_BUCKET=mytrip-prod
```

## Nginx конфигурация

Файл `deploy/nginx.conf`:
- Раздаёт статику из `/usr/share/nginx/html` (frontend/dist)
- Прокидывает `/api` на upstream `server:8080`
- Обрабатывает SPA-роутинг (fallback на `index.html`)

## Сетевая модель

Все контейнеры находятся в bridge-сети `mytrip`. Взаимодействие между сервисами происходит по именам:
- сервер → postgres: `postgres:5432`
- сервер → minio: `minio:9000`

## Полезные команды

```bash
# Интерактивный shell в контейнере сервера
make infra.server-shell

# Подключение к PostgreSQL
make infra.postgres-shell

# Проверка доступности сервера
make infra.up

# Просмотр файлов в MinIO
make infra.minio-mc
```

## E2E-прогон

```bash
make e2e
```

Команда поднимает стек, ждёт готовности (healthcheck до 60s), запускает Playwright-тесты и останавливает стек.
