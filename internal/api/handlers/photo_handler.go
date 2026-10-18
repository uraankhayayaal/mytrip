package handlers

// PhotoHandler — CRUD /trips/{id}/photos + presign (CHAT-07: A7-04).
// Контракт: internal/api/handlers/handlers.go (PhotoHandler).
// Маршруты (относительно /api/v1):
//
//	GET    /trips/{id}/photos        → 200 [PhotoJSON]
//	POST   /trips/{id}/photos        → 201 PhotoJSON; body {object_key, url, cover?}
//	DELETE /trips/{id}/photos/{photoId} → 204
//	POST   /trips/{id}/photos/presign   → 200 {url, expires_at} (frontend PhotoUrlResponse)
//
// JSON — через middleware.WriteJSON; ошибки — через middleware.WriteError.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/domain"
	"mytrip/internal/service"
)

// photoServiceWithPresign — PhotoService + presign (StoragePort) для PhotoHandler.
// Presign вынесен из контракта PhotoService (interfaces.go менять нельзя).
type photoServiceWithPresign interface {
	List(ctx context.Context, userID, tripID string) ([]*domain.Photo, error)
	Create(ctx context.Context, userID, tripID, objectKey, url string, cover bool) (*domain.Photo, error)
	Delete(ctx context.Context, userID, tripID, photoID string) error
	PresignPut(ctx context.Context, objectKey string) (string, time.Time, error)
}

// photoHandler — реализация PhotoHandler (identity-реестр, как в tripHandler).
type photoHandler struct {
	photos service.PhotoService
	svc    photoServiceWithPresign
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
func (h *PhotoHandler) Mount(r chi.Router) {
	impl := h.impl()
	r.Route("/trips/{id}/photos", func(r chi.Router) {
		r.Get("", impl.list)
		r.Post("", impl.create)
		r.Post("/presign", impl.presign)
		r.Delete("/{photoId}", impl.delete)
	})
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

// photoRequest — body POST /trips/{id}/photos: {object_key, url, cover?}.
type photoRequest struct {
	ObjectKey string  `json:"object_key"`
	URL       string  `json:"url"`
	Cover     *bool   `json:"cover,omitempty"`
}

// presignRequest — body POST /trips/{id}/photos/presign: {object_key}.
type presignRequest struct {
	ObjectKey string `json:"object_key"`
}

// presignResponse — ответ presign: {url, expires_at} (frontend PhotoUrlResponse).
type presignResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
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

// POST /trips/{id}/photos — создание фото (object_key, url, cover?).
func (h *photoHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	var body photoRequest
	if err := decodeJSON(r, &body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	if body.ObjectKey == "" || body.URL == "" {
		middleware.WriteError(w, domain.NewValidation("object_key and url are required", map[string]any{"field": "object_key"}))
		return
	}
	cover := body.Cover != nil && *body.Cover
	p, err := h.photos.Create(r.Context(), userID, tripID, body.ObjectKey, body.URL, cover)
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

// POST /trips/{id}/photos/presign — pre-signed PUT URL для загрузки фото.
// Ответ: {url, expires_at}.
func (h *photoHandler) presign(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	// Владение поездкой проверяется через PhotoService.List (TripRepo.Get внутри).
	if _, err := h.photos.List(r.Context(), userID, tripID); err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	var body presignRequest
	if err := decodeJSON(r, &body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	if body.ObjectKey == "" {
		middleware.WriteError(w, domain.NewValidation("object_key is required", map[string]any{"field": "object_key"}))
		return
	}
	url, expiresAt, err := h.svc.PresignPut(r.Context(), body.ObjectKey)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, presignResponse{URL: url, ExpiresAt: expiresAt})
}

// decodeJSON — чтение JSON-тела с защитой от пустого/слишком большого тела.
func decodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	return dec.Decode(v)
}
