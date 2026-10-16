package handlers

// TripHandler — CRUD /trips (CHAT-07: A7-02).
// Контракт: internal/api/handlers/handlers.go (TripHandler).
// Маршруты (относительно /api/v1): GET/POST /trips, GET/PUT/DELETE /trips/{id}.
// JSON — через middleware.WriteJSON; ошибки — через middleware.WriteError.

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/domain"
	"mytrip/internal/service"
)

type tripHandler struct {
	trips service.TripService
}

// NewTripHandler — конструктор TripHandler.
// Контракт TripHandler зафиксирован в handlers.go — расширяем через embed.
type TripHandler struct {
	*tripHandler
}

// NewTripHandler — конструктор TripHandler.
func NewTripHandler(trips service.TripService) *TripHandler {
	return &TripHandler{tripHandler: &tripHandler{trips: trips}}
}

// Mount — регистрация маршрутов на chi router (путь относительно /api/v1).
// Роуты вешаются ПОД auth-middleware (см. cmd/server/main.go).
func (h *TripHandler) Mount(r chi.Router) {
	r.Get("/trips", h.list)
	r.Post("/trips", h.create)
	r.Get("/trips/{id}", h.get)
	r.Put("/trips/{id}", h.update)
	r.Delete("/trips/{id}", h.delete)
}

type tripJSON struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	CoverPhotoURL *string   `json:"cover_photo_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toTripJSON(t *domain.Trip) tripJSON {
	return tripJSON{
		ID:            t.ID,
		Title:         t.Title,
		Description:   t.Description,
		StartDate:     t.StartDate,
		EndDate:       t.EndDate,
		CoverPhotoURL: t.CoverPhotoURL,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

type tripRequest struct {
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	CoverPhotoURL *string    `json:"cover_photo_url,omitempty"`
}

func (b tripRequest) payload() *service.TripPayload {
	return &service.TripPayload{
		Title:         b.Title,
		Description:   b.Description,
		StartDate:     b.StartDate,
		EndDate:       b.EndDate,
		CoverPhotoURL: b.CoverPhotoURL,
	}
}

// GET /trips — список поездок пользователя.
func (h *tripHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	trips, err := h.trips.List(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	if trips == nil {
		trips = []*domain.Trip{}
	}
	out := make([]tripJSON, 0, len(trips))
	for _, t := range trips {
		out = append(out, toTripJSON(t))
	}
	middleware.WriteJSON(w, http.StatusOK, out)
}

// POST /trips — создание поездки.
func (h *tripHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	var body tripRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	t, err := h.trips.Create(r.Context(), userID, body.payload())
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, toTripJSON(t))
}

// GET /trips/{id} — поездка по id.
func (h *tripHandler) get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	t, err := h.trips.Get(r.Context(), userID, tripID)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, toTripJSON(t))
}

// PUT /trips/{id} — обновление поездки.
func (h *tripHandler) update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	var body tripRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	t, err := h.trips.Update(r.Context(), userID, tripID, body.payload())
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, toTripJSON(t))
}

// DELETE /trips/{id} — удаление поездки → 204.
func (h *tripHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	if err := h.trips.Delete(r.Context(), userID, tripID); err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var _ = strings.TrimSpace