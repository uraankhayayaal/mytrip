package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mytrip/internal/jwtutil"
)

const testSecret = "test-middleware-secret"

func newTestValidator() jwtutil.TokenValidator {
	_, validator, err := jwtutil.NewHS256(testSecret, time.Minute, time.Hour)
	if err != nil {
		panic(err)
	}
	return validator
}

func newTestIssuer() jwtutil.TokenIssuer {
	issuer, _, err := jwtutil.NewHS256(testSecret, time.Minute, time.Hour)
	if err != nil {
		panic(err)
	}
	return issuer
}

type errorResponse struct {
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func doReq(t *testing.T, h http.Handler, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET", "/test", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestAuth_NoHeader(t *testing.T) {
	h := Auth(testSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}))

	w := doReq(t, h, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	var body errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Error.Code != "unauthorized" {
		t.Errorf("expected code unauthorized, got %q", body.Error.Code)
	}
}

func TestAuth_MalformedHeader(t *testing.T) {
	h := Auth(testSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}))

	malformed := []string{
		"Token abc123",
		"Bearer",
		"Bearer ",
		"bearer abc123", // lowercase — accept (strings.EqualFold)
	}

	for i, header := range malformed {
		// "bearer abc123" is actually valid (case-insensitive prefix), so skip it for malformed test.
		if i == 3 {
			continue
		}
		w := doReq(t, h, header)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("test %d (header=%q): expected 401, got %d", i, header, w.Code)
		}
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	// Используем очень короткий TTL, чтобы token истёк.
	shortIssuer, _, err := jwtutil.NewHS256(testSecret, time.Millisecond, time.Hour)
	if err != nil {
		t.Fatalf("NewHS256: %v", err)
	}
	token, err := shortIssuer.IssueAccess("user-123")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	time.Sleep(5 * time.Millisecond)

	h := Auth(testSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}))

	w := doReq(t, h, "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", w.Code)
	}
}

func TestAuth_RefreshInsteadOfAccess(t *testing.T) {
	issuer := newTestIssuer()
	token, err := issuer.IssueRefresh("user-123")
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}

	h := Auth(testSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}))

	w := doReq(t, h, "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for refresh token used as access, got %d", w.Code)
	}

	var body errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Error.Code != "unauthorized" {
		t.Errorf("expected code unauthorized, got %q", body.Error.Code)
	}
}

func TestAuth_MalformedToken(t *testing.T) {
	h := Auth(testSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}))

	w := doReq(t, h, "Bearer not-a-valid-jwt")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for malformed token, got %d", w.Code)
	}
}

func TestAuth_Success(t *testing.T) {
	issuer := newTestIssuer()
	token, err := issuer.IssueAccess("user-123")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	var gotUserID string
	h := Auth(testSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = AuthContext(r.Context())
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}))

	w := doReq(t, h, "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUserID != "user-123" {
		t.Errorf("expected user_id user-123 in context, got %q", gotUserID)
	}
}

func TestAuth_Success_LowercaseBearer(t *testing.T) {
	issuer := newTestIssuer()
	token, err := issuer.IssueAccess("user-456")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	var gotUserID string
	h := Auth(testSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = AuthContext(r.Context())
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}))

	w := doReq(t, h, "bearer "+token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for lowercase bearer, got %d", w.Code)
	}
	if gotUserID != "user-456" {
		t.Errorf("expected user_id user-456, got %q", gotUserID)
	}
}

func TestAuthContext_Empty(t *testing.T) {
	// AuthContext без установленного ключа возвращает пустую строку.
	import_ctx := context.Background()
	if got := AuthContext(import_ctx); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
