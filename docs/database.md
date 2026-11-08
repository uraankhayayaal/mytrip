# База данных

## Обзор

Хранилище данных MyTrip — PostgreSQL 16. Все изменения схемы управляются через SQL-миграции (`golang-migrate`).

## Схема

```
┌──────────────┐         ┌───────────────┐         ┌──────────────┐
│    users     │◄────────┤    trips      │────────►│  trip_stops  │
│──────────────│         │───────────────│         │──────────────│
│ id (UUID PK) │         │ id (UUID PK)  │         │ id (UUID PK) │
│ email (UK)   │         │ user_id (FK)  │         │ trip_id (FK) │
│ name         │         │ title         │         │ title        │
│ password_hash│         │ description   │         │ description  │
│ created_at   │         │ start_date    │         │ lat, lng     │
└──────────────┘         │ end_date      │         │ visit_date   │
        │                │ cover_photo_url│         │ order        │
        │                │ created_at    │         │ created_at   │
        │                │ updated_at    │         └──────────────┘
        │                └───────┬───────┘                   │
        │                        │                           │
        ▼                        ▼                           │
┌──────────────┐          ┌───────────────┐                  │
│refresh_tokens│          │  trip_photos  │◄─────────────────┘
│──────────────│          │───────────────│
│ id (UUID PK) │          │ id (UUID PK)  │
│ user_id (FK) │          │ trip_id (FK)  │
│ token_hash   │          │ url           │
│ expires_at   │          │ created_at    │
│ created_at   │          └───────────────┘
└──────────────┘
```

## Таблицы

### users

Пользователи приложения.

| Столбец | Тип | Ограничения | Описание |
|---|---|---|---|
| `id` | `UUID` | `PRIMARY KEY DEFAULT gen_random_uuid()` | Уникальный идентификатор |
| `email` | `TEXT` | `NOT NULL UNIQUE` | Email (уникальный) |
| `name` | `TEXT` | `NOT NULL` | Имя пользователя |
| `password_hash` | `TEXT` | `NOT NULL` | Хэш пароля (bcrypt) |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | Дата создания |

### refresh_tokens

Refresh-токены для ротации.

| Столбец | Тип | Ограничения | Описание |
|---|---|---|---|
| `id` | `UUID` | `PRIMARY KEY DEFAULT gen_random_uuid()` | Идентификатор |
| `user_id` | `UUID` | `NOT NULL REFERENCES users(id) ON DELETE CASCADE` | Владелец токена |
| `token_hash` | `TEXT` | `NOT NULL UNIQUE` | SHA-256 хэш refresh-токена |
| `expires_at` | `TIMESTAMPTZ` | `NOT NULL` | Срок действия |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | Дата создания |

**Индексы:**
- `idx_refresh_tokens_user` по `(user_id)` — быстрый поиск токенов пользователя.

### trips

Поездки (путешествия).

| Столбец | Тип | Ограничения | Описание |
|---|---|---|---|
| `id` | `UUID` | `PRIMARY KEY DEFAULT gen_random_uuid()` | Идентификатор |
| `user_id` | `UUID` | `NOT NULL REFERENCES users(id) ON DELETE CASCADE` | Владелец поездки |
| `title` | `TEXT` | `NOT NULL` | Название поездки |
| `description` | `TEXT` | `NOT NULL DEFAULT ''` | Описание |
| `start_date` | `TIMESTAMPTZ` | `NOT NULL` | Дата начала |
| `end_date` | `TIMESTAMPTZ` | `NOT NULL` | Дата окончания |
| `cover_photo_url` | `TEXT` | nullable | URL обложки |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | Дата создания |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | Дата обновления |

**Ограничения:**
- `CHECK (end_date >= start_date)` — поездка не может заканчиваться до начала.

**Индексы:**
- `idx_trips_user` по `(user_id)` — быстрый поиск поездок пользователя.

### trip_stops

Точки маршрута внутри поездки.

| Столбец | Тип | Ограничения | Описание |
|---|---|---|---|
| `id` | `UUID` | `PRIMARY KEY DEFAULT gen_random_uuid()` | Идентификатор |
| `trip_id` | `UUID` | `NOT NULL REFERENCES trips(id) ON DELETE CASCADE` | Поездка |
| `title` | `TEXT` | `NOT NULL` | Название точки |
| `description` | `TEXT` | `NOT NULL DEFAULT ''` | Описание |
| `lat` | `DOUBLE PRECISION` | `NOT NULL, CHECK (-90 ≤ lat ≤ 90)` | Широта |
| `lng` | `DOUBLE PRECISION` | `NOT NULL, CHECK (-180 ≤ lng ≤ 180)` | Долгота |
| `visit_date` | `TIMESTAMPTZ` | `NOT NULL` | Дата посещения |
| `order` | `INTEGER` | `NOT NULL` | Порядок в маршруте |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | Дата создания |

**Индексы:**
- `idx_stops_trip` по `(trip_id, order)` — упорядоченный вывод точек.

### trip_photos

Фотографии поездок.

| Столбец | Тип | Ограничения | Описание |
|---|---|---|---|
| `id` | `UUID` | `PRIMARY KEY DEFAULT gen_random_uuid()` | Идентификатор |
| `trip_id` | `UUID` | `NOT NULL REFERENCES trips(id) ON DELETE CASCADE` | Поездка |
| `url` | `TEXT` | `NOT NULL` | URL фотографии в S3 |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | Дата создания |

**Индексы:**
- `idx_photos_trip` по `(trip_id)` — быстрый поиск фото по поездке.

## Миграции

Файлы миграций в `db/migrations/`:

| Файл | Таблицы |
|---|---|
| `0001_init.up.sql` / `.down.sql` | `users`, `refresh_tokens` + extension `pgcrypto` |
| `0002_trips.up.sql` / `.down.sql` | `trips` |
| `0003_stops.up.sql` / `.down.sql` | `trip_stops` |
| `0004_photos.up.sql` / `.down.sql` | `trip_photos` |

Сервер применяет миграции автоматически при запуске:

```go
m, _ := migrate.NewWithInstance("file", source, "postgres", db)
m.Up() // применяет все неприменённые миграции
```

## Внешние ключи и каскадное удаление

Все связи используют `ON DELETE CASCADE`:
- Удаление `user` → удаляются все его `refresh_tokens`, `trips`, `trip_stops`, `trip_photos`.
- Удаление `trip` → удаляются все его `trip_stops` и `trip_photos`.

## Индексы

| Таблица | Индекс | Столбцы | Назначение |
|---|---|---|---|
| `refresh_tokens` | `idx_refresh_tokens_user` | `(user_id)` | Поиск токенов пользователя |
| `trips` | `idx_trips_user` | `(user_id)` | Поиск поездок пользователя |
| `trip_stops` | `idx_stops_trip` | `(trip_id, order)` | Упорядоченный вывод точек |
| `trip_photos` | `idx_photos_trip` | `(trip_id)` | Поиск фото по поездке |
