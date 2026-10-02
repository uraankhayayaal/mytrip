package domain

import "fmt"

// Коды ошибок единого error-контракта.
const (
	ErrCodeValidation   = "validation"
	ErrCodeNotFound     = "not_found"
	ErrCodeUnauthorized = "unauthorized"
	ErrCodeConflict     = "conflict"
	ErrCodeInternal     = "internal"
)

// AppError — единый error-контракт приложения.
// JSON-представление: {"error":{"code","message","details?"}}.
type AppError struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	HTTPStatus int            `json:"-"`
}

// Error реализует interface error.
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewValidation — ошибка валидации (400).
func NewValidation(msg string, details map[string]any) *AppError {
	return &AppError{Code: ErrCodeValidation, Message: msg, Details: details, HTTPStatus: 400}
}

// NewNotFound — сущность не найдена (404).
func NewNotFound(msg string) *AppError {
	return &AppError{Code: ErrCodeNotFound, Message: msg, HTTPStatus: 404}
}

// NewUnauthorized — не авторизован (401).
func NewUnauthorized(msg string) *AppError {
	return &AppError{Code: ErrCodeUnauthorized, Message: msg, HTTPStatus: 401}
}

// NewConflict — конфликт (409).
func NewConflict(msg string) *AppError {
	return &AppError{Code: ErrCodeConflict, Message: msg, HTTPStatus: 409}
}

// NewInternal — внутренняя ошибка (500).
func NewInternal() *AppError {
	return &AppError{Code: ErrCodeInternal, Message: "internal server error", HTTPStatus: 500}
}
