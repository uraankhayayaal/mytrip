package handlers

// Unit-тесты TripHandler (CHAT-07: A7-02).
// httptest + мок TripService; маршруты — GET/POST /trips, GET/PUT/DELETE /trips/{id}.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/domain"
	"mytrip/internal/service"
)

type mockTripService struct {
	listFn   func(ctx context.Context, userID string) ([]*domain.Trip, error)
	getFn    func(ctx context.Context, userID, tripID string) (*domain.Trip, error)
	createFn func(ctx context.Context, userID string, p *service.TripPayload) (*domain.Trip, error)
	updateFn func(ctx context.Context, userID, tripID string, p *service.TripPayload) (*domain.Trip, error)
	deleteFn func(ctx context.Context, userID, tripID string) error
}

func (m *mockTripService) List(ctx context.Context, userID string) ([]*domain.Trip, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockTripService) Get(ctx context.Context, userID, tripID string) (*domain.Trip, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID, tripID)
	}
	return nil, domain.NewNotFound("trip not found")
}

func (m *mockTripService) Create(ctx context.Context, userID string, p *service.TripPayload) (*domain.Trip, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, p)
	}
	return nil, nil
}

func (m *mockTripService) Update(ctx context.Context, userID, tripID string, p *service.TripPayload) (*domain.Trip, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, userID, tripID, p)
	}
	return nil, domain.NewNotFound("trip not found")
}

func (m *mockTripService) Delete(ctx context.Context, userID, tripID string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, tripID)
	}
	return nil
}

var _ service.TripService = (*mockTripService)(nil)

// newTripTestRouter — роутер с маршрутами TripHandler и user_id в контексте.
func newTripTestRouter(svc *mockTripService, userID string) chi.Router {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	NewTripHandler(svc).Mount(r)
	return r
}

func sampleTrip() *domain.Trip {
	return &domain.Trip{
		ID:          "t1",
		UserID:      "u1",
		Title:       "Trip",
		Description: "d",
		StartDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestListTripsHandler_Success(t *testing.T) {
	svc := &mockTripService{listFn: func(_ context.Context, userID string) ([]*domain.Trip, error) {
		if userID != "u1" {
			t.Fatalf("expected user u1, got %q", userID)
		}
		return []*domain.Trip{sampleTrip()}, nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var trips []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &trips); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(trips) != 1 || trips[0]["id"] != "t1" || trips[0]["title"] != "Trip" {
		t.Fatalf("wrong trips: %v", trips)
	}
}

func TestListTripsHandler_Empty(t *testing.T) {
	svc := &mockTripService{listFn: func(context.Context, string) ([]*domain.Trip, error) {
		return nil, nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if body != "[]\n" && body != "[\n]\n" {
		t.Fatalf("expected empty array, got %s", body)
	}
}

func TestCreateTripHandler_Success(t *testing.T) {
	svc := &mockTripService{createFn: func(_ context.Context, userID string, p *service.TripPayload) (*domain.Trip, error) {
		if userID != "u1" {
			t.Fatalf("expected user u1, got %q", userID)
		}
		if p.Title != "New Trip" {
			t.Fatalf("wrong payload title: %q", p.Title)
		}
		return &domain.Trip{ID: "t2", UserID: "u1", Title: "New Trip",
			StartDate: p.StartDate, EndDate: p.EndDate, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
	}}
	body := bytes.NewBufferString(`{"title":"New Trip","start_date":"2026-05-01T00:00:00Z","end_date":"2026-05-10T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res["id"] != "t2" || res["title"] != "New Trip" {
		t.Fatalf("wrong response: %v", res)
	}
}

func TestCreateTripHandler_Validation(t *testing.T) {
	svc := &mockTripService{createFn: func(context.Context, string, *service.TripPayload) (*domain.Trip, error) {
		return nil, domain.NewValidation("title is required", nil)
	}}
	body := bytes.NewBufferString(`{"title":""}`)
	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		Err struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.Err.Code != domain.ErrCodeValidation {
		t.Fatalf("expected validation code, got %q", res.Err.Code)
	}
}

func TestCreateTripHandler_BadJSON(t *testing.T) {
	svc := &mockTripService{}
	req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewBufferString(`{"title":`))
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetTripHandler_Success(t *testing.T) {
	svc := &mockTripService{getFn: func(_ context.Context, userID, id string) (*domain.Trip, error) {
		if userID != "u1" || id != "t1" {
			t.Fatalf("wrong args: %q %q", userID, id)
		}
		return sampleTrip(), nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/trips/t1", nil)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res["id"] != "t1" {
		t.Fatalf("wrong trip: %v", res)
	}
}

func TestGetTripHandler_NotFound(t *testing.T) {
	svc := &mockTripService{}
	req := httptest.NewRequest(http.MethodGet, "/trips/missing", nil)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		Err struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.Err.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %q", res.Err.Code)
	}
}

func TestUpdateTripHandler_Success(t *testing.T) {
	svc := &mockTripService{updateFn: func(_ context.Context, userID, id string, p *service.TripPayload) (*domain.Trip, error) {
		if userID != "u1" || id != "t1" {
			t.Fatalf("wrong args: %q %q", userID, id)
		}
		tr := sampleTrip()
		tr.Title = p.Title
		return tr, nil
	}}
	body := bytes.NewBufferString(`{"title":"Renamed","start_date":"2026-05-01T00:00:00Z","end_date":"2026-05-10T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPut, "/trips/t1", body)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res["title"] != "Renamed" {
		t.Fatalf("wrong response: %v", res)
	}
}

func TestUpdateTripHandler_NotFound(t *testing.T) {
	svc := &mockTripService{}
	body := bytes.NewBufferString(`{"title":"x","start_date":"2026-05-01T00:00:00Z","end_date":"2026-05-02T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPut, "/trips/missing", body)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteTripHandler_Success(t *testing.T) {
	svc := &mockTripService{deleteFn: func(_ context.Context, userID, id string) error {
		if userID != "u1" || id != "t1" {
			t.Fatalf("wrong args: %q %q", userID, id)
		}
		return nil
	}}
	req := httptest.NewRequest(http.MethodDelete, "/trips/t1", nil)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteTripHandler_NotFound(t *testing.T) {
	svc := &mockTripService{deleteFn: func(context.Context, string, string) error {
		return domain.NewNotFound("trip not found")
	}}
	req := httptest.NewRequest(http.MethodDelete, "/trips/missing", nil)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
