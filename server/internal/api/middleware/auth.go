package middleware

// Auth middleware (CHAT-07: A7-05) — валидация JWT access-токена.
// Контракт зафиксирован ниже; реализация — по задаче A7-05.
// Правило: публичные сигнатуры менять НЕЛЬЗЯ.

import (
	"context"
	"net/http"
)

// ContextKey — тип-ключ для user_id в context (избегает коллизий с built-in ключами).
type ContextKey string

const UserIDKey ContextKey = "user_id"

// AuthContext — получает user_id, установленный Auth.
// Возврат "" означает «нет» — хендлер/middleware сам решает, что с этим делать.
func AuthContext(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

// Auth — middleware валидации Bearer JWT access-токена (CHAT-07: A7-05).
//
// Поведение (зафиксированное контрактом):
//  1. Заголовок Authorization отсутствует или не вида "Bearer <token>" → 401
//     {"error":{"code":"unauthorized","message":"missing or malformed Authorization header"}}.
//  2. Подпись HS256 (ключ cfg.JWTSecret) неверна, exp истёк, тип токена != 'access' → 401
//     {"error":{"code":"unauthorized","message":"invalid or expired token"}}.
//  3. Успех: в контекст кладётся sub (user_id), запрос передаётся дальше.
//
// ВАЖНО (корневая причина бага с «валидацией токенов» из эпика CHAT-07):
// refresh-токен (claim 'type'='refresh') ДОЛЖЕН быть отклонён на защищённых
// маршрутах: проверка type=='access' обязательна. В cmd/server/main.go уже
// смонтирован r.With(authMiddleware) — этот символ и есть точка внедрения.
//
// Реализация (TODO A7-05): golang-jwt/jwt/v5 (уже в go.mod? — проверить; иначе добавить),
// claims: {"sub": user_id, "type": "access"|"refresh", "exp": ...}. Парсинг без panics,
// все ошибки → WriteError(w, domain.NewUnauthorized(...)).
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(A7-05): реализовать валидацию по контракту выше.
		middlewareWriteUnauthorized(w)
		next.ServeHTTP(w, r)
	})
}

// middlewareWriteUnauthorized — помощь для заглушки (заменит реализация в A7-05 на WriteError).
func middlewareWriteUnauthorized(w http.ResponseWriter) {
	WriteJSON(w, http.StatusUnauthorized, map[string]any{
		"error": map[string]any{"code": "unauthorized", "message": "invalid or expired token"},
	})
}
