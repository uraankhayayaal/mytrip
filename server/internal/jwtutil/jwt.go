package jwtutil

// JWT-утилиты (CHAT-07: A7-05) — единая точка выдачи и валидации токенов.
// Используется AuthHandler (issue) и middleware.Auth (validate).
// Контракт зафиксирован ниже; реализация — по задаче A7-05.

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"mytrip/internal/domain"
)

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

type hs256 struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewHS256 — конструктор по секрету и TTL (реализация — адаптер golang-jwt/jwt/v5).
// secret пустой → ошибка на конструировании.
func NewHS256(secret string, accessTTL, refreshTTL time.Duration) (TokenIssuer, TokenValidator, error) {
	if secret == "" {
		return nil, nil, errors.New("jwtutil: secret is empty")
	}
	if accessTTL <= 0 {
		return nil, nil, errors.New("jwtutil: accessTTL must be positive")
	}
	if refreshTTL <= 0 {
		return nil, nil, errors.New("jwtutil: refreshTTL must be positive")
	}
	return &hs256{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}, nil, nil
}

func (h *hs256) sign(userID string, kind TokenKind, ttl time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"type": string(kind),
		"exp":  claims.ExpiresAt.Unix(),
		"iat":  claims.IssuedAt.Unix(),
	})
	return token.SignedString(h.secret)
}

// IssueAccess — access-токен для user_id, живёт cfg.JWTAccessTTL.
func (h *hs256) IssueAccess(userID string) (string, error) {
	return h.sign(userID, TokenKindAccess, h.accessTTL)
}

// IssueRefresh — refresh-токен для user_id, живёт cfg.JWTRefreshTTL.
func (h *hs256) IssueRefresh(userID string) (string, error) {
	return h.sign(userID, TokenKindRefresh, h.refreshTTL)
}

// Validate — проверяет подпись (HS256, secret), exp и (опционально) kind.
func (h *hs256) Validate(token string, expectedKind TokenKind) (*Claims, error) {
	claims, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return h.secret, nil
	})
	if err != nil || !claims.Valid {
		return nil, domain.NewUnauthorized("invalid or expired token")
	}
	m, ok := claims.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.NewUnauthorized("invalid or expired token")
	}
	sub, _ := m["sub"].(string)
	typ, _ := m["type"].(string)
	exp, _ := m["exp"].(float64)
	c := &Claims{Subject: sub, Type: TokenKind(typ), Expires: int64(exp)}
	if c.Subject == "" {
		return nil, domain.NewUnauthorized("invalid or expired token")
	}
	if expectedKind != "" && c.Type != expectedKind {
		return nil, domain.NewUnauthorized("invalid or expired token")
	}
	return c, nil
}
