package middleware

// Auth middleware (CHAT-07: A7-05) — валидация JWT access-токена.

import (
	"context"
	"net/http"
)

// ContextKey — тип-ключ для user_id в context.
type ContextKey string

const UserIDKey ContextKey = "user_id"

// AuthContext — получает user_id, установленный Auth.
func AuthContext(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

// Auth — middleware валидации Bearer JWT access-токена (CHAT-07: A7-05).
// TODO(A7-05): реализовать валидацию по контракту.
func Auth(next http.Handler) http.Handler {
	return next
}
