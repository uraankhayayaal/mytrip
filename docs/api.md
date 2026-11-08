# REST API

## База

```
Base URL: http://localhost:8080/api/v1
Content-Type: application/json
```

## Авторизация

Большинство эндпоинтов требует авторизации через Bearer JWT:

```
Authorization: Bearer <access_token>
```

Refresh-токен хранится в httpOnly cookie `rt` и используется только для `POST /auth/refresh`.

## Ошибки

Все ошибки возвращаются в едином формате:

```json
{
  "error": {
    "code": "not_found",
    "message": "trip not found",
    "details": {"id": "abc-123"}
  }
}
```

| Код ошибки | HTTP | Описание |
|---|---|---|
| `validation` | 400 | Ошибка валидации запроса |
| `unauthorized` | 401 | Нет авторизации / токен невалиден |
| `not_found` | 404 | Ресурс не найден |
| `conflict` | 409 | Конфликт (например, email уже занят) |
| `internal` | 500 | Внутренняя ошибка сервера |

---

## Health

### GET /healthz

Проверка здоровья сервера.

```http
GET /healthz
```

**Ответ:** `200 OK`
```json
{"status": "ok"}
```

---

## Auth

### POST /auth/register

Регистрация нового пользователя. Возвращает access-токен и устанавливает refresh-токен в cookie.

**Запрос:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "Иван"
}
```

**Ответ:** `201 Created`
```json
{
  "access_token": "eyJhbG...",
  "user": {
    "id": "550e8400-...",
    "email": "user@example.com",
    "name": "Иван",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Ошибки:**
- `400 validation` — пустой email/password/name
- `409 conflict` — email уже занят

### POST /auth/login

Вход в систему.

**Запрос:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Ответ:** `200 OK`
```json
{
  "access_token": "eyJhbG...",
  "user": {
    "id": "550e8400-...",
    "email": "user@example.com",
    "name": "Иван",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

Cookie в ответе: `rt=<refresh_token>; HttpOnly; Secure; SameSite=Strict; Max-Age=<TTL>`

**Ошибки:**
- `400 validation` — пустые поля
- `401 unauthorized` — неверный пароль

### POST /auth/refresh

Получение нового access-токена по refresh-cookie.

**Запрос:** Cookie: `rt=<refresh_token>`

**Ответ:** `200 OK`
```json
{
  "access_token": "eyJhbG..."
}
```

Новый refresh-токен устанавливается в cookie (ротация).

**Ошибки:**
- `401 unauthorized` — невалидный или истёкший refresh-токен

### POST /auth/logout

Выход из системы. Удаляет refresh-токен и очищает cookie.

**Ответ:** `204 No Content`

### GET /users/me

Информация о текущем пользователе (требует авторизации).

**Ответ:** `200 OK`
```json
{
  "id": "550e8400-...",
  "email": "user@example.com",
  "name": "Иван",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Ошибки:**
- `401 unauthorized`

---

## Trips (Поездки)

Все маршруты требуют авторизации (`Authorization: Bearer <token>`).

### GET /trips

Список всех поездок текущего пользователя.

**Ответ:** `200 OK`
```json
[
  {
    "id": "550e8400-...",
    "title": "Поездка в горы",
    "description": "Летний отдых",
    "start_date": "2024-07-01T00:00:00Z",
    "end_date": "2024-07-14T00:00:00Z",
    "cover_photo_url": "https://...",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

Пустой список возвращается как `[]`.

### POST /trips

Создание новой поездки.

**Запрос:**
```json
{
  "title": "Поездка в горы",
  "description": "Летний отдых",
  "start_date": "2024-07-01T00:00:00Z",
  "end_date": "2024-07-14T00:00:00Z"
}
```

**Ответ:** `201 Created` — объект Trip

**Ошибки:**
- `400 validation` — пустой title, end_date < start_date

### GET /trips/{id}

Получение поездки по ID (только своей).

**Ответ:** `200 OK` — объект Trip

**Ошибки:**
- `404 not_found` — поездка не найдена или не принадлежит пользователю

### PUT /trips/{id}

Обновление поездки.

**Запрос:**
```json
{
  "title": "Новое название",
  "description": "Обновлённое описание",
  "start_date": "2024-07-01T00:00:00Z",
  "end_date": "2024-07-14T00:00:00Z"
}
```

**Ответ:** `200 OK` — обновлённый объект Trip

### DELETE /trips/{id}

Удаление поездки (каскадно: удаляются все stops и photos).

**Ответ:** `204 No Content`

---

## Stops (Точки маршрута)

Все маршруты требуют авторизации.

### GET /trips/{tripId}/stops

Список точек маршрута поездки (упорядочены по `order`).

**Ответ:** `200 OK`
```json
[
  {
    "id": "...",
    "trip_id": "...",
    "title": "Горная долина",
    "description": "Красивое место",
    "lat": 45.123,
    "lng": 60.456,
    "visit_date": "2024-07-02T00:00:00Z",
    "order": 1,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### POST /trips/{tripId}/stops

Добавление точки в маршрут.

**Запрос:**
```json
{
  "title": "Горная долина",
  "description": "Красивое место",
  "lat": 45.123,
  "lng": 60.456,
  "visit_date": "2024-07-02T00:00:00Z",
  "order": 1
}
```

**Ответ:** `201 Created` — объект Stop

### GET /trips/{tripId}/stops/{stopId}

Получение точки маршрута.

**Ответ:** `200 OK` — объект Stop

### PUT /trips/{tripId}/stops/{stopId}

Обновление точки маршрута.

**Запрос:** те же поля, что и при создании.

**Ответ:** `200 OK` — обновлённый объект Stop

### DELETE /trips/{tripId}/stops/{stopId}

Удаление точки маршрута.

**Ответ:** `204 No Content`

---

## Photos (Фотографии)

Все маршруты требуют авторизации.

### GET /trips/{tripId}/photos

Список фотографий поездки.

**Ответ:** `200 OK`
```json
[
  {
    "id": "...",
    "trip_id": "...",
    "url": "https://s3-endpoint/bucket/key",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### POST /trips/{tripId}/photos

Загрузка фотографии (через presigned URL S3).

**Ответ:** `201 Created` — объект Photo

### DELETE /trips/{tripId}/photos/{photoId}

Удаление фотографии.

**Ответ:** `204 No Content`

---

## Export (Экспорт)

Все маршруты требуют авторизации.

### GET /trips/{tripId}/export/pdf

Экспорт поездки в PDF-формат.

**Ответ:** `200 OK` (Content-Type: `application/pdf`)

### GET /trips/{tripId}/export/csv

Экспорт точек маршрута в CSV.

**Ответ:** `200 OK` (Content-Type: `text/csv`)
