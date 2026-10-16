package jwtutil

// JWT-утилиты — единая точка выдачи и валидации токенов.

import "time"

// TokenKind — тип токена (claim 'type').
type TokenKind string

const (
	TokenKindAccess  TokenKind = "access"
	TokenKindRefresh TokenKind = "refresh"
)

// Claims — единые claims приложения.
type Claims struct {
	Subject string    `json:"sub"`
	Type    TokenKind `json:"type"`
	Expires int64     `json:"exp"`
}

// TokenIssuer — выдача токенов.
type TokenIssuer interface {
	IssueAccess(userID string) (token string, err error)
	IssueRefresh(userID string) (token string, err error)
}

// TokenValidator — валидация токенов.
type TokenValidator interface {
	Validate(token string, expectedKind TokenKind) (*Claims, error)
}

// NewHS256 — конструктор по секрету и TTL (реализация — адаптер golang-jwt/jwt/v5).
// Реализация — jwt.go (задача A7-01/A7-05).
func NewHS256(secret string, accessTTL, refreshTTL time.Duration) (TokenIssuer, TokenValidator, error) {
	p, err := newHS256(secret, accessTTL, refreshTTL)
	if err != nil {
		return nil, nil, err
	}
	return p, p, nil
}
