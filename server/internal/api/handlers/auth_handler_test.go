package handlers

// Unit-тесты AuthHandler (CHAT-07: A7-01).
// httptest + мок AuthService.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/domain"
	"mytrip/internal/service"
)

type mockAuthService struct {
	registerFn func(ctx context.Context, email, password, name string) (*domain.User, error)
	loginFn    func(ctx context.Context, email, password string) (*domain.User, string, string, error)
	refreshFn  func(ctx context.Context, refreshToken string) (string, string, error)
	logoutFn   func(ctx context.Context, refreshToken string) error
	meFn       func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *mockAuthService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, email, password, name)
	}
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, email, password)
	}
	return nil, "", "", nil
}

func (m *mockAuthService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	if m.refreshFn != nil {
		return m.refreshFn(ctx, refreshToken)
	}
	return "", "", nil
}

func (m *mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, refreshToken)
	}
	return nil
}

func (m *mockAuthService) Me(ctx context.Context, userID string) (*domain.User, error) {
	if m.meFn != nil {
		return m.meFn(ctx, userID)
	}
	return nil, nil
}

var _ service.AuthService = (*mockAuthService)(nil)

func newTestRouter(auth *mockAuthService) chi.Router {
	r := chi.NewRouter()
	NewAuthHandler(auth).Mount(r)
	return r
}

func TestRegister_Success(t *testing.T) {
	auth := &mockAuthService{registerFn: func(_ context.Context, _, _, _ string) (*domain.User, error) {
		return &domain.User{ID: "u1", Email: "a@b.com", Name: "Alice", CreatedAt: time.Now()}, nil
	}}
	body := bytes.NewBufferString(`{"email":"a@b.com","password":"password123","name":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	w := httptest.NewRecorder()
	newTestRouter(auth).ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.User.ID != "u1" || res.User.Email != "a@b.com" || res.User.Name != "Alice" {
		t.Fatalf("wrong user: %+v", res.User)
	}
}

func TestRegister_Conflict(t *testing.T) {
	auth := &mockAuthService{registerFn: func(_ context.Context, _, _, _ string) (*domain.User, error) {
		return nil, domain.NewConflict("email already registered")
	}}
	body := bytes.NewBufferString(`{"email":"a@b.com","password":"password123","name":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	w := httptest.NewRecorder()
	newTestRouter(auth).ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogin_Success(t *testing.T) {
	auth := &mockAuthService{loginFn: func(_ context.Context, _, _ string) (*domain.User, string, string, error) {
		return &domain.User{ID: "u1", Email: "a@b.com", Name: "Alice", CreatedAt: time.Now()}, "acc", "ref", nil
	}}
	body := bytes.NewBufferString(`{"email":"a@b.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	w := httptest.NewRecorder()
	newTestRouter(auth).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.AccessToken != "acc" {
		t.Fatalf("wrong token: %q", res.AccessToken)
	}
	cookies := w.Result().Cookies()
	var found *http.Cookie
	for _, c := range cookies {
		if c.Name == "rt" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("rt cookie not set")
	}
	if !found.HttpOnly || found.Path != "/" || found.SameSite != http.SameSiteLaxMode || found.Value != "ref" {
		t.Fatalf("wrong cookie: %+v", found)
	}
	if found.MaxAge != RefreshCookieMaxAge {
		t.Fatalf("wrong MaxAge: %d", found.MaxAge)
	}
}

func TestLogin_Unauthorized(t *testing.T) {
	auth := &mockAuthService{loginFn: func(_ context.Context, _, _ string) (*domain.User, string, string, error) {
		return nil, "", "", domain.NewUnauthorized("invalid email or password")
	}}
	body := bytes.NewBufferString(`{"email":"a@b.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	w := httptest.NewRecorder()
	newTestRouter(auth).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRefresh_Success(t *testing.T) {
	auth := &mockAuthService{refreshFn: func(_ context.Context, rt string) (string, string, error) {
		if rt != "old-ref" {
			t.Fatalf("expected cookie token, got %q", rt)
		}
		return "acc2", "new-ref", nil
	}}
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "rt", Value: "old-ref"})
	w := httptest.NewRecorder()
	newTestRouter(auth).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.AccessToken != "acc2" {
		t.Fatalf("wrong token: %q", res.AccessToken)
	}
	var found *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "rt" {
			found = c
		}
	}
	if found == nil || found.Value != "new-ref" {
		t.Fatalf("expected rotated cookie, got %+v", found)
	}
}

func TestRefresh_NoCookie(t *testing.T) {
	auth := &mockAuthService{refreshFn: func(_ context.Context, rt string) (string, string, error) {
		return "", "", domain.NewUnauthorized("invalid or expired token")
	}}
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	w := httptest.NewRecorder()
	newTestRouter(auth).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogout_Success(t *testing.T) {
	called := false
	auth := &mockAuthService{logoutFn: func(_ context.Context, rt string) error {
		called = true
		if rt != "ref" {
			t.Fatalf("expected token from cookie, got %q", rt)
		}
		return nil
	}}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "rt", Value: "ref"})
	w := httptest.NewRecorder()
	newTestRouter(auth).ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
	if !called {
		t.Fatal("logout not called")
	}
	var found *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "rt" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("cookie not cleared")
	}
	if found.MaxAge != -1 {
		t.Fatalf("expected MaxAge=-1, got %d", found.MaxAge)
	}
}

func TestMe_Success(t *testing.T) {
	auth := &mockAuthService{meFn: func(_ context.Context, uid string) (*domain.User, error) {
		if uid != "u1" {
			t.Fatalf("expected u1 from context, got %q", uid)
		}
		return &domain.User{ID: "u1", Email: "a@b.com", Name: "Alice", CreatedAt: time.Now()}, nil
	}}
	r := chi.NewRouter()
	NewAuthHandler(auth).Mount(r)
	// обёртка: кладём user_id в контекст как auth middleware
	inner := r
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), chiRouteKey{}, "u1")
		// используем middleware.AuthContext ключ
		ctx = context.WithValue(ctx, contextKeyUserID, "u1")
		inner.ServeHTTP(w, req.WithContext(ctx))
	})
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.ID != "u1" || res.Email != "a@b.com" {
		t.Fatalf("wrong user: %+v", res)
	}
}

func TestMe_Unauthorized(t *testing.T) {
	auth := &mockAuthService{meFn: func(_ context.Context, uid string) (*domain.User, error) {
		return nil, domain.NewUnauthorized("unauthorized")
	}}
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()
	newTestRouter(auth).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

type chiRouteKey struct{}
