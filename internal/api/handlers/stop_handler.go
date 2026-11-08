package handlers

// StopHandler — CRUD /trips/{id}/stops (CHAT-07: A7-03).
// Контракт: internal/api/handlers/handlers.go (StopHandler, StopJSON).
// Маршруты (относительно /api/v1): GET/POST /trips/{tripId}/stops,
// GET/PUT/DELETE /trips/{tripId}/stops/{stopId}.
// JSON — через middleware.WriteJSON; ошибки — через middleware.WriteError.

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/domain"
	"mytrip/internal/service"
)

// stopHandler — реализация StopHandler.
type stopHandler struct {
	stops service.StopService
}

// stopHandlers — реестр созданных экземпляров: ключ *StopHandler (identity),
// значение — внутренняя реализация.
var stopHandlers = make(map[*StopHandler]*stopHandler)

// NewStopHandler — конструктор StopHandler.
func NewStopHandler(stops service.StopService) *StopHandler {
	h := &StopHandler{}
	stopHandlers[h] = &stopHandler{stops: stops}
	return h
}

// Mount — регистрация маршрутов на chi router (путь относительно /api/v1).
func (h *StopHandler) Mount(r chi.Router) {
	impl := h.impl()
	r.Get("/trips/{tripId}/stops", impl.list)
	r.Post("/trips/{tripId}/stops", impl.create)
	r.Get("/trips/{tripId}/stops/{stopId}", impl.get)
	r.Put("/trips/{tripId}/stops/{stopId}", impl.update)
	r.Delete("/trips/{tripId}/stops/{stopId}", impl.delete)
}

// impl — приватный доступ к реализации по identity-ключу.
func (h *StopHandler) impl() *stopHandler {
	if impl, ok := stopHandlers[h]; ok {
		return impl
	}
	return &stopHandler{stops: nil}
}

// toStopJSON — domain.Stop → StopJSON.
func toStopJSON(s *domain.Stop) StopJSON {
	return StopJSON{
		ID:          s.ID,
		TripID:      s.TripID,
		Title:       s.Title,
		Description: s.Description,
		Location: StopLoc{
			Lat: s.Lat,
			Lng: s.Lng,
		},
		VisitDate:  s.VisitDate,
		Order:      s.Order,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.CreatedAt, // trip_stops has no updated_at
	}
}

// parseTime — парсинг даты из строки RFC3339.
func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}

type stopRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	VisitDate   string `json:"visit_date"`
	Order       int    `json:"order"`
}

// GET /trips/{tripId}/stops — список точек маршрута.
func (h *stopHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "tripId")
	stops, err := h.stops.List(r.Context(), userID, tripID)
	if err != nil {
		middleware.WriteError(w, middleware.ToAppError(err))
		return
	}
	if stops == nil {
		stops = []*domain.Stop{}
	}
	out := make([]StopJSON, 0, len(stops))
	for _, s := range stops {
		out = append(out, toStopJSON(s))
	}
	middleware.WriteJSON(w, http.StatusOK, out)
}

// POST /trips/{tripId}/stops — создание точки маршрута.
func (h *stopHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "tripId")
	var body stopRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	if body.VisitDate == "" {
		middleware.WriteError(w, domain.NewValidation("visit_date is required", map[string]any{"field": "visit_date"}))
		return
	}
	s, err := h.stops.Create(r.Context(), userID, tripID, &service.StopPayload{
		Title:       body.Title,
		Description: body.Description,
		Lat:         body.Lat,
		Lng:         body.Lng,
		VisitDate:   parseTime(body.VisitDate),
		Order:       body.Order,
	})
	if err != nil {
		middleware.WriteError(w, middleware.ToAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, toStopJSON(s))
}

// GET /trips/{tripId}/stops/{stopId} — точка по id.
func (h *stopHandler) get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "tripId")
	stopID := chi.URLParam(r, "stopId")
	s, err := h.stops.Get(r.Context(), userID, tripID, stopID)
	if err != nil {
		middleware.WriteError(w, middleware.ToAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, toStopJSON(s))
}

// PUT /trips/{tripId}/stops/{stopId} — обновление точки маршрута.
func (h *stopHandler) update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "tripId")
	stopID := chi.URLParam(r, "stopId")
	var body stopRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	if body.VisitDate == "" {
		middleware.WriteError(w, domain.NewValidation("visit_date is required", map[string]any{"field": "visit_date"}))
		return
	}
	s, err := h.stops.Update(r.Context(), userID, tripID, stopID, &service.StopPayload{
		Title:       body.Title,
		Description: body.Description,
		Lat:         body.Lat,
		Lng:         body.Lng,
		VisitDate:   parseTime(body.VisitDate),
		Order:       body.Order,
	})
	if err != nil {
		middleware.WriteError(w, middleware.ToAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, toStopJSON(s))
}

// DELETE /trips/{tripId}/stops/{stopId} — удаление точки маршрута → 204.
func (h *stopHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "tripId")
	stopID := chi.URLParam(r, "stopId")
	if err := h.stops.Delete(r.Context(), userID, tripID, stopID); err != nil {
		middleware.WriteError(w, middleware.ToAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
