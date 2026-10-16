package service

// AuthService — аутентификация (CHAT-07: A7-01).
// Контракт: internal/service/interfaces.go (интерфейс AuthService).
//
// Логика:
//   - Register: bcrypt-хэш пароля, INSERT users; дубль email — domain.NewConflict (из репо).
//   - Login: bcrypt CompareHashAndPassword; выдача access + refresh JWT через jwtutil.
//     SHA-256 hash нового refresh-токена сохраняется в refresh_tokens (база ротации).
//   - Refresh: валидация refresh-токена (подпись/exp/type='refresh'), проверка в БД,
//     ротация: старый hash — удалить, новый — вставить. Нет в БД (revoked) — domain.NewUnauthorized.
//   - Logout: удалить hash refresh-токена из БД (идемпотентно).
//   - Me: пользователь по user_id (из auth middleware контекста).

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"mytrip/internal/domain"
	"mytrip/internal/jwtutil"
	"mytrip/internal/repo"
)

// newUserID — UUID v4 без внешних зависимостей (RFC 4122, вариант 1).
func newUserID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// RefreshTokenDeleter — удаление refresh-токена с индикатором (для ротации: 0 строк → отозван).
type RefreshTokenDeleter interface {
	DeleteByHash(ctx context.Context, tokenHash string) (int64, error)
}

type authService struct {
	users      repo.UserRepo
	tokens     repo.RefreshTokenRepo
	deleter    RefreshTokenDeleter
	issuer     jwtutil.TokenIssuer
	validator  jwtutil.TokenValidator
	refreshTTL time.Duration
}

// NewAuthService — конструктор AuthService.
// deleter — реализация RefreshTokenDeleter (обычно репо с DeleteByHashWithResult); nil → ротация без проверки «отозван».
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
	u := &domain.User{ID: newUserID(), Email: email, Name: name, CreatedAt: time.Now()}
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
	deleted, err := s.deleteOld(ctx, oldHash)
	if err != nil {
		return "", "", domain.NewInternal()
	}
	if !deleted {
		return "", "", domain.NewUnauthorized("invalid or expired token")
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

// deleteOld — удаление старого hash; true если строка была удалена.
func (s *authService) deleteOld(ctx context.Context, hash string) (bool, error) {
	if s.deleter != nil {
		n, err := s.deleter.DeleteByHash(ctx, hash)
		if err != nil {
			return false, err
		}
		return n > 0, nil
	}
	if err := s.tokens.DeleteByHash(ctx, hash); err != nil {
		return false, err
	}
	return true, nil
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
