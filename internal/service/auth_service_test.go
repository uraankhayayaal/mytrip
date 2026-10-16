package service

// Unit-тесты AuthService (CHAT-07: A7-01).
// Мок UserRepo/RefreshTokenRepo — в этом файле.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"mytrip/internal/domain"
	"mytrip/internal/jwtutil"
	"golang.org/x/crypto/bcrypt"
)

// --- mocks ---

type mockUserRepo struct {
	byEmailFn func(ctx context.Context, email string) (*domain.User, string, error)
	byIDFn    func(ctx context.Context, id string) (*domain.User, error)
	insertFn  func(ctx context.Context, u *domain.User, passwordHash string) error
}

func (m *mockUserRepo) Insert(ctx context.Context, u *domain.User, passwordHash string) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, u, passwordHash)
	}
	return nil
}

func (m *mockUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	if m.byEmailFn != nil {
		return m.byEmailFn(ctx, email)
	}
	return nil, "", domain.NewNotFound("user not found")
}

func (m *mockUserRepo) ByID(ctx context.Context, id string) (*domain.User, error) {
	if m.byIDFn != nil {
		return m.byIDFn(ctx, id)
	}
	return nil, domain.NewNotFound("user not found")
}

type mockRefreshTokenRepo struct {
	insertFn   func(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	deleteFn   func(ctx context.Context, tokenHash string) error
	deleteUser func(ctx context.Context, userID string) error
}

func (m *mockRefreshTokenRepo) Insert(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, userID, tokenHash, expiresAt)
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteByHash(ctx context.Context, tokenHash string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, tokenHash)
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteByUser(ctx context.Context, userID string) error {
	if m.deleteUser != nil {
		return m.deleteUser(ctx, userID)
	}
	return nil
}

type mockDeleter struct {
	deleted int64
	err     error
}

func (m *mockDeleter) DeleteByHashWithResult(ctx context.Context, tokenHash string) (int64, error) {
	return m.deleted, m.err
}

type mockIssuer struct {
	access  string
	refresh string
	err     error
}

func (m *mockIssuer) IssueAccess(userID string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.access, nil
}

func (m *mockIssuer) IssueRefresh(userID string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.refresh, nil
}

type mockValidator struct {
	claims *jwtutil.Claims
	err    error
}

func (m *mockValidator) Validate(token string, expectedKind jwtutil.TokenKind) (*jwtutil.Claims, error) {
	return m.claims, m.err
}

func newTestService(users *mockUserRepo, tokens *mockRefreshTokenRepo, del *mockDeleter, iss *mockIssuer, val *mockValidator) AuthService {
	return NewAuthService(users, tokens, del, iss, val, time.Hour)
}
func testHash(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h)
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	var inserted *domain.User
	var insertedHash string
	users := &mockUserRepo{insertFn: func(_ context.Context, u *domain.User, hash string) error {
		inserted = u
		insertedHash = hash
		return nil
	}}
	svc := newTestService(users, &mockRefreshTokenRepo{}, nil, &mockIssuer{}, &mockValidator{})
	u, err := svc.Register(context.Background(), "a@b.com", "password123", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email != "a@b.com" || u.Name != "Alice" {
		t.Fatalf("wrong user: %+v", u)
	}
	if inserted == nil || insertedHash == "" {
		t.Fatal("user not inserted or hash empty")
	}
	if !strings.HasPrefix(insertedHash, "$2") {
		t.Fatalf("hash is not bcrypt: %s", insertedHash)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	svc := newTestService(&mockUserRepo{}, &mockRefreshTokenRepo{}, nil, &mockIssuer{}, &mockValidator{})
	_, err := svc.Register(context.Background(), "a@b.com", "short", "Alice")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeValidation {
		t.Fatalf("expected validation error, got: %v", err)
	}
}

func TestRegister_Conflict(t *testing.T) {
	users := &mockUserRepo{insertFn: func(_ context.Context, _ *domain.User, _ string) error {
		return domain.NewConflict("email already registered")
	}}
	svc := newTestService(users, &mockRefreshTokenRepo{}, nil, &mockIssuer{}, &mockValidator{})
	_, err := svc.Register(context.Background(), "a@b.com", "password123", "Alice")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeConflict {
		t.Fatalf("expected conflict, got: %v", err)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	users := &mockUserRepo{byEmailFn: func(_ context.Context, _ string) (*domain.User, string, error) {
		return &domain.User{ID: "u1", Email: "a@b.com", Name: "Alice"}, testHash("password123"), nil
	}}
	tokens := &mockRefreshTokenRepo{}
	iss := &mockIssuer{access: "acc", refresh: "ref"}
	svc := newTestService(users, tokens, nil, iss, &mockValidator{})
	u, acc, ref, err := svc.Login(context.Background(), "a@b.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != "u1" || acc != "acc" || ref != "ref" {
		t.Fatalf("wrong result: %+v %q %q", u, acc, ref)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	users := &mockUserRepo{byEmailFn: func(_ context.Context, _ string) (*domain.User, string, error) {
		return nil, "", domain.NewNotFound("user not found")
	}}
	svc := newTestService(users, &mockRefreshTokenRepo{}, nil, &mockIssuer{}, &mockValidator{})
	_, _, _, err := svc.Login(context.Background(), "a@b.com", "password123")
	if err == nil {
		t.Fatal("expected unauthorized")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeUnauthorized {
		t.Fatalf("expected unauthorized, got: %v", err)
	}
}

// --- Refresh ---

func TestRefresh_Success(t *testing.T) {
	users := &mockUserRepo{byIDFn: func(_ context.Context, _ string) (*domain.User, error) {
		return &domain.User{ID: "u1"}, nil
	}}
	var insertedHash string
	tokens := &mockRefreshTokenRepo{insertFn: func(_ context.Context, _, h string, _ time.Time) error {
		insertedHash = h
		return nil
	}}
	iss := &mockIssuer{access: "acc2", refresh: "ref2"}
	val := &mockValidator{claims: &jwtutil.Claims{Subject: "u1"}, err: nil}
	svc := newTestService(users, tokens, &mockDeleter{deleted: 1}, iss, val)
	acc, ref, err := svc.Refresh(context.Background(), "old-refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc != "acc2" || ref != "ref2" {
		t.Fatalf("wrong tokens: %q %q", acc, ref)
	}
	if insertedHash == "" {
		t.Fatal("new token hash not inserted")
	}
}

func TestRefresh_Revoked(t *testing.T) {
	users := &mockUserRepo{byIDFn: func(_ context.Context, _ string) (*domain.User, error) {
		return &domain.User{ID: "u1"}, nil
	}}
	val := &mockValidator{claims: &jwtutil.Claims{Subject: "u1"}, err: nil}
	svc := newTestService(users, &mockRefreshTokenRepo{}, &mockDeleter{deleted: 0}, &mockIssuer{}, val)
	_, _, err := svc.Refresh(context.Background(), "old-refresh")
	if err == nil {
		t.Fatal("expected unauthorized for revoked token")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeUnauthorized {
		t.Fatalf("expected unauthorized, got: %v", err)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	val := &mockValidator{claims: nil, err: domain.NewUnauthorized("invalid or expired token")}
	svc := newTestService(&mockUserRepo{}, &mockRefreshTokenRepo{}, nil, &mockIssuer{}, val)
	_, _, err := svc.Refresh(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Logout ---

func TestLogout_Success(t *testing.T) {
	deleted := false
	tokens := &mockRefreshTokenRepo{deleteFn: func(_ context.Context, _ string) error {
		deleted = true
		return nil
	}}
	val := &mockValidator{claims: &jwtutil.Claims{Subject: "u1"}, err: nil}
	svc := newTestService(&mockUserRepo{}, tokens, nil, &mockIssuer{}, val)
	if err := svc.Logout(context.Background(), "ref"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("token hash not deleted")
	}
}

func TestLogout_EmptyToken(t *testing.T) {
	deleted := false
	tokens := &mockRefreshTokenRepo{deleteFn: func(_ context.Context, _ string) error {
		deleted = true
		return nil
	}}
	svc := newTestService(&mockUserRepo{}, tokens, nil, &mockIssuer{}, &mockValidator{})
	if err := svc.Logout(context.Background(), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Fatal("should not delete on empty token")
	}
}

// --- Me ---

func TestMe_Success(t *testing.T) {
	users := &mockUserRepo{byIDFn: func(_ context.Context, id string) (*domain.User, error) {
		return &domain.User{ID: id, Email: "a@b.com", Name: "Alice"}, nil
	}}
	svc := newTestService(users, &mockRefreshTokenRepo{}, nil, &mockIssuer{}, &mockValidator{})
	u, err := svc.Me(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != "u1" {
		t.Fatalf("wrong user: %+v", u)
	}
}

func TestMe_EmptyUserID(t *testing.T) {
	svc := newTestService(&mockUserRepo{}, &mockRefreshTokenRepo{}, nil, &mockIssuer{}, &mockValidator{})
	_, err := svc.Me(context.Background(), "")
	if err == nil {
		t.Fatal("expected unauthorized")
	}
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeUnauthorized {
		t.Fatalf("expected unauthorized, got: %v", err)
	}
}
