package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"mytrip/internal/domain"
)

// errorBody — JSON-обёртка ошибки: {"error":{"code","message","details?"}}.
type errorBody struct {
	Error *domain.AppError `json:"error"`
}

// WriteJSON пишет v в ответ с кодом status и Content-Type: application/json.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError пишет *AppError в ответ с её HTTPStatus.
func WriteError(w http.ResponseWriter, err *domain.AppError) {
	WriteJSON(w, err.HTTPStatus, errorBody{Error: err})
}

// ErrorHandler — middleware: если handler вернул *AppError (через context),
// пишет её в ответ; иначе 500 {"error":{"code":"internal","message":"internal server error"}}.
//
// Примечание: в chi-хендлерах ошибки обычно пишутся напрямую через WriteError
// (см. internal/api/handlers). Этот middleware — страховочный слой: если handler
// ничего не записал в ответ и вернул *AppError через panic-канал — он перехватит её.
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				var appErr *domain.AppError
				if errors.As(rec.(error), &appErr) {
					WriteError(w, appErr)
					return
				}
				WriteError(w, domain.NewInternal())
			}
		}()
		next.ServeHTTP(w, r)
	})
}
