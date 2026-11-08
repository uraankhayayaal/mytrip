# Безопасность

## JWT-аутентификация

### Двойная схема токенов

MyTrip использует две пары токенов:

| Токен | Хранение | TTL | Назначение |
|---|---|---|---|
| **Access token** | Client memory / localStorage | 15 минут (конфигурируемо) | Авторизация запросов к API |
| **Refresh token** | HttpOnly cookie (`rt`) | 30 дней (конфигурируемо) | Получение нового access-токена |

### Структура claims

```json
{
  "sub": "<user-uuid>",
  "type": "access" | "refresh",
  "exp": 1704067200,
  "iat": 1704067200
}
```

Алгоритм подписи: **HS256** (HMAC-SHA256 с общим секретом `JWT_SECRET`).

### Flow: авторизованный запрос

```
Client ──GET /api/v1/trips──▶ Server
            Authorization: Bearer <access>
Server: validate JWT HS256
  ✓ → extract sub → context.WithValue(UserIDKey) → handler
  ✗ → 401 unauthorized
```

### Flow: refresh токена

```
1. Клиент получает 401 от API (access-токен истёк)
2. Axios interceptor ловит 401 → POST /api/v1/auth/refresh
3. Сервер читает refresh-токен из Cookie: rt
4. SHA-256(rt) → поиск в refresh_tokens (должен быть в БД)
5. Удаление старого хэша, выдача новых токенов
6. Insert нового хэша в БД
7. Set-Cookie: rt=<new_rt> (ротация)
8. Перезапрос оригинального запроса с новым access-токеном
```

**Ротация (rotation)**: каждый refresh создаёт новую пару токенов. Старый refresh-токен становится невалидным. Это защищает от повторного использования stolen refresh-токена.

## Парольная безопасность

### Хэширование паролей

Пароли хэшируются через **bcrypt** перед сохранением в БД:

```go
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// → password_hash в таблице users
```

При входе:
```go
bcrypt.CompareHashAndPassword(storedHash, []byte(inputPassword))
```

Пароли никогда не передаются в ответе API и не логируются.

## Защита от CSRF

Refresh-токен хранится в **httpOnly cookie** с флагом `SameSite=Strict`:

```
Set-Cookie: rt=<token>; HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth/refresh; Max-Age=<TTL>
```

Это означает:
- Cookie не доступен из JavaScript (защита от XSS-кражи токена).
- Cookie не отправляется при cross-site запросах (защита от CSRF).
- `Secure` — cookie отправляется только по HTTPS (в production).

## Контроль доступа

### Проверка владения ресурсами

Каждый CRUD-запрос к поездкам, точкам и фото проверяет, что ресурс принадлежит пользователю из токена:

```go
// TripRepo.Get — принимает userID, возвращает ошибку если trip.user_id != userID
func (r *PGTripRepo) Get(ctx context.Context, userID, id string) (*domain.Trip, error) {
    row := r.pool.QueryRow(ctx, "SELECT ... FROM trips WHERE id=$1 AND user_id=$2", id, userID)
    // если нет строки → domain.NewNotFound("trip not found")
}
```

**Никогда** не доверяйте ID из запроса клиента без проверки владельца.

### Middleware-цепочка

```
Запрос → ErrorHandler (panic recovery)
    → chi router
        → /api/v1/auth/* → AuthHandler (без auth)
        → /api/v1/* → authMiddleware → TripHandler/StopHandler (с auth)
```

## Валидация входных данных

На уровне сервисов:

```go
// TripService.Create
func (s *TripService) Create(ctx context.Context, userID string, p *TripPayload) error {
    // Валидация title
    if strings.TrimSpace(p.Title) == "" {
        return domain.NewValidation("title is required", nil)
    }
    // Валидация дат
    if p.EndDate.Before(p.StartDate) {
        return domain.NewValidation("end_date must be >= start_date", nil)
    }
}
```

На уровне БД:

```sql
CHECK (end_date >= start_date)          -- trips
CHECK (lat BETWEEN -90 AND 90)          -- trip_stops
CHECK (lng BETWEEN -180 AND 180)        -- trip_stops
```

## Безопасность ответов API

### Error-ответы не раскрывают internals

Внутренние ошибки (500) всегда возвращают универсальное сообщение:

```json
{"error": {"code": "internal", "message": "internal server error"}}
```

Детали ошибки логируются на сервере, но не отправляются клиенту.

### userId скрыт в JSON-ответах

```go
type Trip struct {
    UserID string `json:"-"`  // не сериализуется
}
```

## S3-безопасность

Фотографии загружаются через **presigned URLs**:
- Клиент не имеет прямого доступа к S3.
- Сервер генерирует временный URL на PUT.
- URL истекает через короткий промежуток времени.

## Рекомендации для production

| Настройка | Dev | Production |
|---|---|---|
| `JWT_SECRET` | короткая строка | минимум 32 случайных символа |
| `SameSite` cookie | Strict | Strict |
| `Secure` cookie | optional | mandatory (HTTPS) |
| `JWT_ACCESS_TTL` | 15m | 15m (или короче) |
| Rate limiting | нет | добавить middleware |
| CORS | прокси (нет проблемы) | настроить Allow-Origin |
| HTTPS | нет (localhost) | mandatory (TLS terminator на nginx) |
| `golangci-lint` | опционально | обязательный в CI |
| SQL injection | pgx параметры (безопасно) | параметры + code review |
