package handlers

// Unit-тесты PhotoHandler (CHAT-07: A7-04).
// httptest + мок PhotoService/StoragePort.
// Маршруты: GET/POST /trips/{id}/photos, DELETE /trips/{id}/photos/{photoId},
// POST /trips/{id}/photos/presign.

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

type mockPhotoService struct {
	listFn   func(ctx context.Context, userID, tripID string) ([]*domain.Photo, error)
	createFn func(ctx context.Context, userID, tripID, objectKey, url string, cover bool) (*domain.Photo, error)
	deleteFn func(ctx context.Context, userID, tripID, photoID string) error
}

func (m *mockPhotoService) List(ctx context.Context, userID, tripID string) ([]*domain.Photo, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, tripID)
	}
	return nil, nil
}

func (m *mockPhotoService) Create(ctx context.Context, userID, tripID, objectKey, url string, cover bool) (*domain.Photo, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, tripID, objectKey, url, cover)
	}
	return nil, domain.NewNotFound("mock: create not configured")
}

func (m *mockPhotoService) Delete(ctx context.Context, userID, tripID, photoID string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, tripID, photoID)
	}
	return nil
}

var _ service.PhotoService = (*mockPhotoService)(nil)

type mockStoragePort struct {
	presignFn  func(ctx context.Context, objectKey string) (string, time.Time, error)
	deleteFn   func(ctx context.Context, objectKey string) error
	presigned  []string
	deletedKeys []string
}

func (m *mockStoragePort) PresignPut(ctx context.Context, objectKey string) (string, time.Time, error) {
	m.presigned = append(m.presigned, objectKey)
	if m.presignFn != nil {
		return m.presignFn(ctx, objectKey)
	}
	return "https://s3.example/presigned/" + objectKey, time.Now().Add(15*time.Minute), nil
}

func (m *mockStoragePort) DeleteObject(ctx context.Context, objectKey string) error {
	m.deletedKeys = append(m.deletedKeys, objectKey)
	if m.deleteFn != nil {
		return m.deleteFn(ctx, objectKey)
	}
	return nil
}

var _ service.StoragePort = (*mockStoragePort)(nil)

// newPhotoTestRouter — роутер с маршрутами PhotoHandler и user_id в контексте.
func newPhotoTestRouter(svc *mockPhotoService, storage service.StoragePort, userID string) chi.Router {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	NewPhotoHandler(svc, storage).Mount(r)
	return r
}

func samplePhoto() *domain.Photo {
	return &domain.Photo{
		ID:        "p1",
		TripID:    "t1",
		URL:       "https://s3.example/t1/p1.jpg",
		CreatedAt: time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC),
	}
}

// --- GET /trips/{id}/photos ---

func TestListPhotosHandler_Success(t *testing.T) {
	svc := &mockPhotoService{listFn: func(_ context.Context, userID, tripID string) ([]*domain.Photo, error) {
		if userID != "u1" || tripID != "t1" {
			t.Fatalf("wrong args: %q %q", userID, tripID)
		}
		return []*domain.Photo{samplePhoto()}, nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/trips/t1/photos", nil)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var photos []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &photos); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(photos) != 1 || photos[0]["id"] != "p1" || photos[0]["url"] != "https://s3.example/t1/p1.jpg" {
		t.Fatalf("wrong photos: %v", photos)
	}
}

func TestListPhotosHandler_Empty(t *testing.T) {
	svc := &mockPhotoService{listFn: func(context.Context, string, string) ([]*domain.Photo, error) {
		return nil, nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/trips/t1/photos", nil)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var photos []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &photos); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(photos) != 0 {
		t.Fatalf("expected empty array, got %v", photos)
	}
}

func TestListPhotosHandler_TripNotFound(t *testing.T) {
	svc := &mockPhotoService{listFn: func(context.Context, string, string) ([]*domain.Photo, error) {
		return nil, domain.NewNotFound("trip not found")
	}}
	req := httptest.NewRequest(http.MethodGet, "/trips/missing/photos", nil)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
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

// --- POST /trips/{id}/photos ---

func TestCreatePhotoHandler_Success(t *testing.T) {
	svc := &mockPhotoService{createFn: func(_ context.Context, userID, tripID, objectKey, url string, cover bool) (*domain.Photo, error) {
		if userID != "u1" || tripID != "t1" || objectKey != "t1/p1.jpg" || url != "https://s3/p1.jpg" || cover {
			t.Fatalf("wrong args: %q %q %q %q %v", userID, tripID, objectKey, url, cover)
		}
		return samplePhoto(), nil
	}}
	body := bytes.NewBufferString(`{"object_key":"t1/p1.jpg","url":"https://s3/p1.jpg"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/photos", body)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res["id"] != "p1" || res["url"] != "https://s3/p1.jpg" || res["trip_id"] != "t1" {
		t.Fatalf("wrong response: %v", res)
	}
}

func TestCreatePhotoHandler_CoverTrue(t *testing.T) {
	var gotCover bool
	svc := &mockPhotoService{createFn: func(_ context.Context, _ string, _ string, _, _ string, cover bool) (*domain.Photo, error) {
		gotCover = cover
		return samplePhoto(), nil
	}}
	body := bytes.NewBufferString(`{"object_key":"k","url":"u","cover":true}`)
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/photos", body)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if !gotCover {
		t.Fatal("cover=true must be passed to service")
	}
}

func TestCreatePhotoHandler_MissingFields(t *testing.T) {
	svc := &mockPhotoService{createFn: func(context.Context, string, string, string, string, bool) (*domain.Photo, error) {
		t.Fatal("service must not be called when body fields missing")
		return nil, nil
	}}
	body := bytes.NewBufferString(`{"object_key":"k"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/photos", body)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
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
		t.Fatalf("expected validation, got %q", res.Err.Code)
	}
}

func TestCreatePhotoHandler_BadJSON(t *testing.T) {
	svc := &mockPhotoService{}
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/photos", bytes.NewBufferString(`{"object_key":`))
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreatePhotoHandler_UnknownField(t *testing.T) {
	svc := &mockPhotoService{}
	body := bytes.NewBufferString(`{"object_key":"k","url":"u","bogus":1}`)
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/photos", body)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown field, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreatePhotoHandler_ServiceError(t *testing.T) {
	svc := &mockPhotoService{createFn: func(context.Context, string, string, string, string, bool) (*domain.Photo, error) {
		return nil, domain.NewNotFound("trip not found")
	}}
	body := bytes.NewBufferString(`{"object_key":"k","url":"u"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips/missing/photos", body)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- DELETE /trips/{id}/photos/{photoId} ---

func TestDeletePhotoHandler_Success(t *testing.T) {
	svc := &mockPhotoService{deleteFn: func(_ context.Context, userID, tripID, photoID string) error {
		if userID != "u1" || tripID != "t1" || photoID != "p1" {
			t.Fatalf("wrong args: %q %q %q", userID, tripID, photoID)
		}
		return nil
	}}
	req := httptest.NewRequest(http.MethodDelete, "/trips/t1/photos/p1", nil)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeletePhotoHandler_NotFound(t *testing.T) {
	svc := &mockPhotoService{deleteFn: func(context.Context, string, string, string) error {
		return domain.NewNotFound("photo not found")
	}}
	req := httptest.NewRequest(http.MethodDelete, "/trips/t1/photos/missing", nil)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, nil, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- POST /trips/{id}/photos/presign ---

func TestPresignHandler_Success(t *testing.T) {
	storage := &mockStoragePort{}
	svc := &mockPhotoService{listFn: func(context.Context, string, string) ([]*domain.Photo, error) {
		return nil, nil
	}}
	body := bytes.NewBufferString(`{"object_key":"t1/p1.jpg"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/photos/presign", body)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, storage, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		URL       string    `json:"url"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if res.URL != "https://s3.example/presigned/t1/p1.jpg" {
		t.Fatalf("wrong url: %q", res.URL)
	}
	if time.Since(res.ExpiresAt) < -14*time.Minute || time.Since(res.ExpiresAt) > -10*time.Minute {
		t.Fatalf("unexpected expires_at: %v", res.ExpiresAt)
	}
	if len(storage.presigned) != 1 || storage.presigned[0] != "t1/p1.jpg" {
		t.Fatalf("wrong presigned keys: %v", storage.presigned)
	}
}

func TestPresignHandler_TripNotFound(t *testing.T) {
	svc := &mockPhotoService{listFn: func(context.Context, string, string) ([]*domain.Photo, error) {
		return nil, domain.NewNotFound("trip not found")
	}}
	storage := &mockStoragePort{}
	body := bytes.NewBufferString(`{"object_key":"k"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips/missing/photos/presign", body)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, storage, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if len(storage.presigned) != 0 {
		t.Fatal("storage must not be called when trip is not found")
	}
}

func TestPresignHandler_MissingObjectKey(t *testing.T) {
	svc := &mockPhotoService{listFn: func(context.Context, string, string) ([]*domain.Photo, error) {
		return nil, nil
	}}
	storage := &mockStoragePort{}
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/photos/presign", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, storage, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if len(storage.presigned) != 0 {
		t.Fatal("storage must not be called when object_key missing")
	}
}

func TestPresignHandler_StorageError(t *testing.T) {
	storageErr := domain.NewInternal()
	svc := &mockPhotoService{listFn: func(context.Context, string, string) ([]*domain.Photo, error) {
		return nil, nil
	}}
	storage := &mockStoragePort{presignFn: func(context.Context, string) (string, time.Time, error) {
		return "", time.Time{}, storageErr
	}}
	body := bytes.NewBufferString(`{"object_key":"k"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips/t1/photos/presign", body)
	w := httptest.NewRecorder()
	newPhotoTestRouter(svc, storage, "u1").ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPhotoHandler_RoutesRegistered(t *testing.T) {
	svc := &mockPhotoService{}
	storage := &mockStoragePort{}
	r := newPhotoTestRouter(svc, storage, "u1")

	// Методы, которые должны существовать на маршрутах.
	cases := []struct {
		method string
		target string
		want   int
	}{
		{http.MethodGet, "/trips/t1/photos", http.StatusOK},
		{http.MethodPost, "/trips/t1/photos", http.StatusCreated},
		{http.MethodPost, "/trips/t1/photos/presign", http.StatusOK},
		{http.MethodDelete, "/trips/t1/photos/p1", http.StatusNoContent},
		{http.MethodGet, "/trips/other/photos", http.StatusNotFound}, // не зарегистрировано → 404
	}
	for _, tc := range cases {
		var req *http.Request
		if tc.method == http.MethodPost {
			req = httptest.NewRequest(tc.method, tc.target, bytes.NewBufferString(`{"object_key":"k","url":"u"}`))
		} else {
			req = httptest.NewRequest(tc.method, tc.target, nil)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s %s: expected %d, got %d: %s", tc.method, tc.target, tc.want, w.Code, w.Body.String())
		}
	}
}
