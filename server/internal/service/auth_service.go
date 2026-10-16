package service

// AuthService — аутентификация (CHAT-07: A7-01).
// Контракт: internal/service/interfaces.go (интерфейс AuthService).
//
// Логика:
//   - Register: bcrypt-хэш пароля, INSERT users; дубль email — domain.NewConflict (из репо).
//   - Login: bcrypt CompareHashAndPassword; выдача access + refresh JWT через jwtutil.
//     Новый refresh-токен (SHA-256 hash) сохраняется в refresh_tokens (ротация-база).
//   - Refresh: валидация refresh-токена (подпись/exp/type='refresh'), проверка в БД,
//     ротация: старый hash — удалить, новый — вставить. Нет в БД (revoked) — domain.NewUnauthorized.
//   - Logout: удалить hash refresh-токена из БД (идемпотентно).
//   - Me: пользователь по user_id (из auth middleware контекста).

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"mytrip/internal/domain"
	"mytrip/internal/jwtutil"
	"mytrip/internal/repo"
)

// RefreshTokenDeleter — удаление refresh-токена с индикатором (для ротации: 0 строк → отозван).
type RefreshTokenDeleter interface {
	DeleteByHash(ctx context.Context, tokenHash string) (int64, error)
}

type authService struct {
	users       repo.UserRepo
	tokens      repo.RefreshTokenRepo
	deleter     RefreshTokenDeleter
	issuer      jwtutil.TokenIssuer
	validator   jwtutil.TokenValidator
	refreshTTL  time.Duration
}

// NewAuthService — конструктор AuthService.
// deleter — реализация RefreshTokenDeleter (обычно тот же PGRefreshTokenRepo); nil → ротация без проверки «отозван».
func NewAuthService(users repo.UserRepo, tokens repo.RefreshTokenRepo, deleter RefreshTokenDeleter, issuer jwtutil.TokenIssuer, validator jwtutil.TokenValidator, refreshTTL time.Duration) AuthService {
	return &authService{users: users, tokens: tokens, deleter: deleter, issuer: issuer, validator: validator, refreshTTL: refreshTTL}
}

// hashToken — SHA-256 hex refresh-токена (хранится в БД вместо самого JWT).
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Register — регистрация: bcrypt-хэш, INSERT users.
func (s *authService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	if email == "" || name == "" {
		return nil, domain.NewValidation("email and name are required", nil)
	}
	if len(password) < 8 {
		return nil, domain.NewValidation("password must be at least 8 characters", nil)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.NewInternal()
	}
	u := &domain.User{ID: uuid.NewString(), Email: email, Name: name, CreatedAt: time.Now()}
	if err := s.users.Insert(ctx, u, string(hash)); err != nil {
		return nil, err
	}
	return u, nil
}

// Login — проверка пароля, выдача access + refresh JWT.
func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	u, hash, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return nil, "", "", domain.NewUnauthorized("invalid email or password")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, "", "", domain.NewUnauthorized("invalid email or password")
	}
	access, err := s.issuer.IssueAccess(u.ID)
	if err != nil {
		return nil, "", "", domain.NewInternal()
	}
	refresh, err := s.issuer.IssueRefresh(u.ID)
	if err != nil {
		return nil, "", "", domain.NewInternal()
	}
	if err := s.tokens.Insert(ctx, u.ID, hashToken(refresh), time.Now().Add(s.refreshTTL)); err != nil {
		return nil, "", "", domain.NewInternal()
	}
	return u, access, refresh, nil
}

// Refresh — ротация refresh-токена: валидация JWT, проверка в БД, новый access + refresh.
func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := s.validator.Validate(refreshToken, jwtutil.TokenKindRefresh)
	if err != nil {
		return "", "", err
	}
	oldHash := hashToken(refreshToken)
	if _, err := s.users.ByID(ctx, claims.Subject); err != nil {
		return "", "", domain.NewUnauthorized("invalid or expired token")
	}
	if s.deleter != nil {
		n, err := s.deleter.DeleteByHash(ctx, oldHash)
		if err != nil {
			return "", "", domain.NewInternal()
		}
		if n == 0 {
			return "", "", domain.NewUnauthorized("invalid or expired token")
		}
	} else {
		if err := s.tokens.DeleteByHash(ctx, oldHash); err != nil {
			return "", "", domain.NewInternal()
		}
	}
	access, err := s.issuer.IssueAccess(claims.Subject)
	if err != nil {
		return "", "", domain.NewInternal()
	}
	newRefresh, err := s.issuer.IssueRefresh(claims.Subject)
	if err != nil {
		return "", "", domain.NewInternal()
	}
	if err := s.tokens.Insert(ctx, claims.Subject, hashToken(newRefresh), time.Now().Add(s.refreshTTL)); err != nil {
		return "", "", domain.NewInternal()
	}
	return access, newRefresh, nil
}

// Logout — удаление refresh-токена (hash) из БД; идемпотентно.
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	if _, err := s.validator.Validate(refreshToken, jwtutil.TokenKindRefresh); err != nil {
		return nil // токен уже невалиден — считаем logout выполненным
	}
	if err := s.tokens.DeleteByHash(ctx, hashToken(refreshToken)); err != nil {
		return domain.NewInternal()
	}
	return nil
}

// Me — пользователь по user_id.
func (s *authService) Me(ctx context.Context, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, domain.NewUnauthorized("unauthorized")
	}
	return s.users.ByID(ctx, userID)
}
