package jwtutil

import (
	"testing"
	"time"
)

func TestNewHS256_Valid(t *testing.T) {
	issuer, validator, err := NewHS256("test-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256: %v", err)
	}
	if issuer == nil || validator == nil {
		t.Fatal("expected non-nil issuer and validator")
	}
}

func TestNewHS256_EmptySecret(t *testing.T) {
	_, _, err := NewHS256("", time.Minute, time.Hour)
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestNewHS256_InvalidTTLs(t *testing.T) {
	if _, _, err := NewHS256("s", 0, time.Hour); err == nil {
		t.Fatal("expected error for zero accessTTL")
	}
	if _, _, err := NewHS256("s", time.Minute, 0); err == nil {
		t.Fatal("expected error for zero refreshTTL")
	}
}

func TestIssueAndValidate_Access(t *testing.T) {
	issuer, validator, err := NewHS256("test-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256: %v", err)
	}

	token, err := issuer.IssueAccess("user-123")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := validator.Validate(token, TokenKindAccess)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("expected subject user-123, got %q", claims.Subject)
	}
	if claims.Type != TokenKindAccess {
		t.Errorf("expected type access, got %q", claims.Type)
	}
}

func TestIssueAndValidate_Refresh(t *testing.T) {
	issuer, validator, err := NewHS256("test-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256: %v", err)
	}

	token, err := issuer.IssueRefresh("user-123")
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}

	claims, err := validator.Validate(token, TokenKindRefresh)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("expected subject user-123, got %q", claims.Subject)
	}
	if claims.Type != TokenKindRefresh {
		t.Errorf("expected type refresh, got %q", claims.Type)
	}
}

func TestValidate_WrongKind(t *testing.T) {
	issuer, validator, err := NewHS256("test-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256: %v", err)
	}

	// Выдаём refresh и пытаемся валидировать как access → должно упасть.
	token, err := issuer.IssueRefresh("user-123")
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}

	if _, err := validator.Validate(token, TokenKindAccess); err == nil {
		t.Fatal("expected error when validating refresh token as access")
	}

	// И наоборот: access → refresh.
	accessToken, err := issuer.IssueAccess("user-123")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := validator.Validate(accessToken, TokenKindRefresh); err == nil {
		t.Fatal("expected error when validating access token as refresh")
	}
}

func TestValidate_NoKindCheck(t *testing.T) {
	issuer, validator, err := NewHS256("test-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256: %v", err)
	}

	token, err := issuer.IssueAccess("user-123")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	// expectedKind == "" — не проверяем тип.
	claims, err := validator.Validate(token, "")
	if err != nil {
		t.Fatalf("Validate with empty kind: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("expected subject user-123, got %q", claims.Subject)
	}
}

func TestValidate_ExpiredToken(t *testing.T) {
	issuer, validator, err := NewHS256("test-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256: %v", err)
	}

	// Создаём token с уже истёкшим TTL (отрицательный offset).
	// Используем sign напрямую через NewHS256 с отрицательным TTL не получится
	// (конструктор требует положительный), поэтому создадим вручную.
	// Для теста expired используем очень короткий TTL и sleep.
	// Но проще: создадим issuer с 1мс TTL и дадим ему истечь.
	shortIssuer, shortValidator, err := NewHS256("test-secret", time.Millisecond, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256 short: %v", err)
	}

	token, err := shortIssuer.IssueAccess("user-expired")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	// Ждём, чтобы token истёк.
	time.Sleep(5 * time.Millisecond)

	if _, err := shortValidator.Validate(token, TokenKindAccess); err == nil {
		t.Fatal("expected error for expired token")
	}

	_ = issuer
	_ = validator
}

func TestValidate_MalformedToken(t *testing.T) {
	_, validator, err := NewHS256("test-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256: %v", err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"not-jwt", "not-a-jwt-token"},
		{"tampered", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0In0.invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := validator.Validate(tt.token, TokenKindAccess); err == nil {
				t.Errorf("expected error for malformed token %q", tt.token)
			}
		})
	}
}

func TestValidate_WrongSecret(t *testing.T) {
	issuer, _, _ := NewHS256("secret-a", time.Minute, time.Hour)
	_, validatorB, _ := NewHS256("secret-b", time.Minute, time.Hour)

	token, err := issuer.IssueAccess("user-123")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	if _, err := validatorB.Validate(token, TokenKindAccess); err == nil {
		t.Fatal("expected error when validating token signed with different secret")
	}
}
