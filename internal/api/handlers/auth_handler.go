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
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

func (h *authHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (h *authHandler) refreshCookie(r *http.Request) string {
	c, err := r.Cookie(RefreshCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// POST /auth/register — {"email","password","name"} → 201 {"user":{...}}.
func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	u, err := h.auth.Register(r.Context(), body.Email, body.Password, body.Name)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, map[string]any{"user": toUserJSON(u)})
}

// POST /auth/login — {"email","password"} → 200 + Set-Cookie rt.
func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	u, access, refresh, err := h.auth.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	h.setRefreshCookie(w, refresh, RefreshCookieMaxAge)
	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"access_token": access,
		"user":         toUserJSON(u),
	})
}

// POST /auth/refresh — cookie 'rt' → 200 {"access_token"} + Set-Cookie rt (ротация).
func (h *authHandler) refresh(w http.ResponseWriter, r *http.Request) {
	token := h.refreshCookie(r)
	access, newRefresh, err := h.auth.Refresh(r.Context(), token)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	h.setRefreshCookie(w, newRefresh, RefreshCookieMaxAge)
	middleware.WriteJSON(w, http.StatusOK, map[string]any{"access_token": access})
}

// POST /auth/logout — cookie 'rt' → 204 + Clear-Cookie.
func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.Logout(r.Context(), h.refreshCookie(r)); err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	h.clearRefreshCookie(w)
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

func toAppError(err error) *domain.AppError {
	if ae, ok := err.(*domain.AppError); ok {
		return ae
	}
	return domain.NewInternal()
}
