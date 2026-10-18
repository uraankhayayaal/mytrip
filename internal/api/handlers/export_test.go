package handlers

// Тесты ExportHandler — GET /trips/{id}/export (CHAT-07: A7-06).
// Контракт: internal/api/handlers/handlers.go (ExportHandler) + export.go:
//   - 200, Content-Type: text/csv, Content-Disposition: attachment; filename="<slug>.csv"
//   - строка 1: "order,title,description,lat,lng,visit_date"
//   - visit_date — RFC3339; CSV-экранирование полей (",", "\n", "\r", "\"").
//
// Мок TripService — переиспользует mockTripService из trip_handler_test.go
// (один пакет handlers); мок StopService — локальный, без сети и БД.

import (
	"context"
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

type mockStopSvc struct {
	listErr error
	stops   []*domain.Stop
}

func (m *mockStopSvc) Create(ctx context.Context, userID, tripID string, p *service.StopPayload) (*domain.Stop, error) {
	return nil, nil
}
func (m *mockStopSvc) Get(ctx context.Context, userID, tripID, stopID string) (*domain.Stop, error) {
	return nil, domain.NewNotFound("stop not found")
}
func (m *mockStopSvc) List(ctx context.Context, userID, tripID string) ([]*domain.Stop, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.stops, nil
}
func (m *mockStopSvc) Update(ctx context.Context, userID, tripID, stopID string, p *service.StopPayload) (*domain.Stop, error) {
	return nil, domain.NewNotFound("stop not found")
}
func (m *mockStopSvc) Delete(ctx context.Context, userID, tripID, stopID string) error {
	return domain.NewNotFound("stop not found")
}

var _ service.StopService = (*mockStopSvc)(nil)

// newExportTestRouter — роутер с маршрутом export и user_id в контексте.
func newExportTestRouter(tripSvc service.TripService, stopSvc service.StopService, userID string) chi.Router {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	NewExportHandler(tripSvc, stopSvc).Mount(r)
	return r
}

func TestExport_Success(t *testing.T) {
	when := time.Date(2026, 7, 15, 10, 30, 0, 0, time.UTC)
	trip := &domain.Trip{ID: "trip-1", Title: "Trip Title"}
	stops := []*domain.Stop{
		{ID: "s1", TripID: "trip-1", Title: "Stop A", Description: "desc", Lat: 55.75, Lng: 37.61, VisitDate: when, Order: 1},
		{ID: "s2", TripID: "trip-1", Title: "Stop B", Description: "second stop", Lat: 50.0, Lng: 30.0, VisitDate: when.Add(24 * time.Hour), Order: 2},
	}
	svc := &mockTripService{getFn: func(_ context.Context, userID, id string) (*domain.Trip, error) {
		if userID != "user-1" || id != "trip-1" {
			t.Fatalf("wrong args: %q %q", userID, id)
		}
		return trip, nil
	}}
	r := newExportTestRouter(svc, &mockStopSvc{stops: stops}, "user-1")

	req := httptest.NewRequest(http.MethodGet, "/trips/trip-1/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("Content-Type = %q, want prefix text/csv", ct)
	}
	cd := w.Header().Get("Content-Disposition")
	wantCD := `attachment; filename="trip-title.csv"`
	if cd != wantCD {
		t.Errorf("Content-Disposition = %q, want %q", cd, wantCD)
	}

	// Контрактный CSV: header + строки остановок; координаты — 6 знаков,
	// visit_date — RFC3339, поле с пробелом — в кавычках (csvQuote включает
	// пробел в набор символов, требующих кавычек).
	want := "order,title,description,lat,lng,visit_date\n" +
		"1,Stop A,desc,55.750000,37.610000,2026-07-15T10:30:00Z\n" +
		"2,Stop B,\"second stop\",50.000000,30.000000,2026-07-16T10:30:00Z\n"
	if body := w.Body.String(); body != want {
		t.Errorf("body mismatch:\n got: %q\nwant: %q", body, want)
	}
}

func TestExport_EmptyTitle_DefaultFilename(t *testing.T) {
	trip := &domain.Trip{ID: "trip-1", Title: "   "}
	svc := &mockTripService{getFn: func(_ context.Context, _, _ string) (*domain.Trip, error) {
		return trip, nil
	}}
	r := newExportTestRouter(svc, &mockStopSvc{stops: nil}, "user-1")

	req := httptest.NewRequest(http.MethodGet, "/trips/trip-1/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	cd := w.Header().Get("Content-Disposition")
	wantCD := `attachment; filename="trip.csv"`
	if cd != wantCD {
		t.Errorf("Content-Disposition = %q, want %q", cd, wantCD)
	}
	body := w.Body.String()
	if body != "order,title,description,lat,lng,visit_date\n" {
		t.Errorf("body = %q, want only header for empty stops", body)
	}
}

func TestExport_NotFound(t *testing.T) {
	svc := &mockTripService{} // без getFn → NewNotFound по умолчанию
	r := newExportTestRouter(svc, &mockStopSvc{}, "user-1")

	req := httptest.NewRequest(http.MethodGet, "/trips/nope/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestExport_StopListError_500(t *testing.T) {
	trip := &domain.Trip{ID: "trip-1"}
	svc := &mockTripService{getFn: func(_ context.Context, _, _ string) (*domain.Trip, error) {
		return trip, nil
	}}
	r := newExportTestRouter(svc, &mockStopSvc{listErr: domain.NewInternal()}, "user-1")

	req := httptest.NewRequest(http.MethodGet, "/trips/trip-1/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", w.Code, w.Body.String())
	}
}

func TestExport_CsvQuoting(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	stops := []*domain.Stop{
		{ID: "s1", TripID: "t", Title: `a,b"c`, Description: "line1\nline2", Lat: 1.5, Lng: 2.5, VisitDate: when, Order: 1},
	}
	trip := &domain.Trip{ID: "t", Title: "T"}
	svc := &mockTripService{getFn: func(_ context.Context, _, _ string) (*domain.Trip, error) {
		return trip, nil
	}}
	r := newExportTestRouter(svc, &mockStopSvc{stops: stops}, "u")
	req := httptest.NewRequest(http.MethodGet, "/trips/t/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	want := "order,title,description,lat,lng,visit_date\n" +
		"1,\"a,b\"\"c\",\"line1\nline2\",1.500000,2.500000,2026-01-02T03:04:05Z\n"
	if body := w.Body.String(); body != want {
		t.Errorf("body mismatch:\n got: %q\nwant: %q", body, want)
	}
}

func TestExport_RouteRegistered(t *testing.T) {
	// Mount регистрирует GET /trips/{id}/export: другой метод → 405, другой путь → 404.
	trip := &domain.Trip{ID: "x"}
	svc := &mockTripService{getFn: func(_ context.Context, _, _ string) (*domain.Trip, error) {
		return trip, nil
	}}
	r := newExportTestRouter(svc, &mockStopSvc{}, "u")

	delReq := httptest.NewRequest(http.MethodDelete, "/trips/x/export", nil)
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusMethodNotAllowed && delW.Code != http.StatusNotFound {
		t.Errorf("DELETE /trips/x/export = %d, want 405/404", delW.Code)
	}

	otherReq := httptest.NewRequest(http.MethodGet, "/other", nil)
	otherW := httptest.NewRecorder()
	r.ServeHTTP(otherW, otherReq)
	if otherW.Code != http.StatusNotFound {
		t.Errorf("GET /other = %d, want 404", otherW.Code)
	}
}
