package handlers

// Unit-тесты StopHandler (CHAT-07: A7-03).
// Контракт: internal/api/handlers/handlers.go (StopHandler, StopJSON, StopLoc).
// Моки: in-memory service.StopService.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/domain"
	"mytrip/internal/service"
)

// --- мок StopService ---

type mockStopService struct {
	listResult []*domain.Stop
	listErr    error
	getStop    *domain.Stop
	getErr     error
	createStop *domain.Stop
	createErr  error
	updateStop *domain.Stop
	updateErr  error
	deleteErr  error
	calls       []string
}

func (m *mockStopService) List(ctx context.Context, userID, tripID string) ([]*domain.Stop, error) {
	m.calls = append(m.calls, "List")
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.listResult == nil {
		return nil, nil
	}
	return m.listResult, nil
}

func (m *mockStopService) Get(ctx context.Context, userID, tripID, stopID string) (*domain.Stop, error) {
	m.calls = append(m.calls, "Get")
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getStop, nil
}

func (m *mockStopService) Create(ctx context.Context, userID, tripID string, p *service.StopPayload) (*domain.Stop, error) {
	m.calls = append(m.calls, "Create")
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createStop, nil
}

func (m *mockStopService) Update(ctx context.Context, userID, tripID, stopID string, p *service.StopPayload) (*domain.Stop, error) {
	m.calls = append(m.calls, "Update")
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return m.updateStop, nil
}

func (m *mockStopService) Delete(ctx context.Context, userID, tripID, stopID string) error {
	m.calls = append(m.calls, "Delete")
	return m.deleteErr
}

var _ service.StopService = (*mockStopService)(nil)

// newStopTestRouter — роутер с маршрутами StopHandler и user_id в контексте.
func newStopTestRouter(svc *mockStopService, userID string) chi.Router {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	NewStopHandler(svc).Mount(r)
	return r
}

func TestStopHandler_List_OK(t *testing.T) {
	m := &mockStopService{listResult: []*domain.Stop{{
		ID: "s1", TripID: "t1", Title: "a", Lat: 1.5, Lng: 2.5, Order: 1,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}}}
	r := newStopTestRouter(m, "u1")

	req := httptest.NewRequest(http.MethodGet, "/trips/t1/stops", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var out []StopJSON
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 || out[0].ID != "s1" || out[0].Location.Lat != 1.5 || out[0].Location.Lng != 2.5 {
		t.Fatalf("unexpected body: %+v", out)
	}
}

func TestStopHandler_List_Empty(t *testing.T) {
	m := &mockStopService{listResult: nil}
	r := newStopTestRouter(m, "u1")

	req := httptest.NewRequest(http.MethodGet, "/trips/t1/stops", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "[]") {
		t.Fatalf("want empty array, got %s", body)
	}
}

func TestStopHandler_List_NotFound(t *testing.T) {
	m := &mockStopService{listErr: domain.NewNotFound("trip not found")}
	r := newStopTestRouter(m, "u1")

	req := httptest.NewRequest(http.MethodGet, "/trips/t1/stops", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestStopHandler_Create_OK(t *testing.T) {
	m := &mockStopService{createStop: &domain.Stop{ID: "s1", TripID: "t1", Title: "a", Lat: 1, Lng: 2, Order: 1}}
	r := newStopTestRouter(m, "u1")

	body := `{"title":"a","description":"d","location":{"lat":1,"lng":2},"visit_date":"2026-01-01T00:00:00Z","order":1}`
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/stops", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d (%s)", rec.Code, rec.Body.String())
	}
	var out StopJSON
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ID != "s1" || out.Location.Lat != 1 || out.Location.Lng != 2 {
		t.Fatalf("unexpected body: %+v", out)
	}
}

func TestStopHandler_Create_ValidationErr(t *testing.T) {
	m := &mockStopService{createErr: domain.NewValidation("bad", nil)}
	r := newStopTestRouter(m, "u1")

	body := `{"title":"","location":{"lat":0,"lng":0}}`
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/stops", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestStopHandler_Create_InvalidJSON(t *testing.T) {
	m := &mockStopService{}
	r := newStopTestRouter(m, "u1")

	req := httptest.NewRequest(http.MethodPost, "/trips/t1/stops", strings.NewReader("{invalid"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestStopHandler_Get_OK(t *testing.T) {
	m := &mockStopService{getStop: &domain.Stop{ID: "s1", TripID: "t1", Title: "a", Lat: 1, Lng: 2, Order: 1}}
	r := newStopTestRouter(m, "u1")

	req := httptest.NewRequest(http.MethodGet, "/trips/t1/stops/s1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var out StopJSON
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ID != "s1" || out.TripID != "t1" {
		t.Fatalf("unexpected body: %+v", out)
	}
}

func TestStopHandler_Update_OK(t *testing.T) {
	m := &mockStopService{updateStop: &domain.Stop{ID: "s1", TripID: "t1", Title: "new", Lat: 1, Lng: 2, Order: 3}}
	r := newStopTestRouter(m, "u1")

	body := `{"title":"new","location":{"lat":1,"lng":2},"order":3}`
	req := httptest.NewRequest(http.MethodPut, "/trips/t1/stops/s1", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var out StopJSON
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Title != "new" || out.Order != 3 {
		t.Fatalf("unexpected body: %+v", out)
	}
}

func TestStopHandler_Delete_OK(t *testing.T) {
	m := &mockStopService{deleteErr: nil}
	r := newStopTestRouter(m, "u1")

	req := httptest.NewRequest(http.MethodDelete, "/trips/t1/stops/s1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
}

func TestStopHandler_Delete_NotFound(t *testing.T) {
	m := &mockStopService{deleteErr: domain.NewNotFound("stop not found")}
	r := newStopTestRouter(m, "u1")

	req := httptest.NewRequest(http.MethodDelete, "/trips/t1/stops/missing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

// TestStopHandler_InternalError — мок возвращает ошибку не из *AppError;
// handler конвертирует её в 500.
func TestStopHandler_InternalError(t *testing.T) {
	m := &mockStopService{listErr: errors.New("boom")}
	r := newStopTestRouter(m, "u1")

	req := httptest.NewRequest(http.MethodGet, "/trips/t1/stops", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}
