package jwtutil

// JWT-утилиты (CHAT-07: A7-05) — единая точка выдачи и валидации токенов.
// Используется AuthHandler (issue) и middleware.Auth (validate).
// Контракт зафиксирован ниже; реализация — по задаче A7-05.

import "time"

// TokenKind — тип токена (claim 'type').
type TokenKind string

const (
	TokenKindAccess  TokenKind = "access"
	TokenKindRefresh TokenKind = "refresh"
)

// Claims — единые claims приложения.
// sub  — user_id (UUID как string).
// type — TokenKindAccess | TokenKindRefresh.
// exp  — стандартный (секунды Unix).
type Claims struct {
	Subject string    `json:"sub"`
	Type    TokenKind `json:"type"`
	Expires int64     `json:"exp"`
}

// TokenIssuer — выдача токенов (использует secret + TTL из config).
type TokenIssuer interface {
	// IssueAccess — access-токен для user_id, живёт cfg.JWTAccessTTL.
	IssueAccess(userID string) (token string, err error)
	// IssueRefresh — refresh-токен для user_id, живёт cfg.JWTRefreshTTL.
	IssueRefresh(userID string) (token string, err error)
}

// TokenValidator — валидация токенов.
type TokenValidator interface {
	// Validate — проверяет подпись (HS256, secret), exp и (опционально) kind.
	// expectedKind=="" — не проверять тип. Ошибка → *domain.AppError (unauthorized).
	Validate(token string, expectedKind TokenKind) (*Claims, error)
}

// NewHS256 — конструктор по секрету и TTL (реализация — адаптер golang-jwt/jwt/v5).
// TODO(A7-05): реализовать; secret пустой → ошибка на конструировании.
func NewHS256(secret string, accessTTL, refreshTTL time.Duration) (TokenIssuer, TokenValidator, error) {
	return nil, nil, nil // TODO(A7-05)
}
