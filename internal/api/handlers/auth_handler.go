package handlers

// AuthHandler — /api/v1/auth/* (CHAT-07: A7-01).
// Контракты: см. internal/api/handlers/handlers.go (комментарий AuthHandler).
// Тела — через service.AuthService; парсинг JWT и Set-Cookie — здесь (handler).
// Cookie-имя 'rt', HttpOnly, Path=/, SameSite=Lax, Max-Age = refresh TTL (2592000).

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/domain"
	"mytrip/internal/service"
)

// RefreshCookieName — имя httpOnly cookie для refresh-токена.
const RefreshCookieName = "rt"

// RefreshCookieMaxAge — TTL cookie 'rt' в секундах (30 дней = 2592000).
const RefreshCookieMaxAge = 2592000

type authHandler struct {
	auth service.AuthService
}

// NewAuthHandler — конструктор AuthHandler.
func NewAuthHandler(auth service.AuthService) *AuthHandler {
	return &AuthHandler{handler: &authHandler{auth: auth}}
}

// AuthHandler — реализация контракта из handlers.go.
type AuthHandler struct {
	handler *authHandler
}

// Mount — регистрация маршрутов на chi router (путь относительно /api/v1).
func (h *AuthHandler) Mount(r chi.Router) {
	r.Post("/auth/register", h.handler.register)
	r.Post("/auth/login", h.handler.login)
	r.Post("/auth/refresh", h.handler.refresh)
	r.Post("/auth/logout", h.handler.logout)
	r.Get("/users/me", h.handler.me)
}

type userJSON struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserJSON(u *domain.User) userJSON {
	return userJSON{ID: u.ID, Email: u.Email, Name: u.Name, CreatedAt: u.CreatedAt}
}

func (h *authHandler) setRefreshCookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// POST /auth/register — 201 {user, access_token} + cookie rt.
func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var body authRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	user, tokens, err := h.auth.Register(r.Context(), body.Email, body.Password, body.Name)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	h.setRefreshCookie(w, tokens.RefreshToken, RefreshCookieMaxAge)
	middleware.WriteJSON(w, http.StatusCreated, map[string]any{
		"user":         toUserJSON(user),
		"access_token": tokens.AccessToken,
	})
}

// POST /auth/login — 200 {user, access_token} + cookie rt.
func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var body authRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	user, tokens, err := h.auth.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	h.setRefreshCookie(w, tokens.RefreshToken, RefreshCookieMaxAge)
	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"user":         toUserJSON(user),
		"access_token": tokens.AccessToken,
	})
}

// POST /auth/refresh — cookie rt → 200 {access_token} + ротация cookie.
func (h *authHandler) refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil || cookie.Value == "" {
		middleware.WriteError(w, domain.NewUnauthorized("refresh cookie is missing"))
		return
	}
	user, tokens, err := h.auth.Refresh(r.Context(), cookie.Value)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	h.setRefreshCookie(w, tokens.RefreshToken, RefreshCookieMaxAge)
	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"user":         toUserJSON(user),
		"access_token": tokens.AccessToken,
	})
}

// POST /auth/logout — 204 + очистка cookie rt.
func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.Logout(r.Context()); err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	w.Header().Set("Set-Cookie", http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}.String())
	w.WriteHeader(http.StatusNoContent)
}

// GET /users/me — user_id из контекста (auth middleware) → 200 {user}.
func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	u, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, toUserJSON(u))
}
