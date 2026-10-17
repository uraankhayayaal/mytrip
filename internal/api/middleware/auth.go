package middleware

// Auth middleware (CHAT-07: A7-05) — валидация JWT access-токена.

import (
	"context"
	"net/http"
	"strings"
	"time"

	"mytrip/internal/domain"
	"mytrip/internal/jwtutil"
)

// ContextKey — тип-ключ для user_id в context.
type ContextKey string

// UserIDKey — ключ user_id в context (ставит Auth, читает AuthContext).
const UserIDKey ContextKey = "user_id"

// AuthContext — получает user_id, установленный Auth.
func AuthContext(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

// Auth — middleware валидации Bearer JWT access-токена (CHAT-07: A7-05).
//
// Поведение:
//   - нет header / не начинается с "Bearer " → 401 WriteError(domain.NewUnauthorized).
//   - токен невалиден / expired / refresh вместо access → 401 WriteError(domain.NewUnauthorized).
//   - успех: context.WithValue(UserIDKey, sub), прокидывает дальше.
func Auth(secret string, next http.Handler) http.Handler {
	validator, err := newValidator(secret)
	if err != nil {
		// Ошибка конфигурации: всегда 401, чтобы не сломать сервер.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			WriteError(w, domain.NewUnauthorized("auth middleware is not configured"))
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := extractBearer(r.Header.Get("Authorization"))
		if !ok {
			WriteError(w, domain.NewUnauthorized("missing or malformed Authorization header"))
			return
		}

		claims, err := validator.Validate(token, jwtutil.TokenKindAccess)
		if err != nil {
			WriteError(w, domain.NewUnauthorized("invalid or expired token"))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// defaultAccessTTL / defaultRefreshTTL — фиксированные TTL для валидатора внутри
// middleware. Валидатор не использует TTL при проверке (только подпись/exp/kind),
// но NewHS256 требует положительные значения.
const (
	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 30 * 24 * time.Hour
)

// newValidator — создаёт TokenValidator по секрету.
func newValidator(secret string) (jwtutil.TokenValidator, error) {
	_, validator, err := jwtutil.NewHS256(secret, defaultAccessTTL, defaultRefreshTTL)
	if err != nil {
		return nil, err
	}
	return validator, nil
}

// extractBearer — парсит "Bearer <token>" из заголовка Authorization.
// Возвращает token и ok=false при пустом/невалидном заголовке.
func extractBearer(header string) (string, bool) {
	if header == "" {
		return "", false
	}
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}
