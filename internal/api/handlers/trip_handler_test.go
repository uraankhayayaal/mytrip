package handlers

// Unit-тесты TripHandler (CHAT-07: A7-02).
// httptest + мок TripService; маршруты — GET/POST /trips, GET/PUT/DELETE /trips/{id}.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	// ВАЖНО: не (nil, nil) — хендлер не обязан гасить nil-ответ без ошибки
	// (toTripJSON(nil) падает). Мок без createFn имитирует «поездка не найдена».
	return nil, domain.NewNotFound("mock: create not configured")
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
	var trips []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &trips); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(trips) != 0 {
		t.Fatalf("expected empty array, got %v", trips)
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

// --- Дополнительные QA-тесты (A7-02) ---

func TestListTripsHandler_ServiceError(t *testing.T) {
	svc := &mockTripService{listFn: func(context.Context, string) ([]*domain.Trip, error) {
		return nil, domain.NewInternal()
	}}
	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		Err struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.Err.Code != domain.ErrCodeInternal {
		t.Fatalf("expected internal code, got %q", res.Err.Code)
	}
}

func TestListTripsHandler_CoverPhotoURLAndDates(t *testing.T) {
	cover := "https://cdn.example/c.jpg"
	tr := sampleTrip()
	tr.CoverPhotoURL = &cover
	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	w := httptest.NewRecorder()
	svc := &mockTripService{listFn: func(context.Context, string) ([]*domain.Trip, error) {
		return []*domain.Trip{tr}, nil
	}}
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var trips []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &trips); err != nil {
		t.Fatalf("json: %v", err)
	}
	if trips[0]["cover_photo_url"] != cover {
		t.Fatalf("cover_photo_url missing: %v", trips[0])
	}
	if _, ok := trips[0]["start_date"]; !ok {
		t.Fatalf("start_date missing in JSON: %v", trips[0])
	}
	if _, ok := trips[0]["end_date"]; !ok {
		t.Fatalf("end_date missing in JSON: %v", trips[0])
	}
	if _, ok := trips[0]["created_at"]; !ok {
		t.Fatalf("created_at missing in JSON: %v", trips[0])
	}
}

func TestCreateTripHandler_CoverPhotoURLPassthrough(t *testing.T) {
	svc := &mockTripService{createFn: func(_ context.Context, _ string, p *service.TripPayload) (*domain.Trip, error) {
		if p.CoverPhotoURL == nil || *p.CoverPhotoURL != "https://cdn.example/x.jpg" {
			t.Fatalf("cover not passed: %+v", p.CoverPhotoURL)
		}
		return sampleTrip(), nil
	}}
	body := bytes.NewBufferString(`{"title":"T","start_date":"2026-05-01T00:00:00Z","end_date":"2026-05-02T00:00:00Z","cover_photo_url":"https://cdn.example/x.jpg"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateTripHandler_OmitCoverField(t *testing.T) {
	var got *service.TripPayload
	svc := &mockTripService{createFn: func(_ context.Context, _ string, p *service.TripPayload) (*domain.Trip, error) {
		got = p
		return sampleTrip(), nil
	}}
	body := bytes.NewBufferString(`{"title":"T","start_date":"2026-05-01T00:00:00Z","end_date":"2026-05-02T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if got.CoverPhotoURL != nil {
		t.Fatalf("expected nil cover, got %v", *got.CoverPhotoURL)
	}
}

func TestUpdateTripHandler_InvalidJSON(t *testing.T) {
	svc := &mockTripService{}
	req := httptest.NewRequest(http.MethodPut, "/trips/t1", bytes.NewBufferString(`{"title":`))
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		Err struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.Err.Code != domain.ErrCodeValidation {
		t.Fatalf("expected validation code, got %q", res.Err.Code)
	}
}

func TestUpdateTripHandler_ServiceValidation(t *testing.T) {
	svc := &mockTripService{updateFn: func(context.Context, string, string, *service.TripPayload) (*domain.Trip, error) {
		return nil, domain.NewValidation("end_date must be after start_date", nil)
	}}
	body := bytes.NewBufferString(`{"title":"x","start_date":"2026-05-02T00:00:00Z","end_date":"2026-05-01T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPut, "/trips/t1", body)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteTripHandler_ServiceError(t *testing.T) {
	svc := &mockTripService{deleteFn: func(context.Context, string, string) error {
		return domain.NewInternal()
	}}
	req := httptest.NewRequest(http.MethodDelete, "/trips/t1", nil)
	w := httptest.NewRecorder()
	newTripTestRouter(svc, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateTripHandler_SuccessWithCover(t *testing.T) {
	cover := "https://cdn.example/c2.jpg"
	svc := &mockTripService{updateFn: func(_ context.Context, userID, id string, p *service.TripPayload) (*domain.Trip, error) {
		if userID != "u1" || id != "t1" {
			t.Fatalf("wrong args: %q %q", userID, id)
		}
		if p.CoverPhotoURL == nil || *p.CoverPhotoURL != cover {
			t.Fatalf("cover not passed: %+v", p.CoverPhotoURL)
		}
		tr := sampleTrip()
		tr.Title = p.Title
		tr.CoverPhotoURL = p.CoverPhotoURL
		return tr, nil
	}}
	body := bytes.NewBufferString(`{"title":"With Cover","start_date":"2026-05-01T00:00:00Z","end_date":"2026-05-10T00:00:00Z","cover_photo_url":"https://cdn.example/c2.jpg"}`)
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
	if res["cover_photo_url"] != cover {
		t.Fatalf("cover missing in response: %v", res)
	}
}

func TestTripHandler_RoutesRegistered(t *testing.T) {
	// Проверка зарегистрированности маршрутов: для каждого метода+пути
	// хендлер ДОЛЖЕН отработать (ответ сервиса/404-от-сервиса), а не 405
	// (метод не зарегистрирован) или 404-от-роутера (путь не зарегистрирован).
	// Мок без fn-обработчиков возвращает NotFound для Get/Update — это
	// доказательство, что запрос дошёл до хендлера.
	r := newTripTestRouter(&mockTripService{}, "u1")
	validBody := `{"title":"t","start_date":"2026-05-01T00:00:00Z","end_date":"2026-05-02T00:00:00Z"}`
	cases := []struct {
		method, path string
		body         io.Reader
		want         int
	}{
		{http.MethodGet, "/trips", nil, http.StatusOK},
		{http.MethodPost, "/trips", bytes.NewBufferString(validBody), http.StatusNotFound},   // create → mock NotFound
		{http.MethodGet, "/trips/t1", nil, http.StatusNotFound},                              // get → mock NotFound
		{http.MethodPut, "/trips/t1", bytes.NewBufferString(validBody), http.StatusNotFound}, // update → mock NotFound
		{http.MethodDelete, "/trips/t1", nil, http.StatusNoContent},                          // delete → ok
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, c.body)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != c.want {
			t.Fatalf("%s %s: expected %d, got %d (маршрут/метод не зарегистрированы или хендлер не отработал)",
				c.method, c.path, c.want, w.Code)
		}
	}
}
