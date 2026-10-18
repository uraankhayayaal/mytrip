package handlers

// StopHandler — CRUD /trips/{id}/stops (CHAT-07: A7-03).
// Контракт: internal/api/handlers/handlers.go (StopHandler, StopJSON, StopLoc).
// Маршруты (относительно /api/v1, под auth-middleware):
//
//	GET/POST /trips/{id}/stops
//	GET/PUT/DELETE /trips/{id}/stops/{stopId}
//
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

// stopHandler — реализация StopHandler (см. tripHandler — паттерн identity-реестра).
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
// Роуты вешаются ПОД auth-middleware (см. cmd/server/main.go).
func (h *StopHandler) Mount(r chi.Router) {
	impl := h.impl()
	r.Get("/trips/{id}/stops", impl.list)
	r.Post("/trips/{id}/stops", impl.create)
	r.Get("/trips/{id}/stops/{stopId}", impl.get)
	r.Put("/trips/{id}/stops/{stopId}", impl.update)
	r.Delete("/trips/{id}/stops/{stopId}", impl.delete)
}

// impl — приватный доступ к реализации по identity-ключу.
func (h *StopHandler) impl() *stopHandler {
	if impl, ok := stopHandlers[h]; ok {
		return impl
	}
	return &stopHandler{stops: nil}
}

// toStopJSON — маппинг domain.Stop → StopJSON (location:{lat,lng}).
func toStopJSON(s *domain.Stop) StopJSON {
	return StopJSON{
		ID:          s.ID,
		TripID:      s.TripID,
		Title:       s.Title,
		Description: s.Description,
		Location:    StopLoc{Lat: s.Lat, Lng: s.Lng},
		VisitDate:   s.VisitDate,
		Order:       s.Order,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.CreatedAt, // domain.Stop не имеет UpdatedAt — используем CreatedAt как fallback
	}
}

// stopRequest — DTO запроса на границе API (location:{lat,lng}).
type stopRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Location    StopLoc `json:"location"`
	VisitDate   time.Time `json:"visit_date"`
	Order       int `json:"order"`
}

func (b stopRequest) payload() *service.StopPayload {
	return &service.StopPayload{
		Title:       b.Title,
		Description: b.Description,
		Lat:         b.Location.Lat,
		Lng:         b.Location.Lng,
		VisitDate:   b.VisitDate,
		Order:       b.Order,
	}
}

// GET /trips/{id}/stops — список точек маршрута поездки.
func (h *stopHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	stops, err := h.stops.List(r.Context(), userID, tripID)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
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

// POST /trips/{id}/stops — создание точки маршрута.
func (h *stopHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	var body stopRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	s, err := h.stops.Create(r.Context(), userID, tripID, body.payload())
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, toStopJSON(s))
}

// GET /trips/{id}/stops/{stopId} — точка по id.
func (h *stopHandler) get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	stopID := chi.URLParam(r, "stopId")
	s, err := h.stops.Get(r.Context(), userID, tripID, stopID)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, toStopJSON(s))
}

// PUT /trips/{id}/stops/{stopId} — обновление точки (смена order → Reindex).
func (h *stopHandler) update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	stopID := chi.URLParam(r, "stopId")
	var body stopRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, domain.NewValidation("invalid request body", nil))
		return
	}
	s, err := h.stops.Update(r.Context(), userID, tripID, stopID, body.payload())
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, toStopJSON(s))
}

// DELETE /trips/{id}/stops/{stopId} — удаление точки → 204.
func (h *stopHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")
	stopID := chi.URLParam(r, "stopId")
	if err := h.stops.Delete(r.Context(), userID, tripID, stopID); err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
