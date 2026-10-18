package service

// Unit-тесты PhotoService (CHAT-07: A7-04).
// Моки PhotoRepo/TripRepo/StoragePort.

import (
	"context"
	"errors"
	"testing"
	"time"

	"mytrip/internal/domain"
)

type mockPhotoRepo struct {
	insertFn   func(ctx context.Context, p *domain.Photo) error
	listFn     func(ctx context.Context, tripID string) ([]*domain.Photo, error)
	deleteFn   func(ctx context.Context, tripID, id string) error
	inserted   *domain.Photo
	deletedIDs [][2]string // (tripID, id)
}

func (m *mockPhotoRepo) Insert(ctx context.Context, p *domain.Photo) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, p)
	}
	m.inserted = p
	return nil
}

func (m *mockPhotoRepo) ListByTrip(ctx context.Context, tripID string) ([]*domain.Photo, error) {
	if m.listFn != nil {
		return m.listFn(ctx, tripID)
	}
	return nil, nil
}

func (m *mockPhotoRepo) Delete(ctx context.Context, tripID, id string) error {
	m.deletedIDs = append(m.deletedIDs, [2]string{tripID, id})
	if m.deleteFn != nil {
		return m.deleteFn(ctx, tripID, id)
	}
	return nil
}

type mockStorage struct {
	presignFn   func(ctx context.Context, objectKey string) (string, time.Time, error)
	deleteFn    func(ctx context.Context, objectKey string) error
	presigned   []string
	deletedKeys []string
}

func (m *mockStorage) PresignPut(ctx context.Context, objectKey string) (string, time.Time, error) {
	m.presigned = append(m.presigned, objectKey)
	if m.presignFn != nil {
		return m.presignFn(ctx, objectKey)
	}
	return "https://s3.example/presigned/" + objectKey, time.Now().Add(15*time.Minute), nil
}

func (m *mockStorage) DeleteObject(ctx context.Context, objectKey string) error {
	m.deletedKeys = append(m.deletedKeys, objectKey)
	if m.deleteFn != nil {
		return m.deleteFn(ctx, objectKey)
	}
	return nil
}

var _ PhotoService = (*photoService)(nil)

func newPhotoTestService(trips *mockTripRepo, photos *mockPhotoRepo, storage StoragePort) PhotoService {
	return NewPhotoService(trips, photos, storage)
}

func newTripRepoWithTrip(trip *domain.Trip) *mockTripRepo {
	return &mockTripRepo{getFn: func(context.Context, string, string) (*domain.Trip, error) {
		return trip, nil
	}}
}

// --- List ---

func TestPhotoService_List_Success(t *testing.T) {
	photos := []*domain.Photo{{ID: "p1", TripID: "t1", URL: "https://s3/p1.jpg", CreatedAt: time.Now()}}
	repo := &mockPhotoRepo{listFn: func(_ context.Context, tripID string) ([]*domain.Photo, error) {
		if tripID != "t1" {
			t.Fatalf("wrong tripID %q", tripID)
		}
		return photos, nil
	}}
	s := newPhotoTestService(newTripRepoWithTrip(newTestTrip(t)), repo, nil)
	got, err := s.List(context.Background(), "u1", "t1")
	if err != nil || len(got) != 1 || got[0].ID != "p1" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestPhotoService_List_TripNotFound(t *testing.T) {
	s := newPhotoTestService(&mockTripRepo{}, &mockPhotoRepo{}, nil)
	_, err := s.List(context.Background(), "u1", "missing")
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}

// --- Create ---

func TestPhotoService_Create_Success(t *testing.T) {
	photos := &mockPhotoRepo{}
	s := newPhotoTestService(newTripRepoWithTrip(newTestTrip(t)), photos, nil)
	created, err := s.Create(context.Background(), "u1", "t1", "trip/t1/p1.jpg", "https://s3/p1.jpg", false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if created.ID == "" || created.TripID != "t1" || created.URL != "https://s3/p1.jpg" {
		t.Fatalf("wrong photo: %+v", created)
	}
	if photos.inserted == nil || photos.inserted.ID != created.ID {
		t.Fatalf("repo.inserted mismatch: %+v", photos.inserted)
	}
}

func TestPhotoService_Create_CoverSetsTripCover(t *testing.T) {
	var coverCalled bool
	var coverURL string
	trips := &mockTripRepo{
		getFn: func(context.Context, string, string) (*domain.Trip, error) {
			return newTestTrip(t), nil
		},
		setCoverFn: func(_ context.Context, userID, tripID, url string) error {
			if userID != "u1" || tripID != "t1" {
				t.Fatalf("wrong setCover args: %q %q", userID, tripID)
			}
			coverCalled = true
			coverURL = url
			return nil
		},
	}
	s := newPhotoTestService(trips, &mockPhotoRepo{}, nil)
	_, err := s.Create(context.Background(), "u1", "t1", "k", "https://s3/c.jpg", true)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !coverCalled || coverURL != "https://s3/c.jpg" {
		t.Fatalf("SetCover not called correctly: called=%v url=%q", coverCalled, coverURL)
	}
}

func TestPhotoService_Create_NoCoverDoesNotCallSetCover(t *testing.T) {
	var coverCalled bool
	trips := &mockTripRepo{
		getFn: func(context.Context, string, string) (*domain.Trip, error) {
			return newTestTrip(t), nil
		},
		setCoverFn: func(context.Context, string, string, string) error {
			coverCalled = true
			return nil
		},
	}
	s := newPhotoTestService(trips, &mockPhotoRepo{}, nil)
	if _, err := s.Create(context.Background(), "u1", "t1", "k", "https://s3/x.jpg", false); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if coverCalled {
		t.Fatal("SetCover must not be called when cover=false")
	}
}

func TestPhotoService_Create_CoverSetCoverError(t *testing.T) {
	trips := &mockTripRepo{
		getFn: func(context.Context, string, string) (*domain.Trip, error) {
			return newTestTrip(t), nil
		},
		setCoverFn: func(context.Context, string, string, string) error {
			return domain.NewConflict("cover conflict")
		},
	}
	photos := &mockPhotoRepo{}
	s := newPhotoTestService(trips, photos, nil)
	_, err := s.Create(context.Background(), "u1", "t1", "k", "url", true)
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if photos.inserted != nil {
		t.Fatal("Insert must not be called when SetCover fails")
	}
}

func TestPhotoService_Create_TripNotFound(t *testing.T) {
	photos := &mockPhotoRepo{}
	s := newPhotoTestService(&mockTripRepo{}, photos, nil)
	_, err := s.Create(context.Background(), "u1", "missing", "k", "url", false)
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
	if photos.inserted != nil {
		t.Fatal("Insert must not be called when trip is not found")
	}
}

func TestPhotoService_Create_InsertError(t *testing.T) {
	repoErr := errors.New("boom")
	photos := &mockPhotoRepo{insertFn: func(context.Context, *domain.Photo) error { return repoErr }}
	s := newPhotoTestService(newTripRepoWithTrip(newTestTrip(t)), photos, nil)
	if _, err := s.Create(context.Background(), "u1", "t1", "k", "url", false); !errors.Is(err, repoErr) {
		t.Fatalf("expected insert error to propagate, got %v", err)
	}
}

// --- Delete ---

func TestPhotoService_Delete_Success(t *testing.T) {
	storage := &mockStorage{}
	photos := &mockPhotoRepo{listFn: func(context.Context, string) ([]*domain.Photo, error) {
		return []*domain.Photo{{ID: "p1", TripID: "t1", URL: "https://s3/t1/p1.jpg"}}, nil
	}}
	s := newPhotoTestService(newTripRepoWithTrip(newTestTrip(t)), photos, storage)
	if err := s.Delete(context.Background(), "u1", "t1", "p1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(photos.deletedIDs) != 1 || photos.deletedIDs[0] != (][2]string{"t1", "p1"}) {
		t.Fatalf("wrong delete calls: %v", photos.deletedIDs)
	}
	if len(storage.deletedKeys) != 1 || storage.deletedKeys[0] != "t1/p1.jpg" {
		t.Fatalf("storage delete not called with object key: %v", storage.deletedKeys)
	}
}

func TestPhotoService_Delete_NotFound(t *testing.T) {
	photos := &mockPhotoRepo{listFn: func(context.Context, string) ([]*domain.Photo, error) {
		return []*domain.Photo{{ID: "other", TripID: "t1", URL: "u"}}, nil
	}}
	s := newPhotoTestService(newTripRepoWithTrip(newTestTrip(t)), photos, nil)
	err := s.Delete(context.Background(), "u1", "t1", "p1")
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}

func TestPhotoService_Delete_TripNotFound(t *testing.T) {
	photos := &mockPhotoRepo{}
	s := newPhotoTestService(&mockTripRepo{}, photos, nil)
	err := s.Delete(context.Background(), "u1", "missing", "p1")
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
	if len(photos.deletedIDs) != 0 {
		t.Fatal("Delete must not be called when trip is not found")
	}
}

func TestPhotoService_Delete_StorageErrorIgnored(t *testing.T) {
	repoErr := errors.New("storage down")
	storage := &mockStorage{deleteFn: func(context.Context, string) error { return repoErr }}
	photos := &mockPhotoRepo{listFn: func(context.Context, string) ([]*domain.Photo, error) {
		return []*domain.Photo{{ID: "p1", TripID: "t1", URL: "https://s3/t1/p1.jpg"}}, nil
	}}
	s := newPhotoTestService(newTripRepoWithTrip(newTestTrip(t)), photos, storage)
	if err := s.Delete(context.Background(), "u1", "t1", "p1"); err != nil {
		t.Fatalf("storage error must not block record delete, got %v", err)
	}
}

// --- objectKeyFromURL ---

func TestObjectKeyFromURL(t *testing.T) {
	tests := []struct{ raw, want string }{
		{"https://s3.example.com/trips/t1/p1.jpg", "trips/t1/p1.jpg"},
		{"https://s3.example.com", ""},
		{"https://s3.example.com/", ""},
		{"https://minio:9000/bucket/trip/abc.jpg?X-Amz-Sig=1", "trip/abc.jpg?X-Amz-Sig=1"},
	}
	for _, tc := range tests {
		if got := objectKeyFromURL(tc.raw); got != tc.want {
			t.Errorf("objectKeyFromURL(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}
