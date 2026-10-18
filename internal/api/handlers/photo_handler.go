package handlers

// PhotoHandler — CRUD /trips/{id}/photos + presign (CHAT-07: A7-04).
// Контракт: internal/api/handlers/handlers.go (PhotoHandler).
// Маршруты (относительно /api/v1) — см. photo_handler_mount.go.
// JSON — через middleware.WriteJSON; ошибки — через middleware.WriteError.

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/domain"
	"mytrip/internal/service"
)

// photoHandler — реализация PhotoHandler (identity-реестр, как в tripHandler).
type photoHandler struct {
	photos service.PhotoService
	svc    *compositePhotoService
}

// photoHandlers — реестр: ключ *PhotoHandler, значение — внутренняя реализация.
var photoHandlers = make(map[*PhotoHandler]*photoHandler)

// NewPhotoHandler — конструктор PhotoHandler.
// photos — PhotoService; presign — StoragePort (порт presign-URL'ов).
func NewPhotoHandler(photos service.PhotoService, presign service.StoragePort) *PhotoHandler {
	h := &PhotoHandler{}
	photoHandlers[h] = &photoHandler{
		photos: photos,
		svc:    &compositePhotoService{PhotoService: photos, Storage: presign},
	}
	return h
}

// compositePhotoService — PhotoService + PresignPut (StoragePort).
type compositePhotoService struct {
	service.PhotoService
	Storage service.StoragePort
}

func (c *compositePhotoService) PresignPut(ctx context.Context, objectKey string) (string, time.Time, error) {
	if c.Storage == nil {
		return "", time.Time{}, domain.NewInternal()
	}
	return c.Storage.PresignPut(ctx, objectKey)
}

// Mount — регистрация маршрутов на chi router (путь относительно /api/v1).
// Роуты вешаются ПОД auth-middleware (см. cmd/server/main.go).
// Паттерны — см. photo_handler_mount.go (mountPhotoRoutes).
func (h *PhotoHandler) Mount(r chi.Router) {
	mountPhotoRoutes(r, h.impl())
}

// impl — приватный доступ к реализации по identity-ключу.
func (h *PhotoHandler) impl() *photoHandler {
	if impl, ok := photoHandlers[h]; ok {
		return impl
	}
	return &photoHandler{photos: nil, svc: nil}
}

// photoJSON — JSON-представление фото на границе API (frontend Photo).
type photoJSON struct {
	ID        string    `json:"id"`
	TripID    string    `json:"trip_id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

func toPhotoJSON(p *domain.Photo) photoJSON {
	return photoJSON{ID: p.ID, TripID: p.TripID, URL: p.URL, CreatedAt: p.CreatedAt}
}

// photoRequest — DTO запроса создания фото: {object_key, url, cover?}.
type photoRequest struct {
	ObjectKey string `json:"object_key"`
	URL       string `json:"url"`
	Cover     bool   `json:"cover"`
}

// GET /trips/{id}/photos — список фото поездки.
func (h *photoHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	photos, err := h.photos.List(r.Context(), userID, tripID)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	if photos == nil {
		photos = []*domain.Photo{}
	}
	out := make([]photoJSON, 0, len(photos))
	for _, p := range photos {
		out = append(out, toPhotoJSON(p))
	}
	middleware.WriteJSON(w, http.StatusOK, out)
}

// POST /trips/{id}/photos — создание фото.
func (h *photoHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	var body photoRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	if body.ObjectKey == "" || body.URL == "" {
		middleware.WriteError(w, domain.NewValidation("object_key and url are required", nil))
		return
	}
	p, err := h.photos.Create(r.Context(), userID, tripID, body.ObjectKey, body.URL, body.Cover)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, toPhotoJSON(p))
}

// DELETE /trips/{id}/photos/{photoId} — удаление фото → 204.
func (h *photoHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	photoID := chi.URLParam(r, "photoId")
	if err := h.photos.Delete(r.Context(), userID, tripID, photoID); err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /trips/{id}/photos/presign — pre-signed PUT URL для прямой загрузки.
// Контракт (frontend PhotoUrlResponse): {"url": "<presigned>", "expires_at": "<RFC3339>"}.
func (h *photoHandler) presign(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ObjectKey string `json:"object_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	if body.ObjectKey == "" {
		middleware.WriteError(w, domain.NewValidation("object_key is required", nil))
		return
	}
	if h.svc == nil {
		middleware.WriteError(w, domain.NewInternal())
		return
	}
	url, expiresAt, err := h.svc.PresignPut(r.Context(), body.ObjectKey)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]any{
		"url":        url,
		"expires_at": expiresAt,
	})
}
