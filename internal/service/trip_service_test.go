package service

// Unit-тесты TripService (CHAT-07: A7-02).
// Мок TripRepo.

import (
	"context"
	"errors"
	"testing"
	"time"

	"mytrip/internal/domain"
)

type mockTripRepo struct {
	insertFn   func(ctx context.Context, t *domain.Trip) error
	listFn     func(ctx context.Context, userID string) ([]*domain.Trip, error)
	getFn      func(ctx context.Context, userID, id string) (*domain.Trip, error)
	updateFn   func(ctx context.Context, t *domain.Trip) error
	deleteFn   func(ctx context.Context, userID, id string) error
	setCoverFn func(ctx context.Context, userID, tripID, coverURL string) error

	inserted *domain.Trip // последний INSERT
	updated  *domain.Trip // последний UPDATE
}

func (m *mockTripRepo) Insert(ctx context.Context, t *domain.Trip) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, t)
	}
	m.inserted = t
	return nil
}

func (m *mockTripRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Trip, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockTripRepo) Get(ctx context.Context, userID, id string) (*domain.Trip, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID, id)
	}
	return nil, domain.NewNotFound("trip not found")
}

func (m *mockTripRepo) Update(ctx context.Context, t *domain.Trip) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, t)
	}
	m.updated = t
	return nil
}

func (m *mockTripRepo) Delete(ctx context.Context, userID, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, id)
	}
	return nil
}

func (m *mockTripRepo) SetCover(ctx context.Context, userID, tripID, coverURL string) error {
	if m.setCoverFn != nil {
		return m.setCoverFn(ctx, userID, tripID, coverURL)
	}
	return nil
}

var _ = (*domain.Trip)(nil)

func newTestTrip(t *testing.T) *domain.Trip {
	t.Helper()
	return &domain.Trip{
		ID:        "t1",
		UserID:    "u1",
		Title:     "Trip",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
	}
}

// --- validatePayload ---

func TestValidatePayload_BlankTitle(t *testing.T) {
	for _, title := range []string{"", "   "} {
		err := validatePayload(&TripPayload{Title: title,
			StartDate: time.Now(), EndDate: time.Now().Add(time.Hour)})
		var appErr *domain.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("expected *domain.AppError for %q, got %v", title, err)
		}
		if appErr.Code != domain.ErrCodeValidation {
			t.Fatalf("expected validation code, got %s", appErr.Code)
		}
	}
}

func TestValidatePayload_EndBeforeStart(t *testing.T) {
	err := validatePayload(&TripPayload{Title: "Trip",
		StartDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeValidation {
		t.Fatalf("expected validation, got %v", err)
	}
}

func TestValidatePayload_OK(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := validatePayload(&TripPayload{Title: "Trip",
		StartDate: start, EndDate: start}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// --- Create ---

func TestTripService_Create_Success(t *testing.T) {
	repo := &mockTripRepo{}
	s := NewTripService(repo)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := &TripPayload{Title: "  Trip  ", Description: "d",
		StartDate: start, EndDate: start.AddDate(0, 1, 0)}
	t, err := s.Create(context.Background(), "u1", p)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if t.ID == "" {
		t.Fatal("expected generated id")
	}
	if t.Title != "Trip" {
		t.Fatalf("expected trimmed title, got %q", t.Title)
	}
	if t.UserID != "u1" || repo.inserted == nil || repo.inserted.ID != t.ID {
		t.Fatalf("repo.inserted mismatch: %+v / %+v", repo.inserted, t)
	}

func TestTripService_Create_Validation(t *testing.T) {
	repo := &mockTripRepo{}
	s := NewTripService(repo)
	_, err := s.Create(context.Background(), "u1", &TripPayload{Title: " "})
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
	if repo.inserted != nil {
		t.Fatal("repo.Insert must not be called on validation error")
	}
}

func TestTripService_Create_RepoError(t *testing.T) {
	repo := &mockTripRepo{insertFn: func(_ context.Context, _ *domain.Trip) error {
		return errors.New("boom")
	}}
	s := NewTripService(repo)
	start := time.Now()
	if _, err := s.Create(context.Background(), "u1", &TripPayload{Title: "x", StartDate: start, EndDate: start}); err == nil {
		t.Fatal("expected error")
	}
}

// --- Get ---

func TestTripService_Get_NotFound(t *testing.T) {
	repo := &mockTripRepo{}
	s := NewTripService(repo)
	_, err := s.Get(context.Background(), "u1", "t1")
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}

func TestTripService_Get_Success(t *testing.T) {
	expected := newTestTrip(t)
	repo := &mockTripRepo{getFn: func(_ context.Context, _, id string) (*domain.Trip, error) {
		if id != expected.ID {
			t.Fatalf("wrong id %q", id)
		}
		return expected, nil
	}}
	s := NewTripService(repo)
	got, err := s.Get(context.Background(), "u1", "t1")
	if err != nil || got != expected {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

// --- Update ---

func TestTripService_Update_Success(t *testing.T) {
	current := newTestTrip(t)
	repo := &mockTripRepo{getFn: func(_ context.Context, _, _ string) (*domain.Trip, error) {
		return current, nil
	}}
	s := NewTripService(repo)
	newStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	p := &TripPayload{Title: "New", Description: "n",
		StartDate: newStart, EndDate: newStart.AddDate(0, 0, 5)}
	got, err := s.Update(context.Background(), "u1", "t1", p)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Title != "New" || repo.updated == nil || repo.updated.Title != "New" {
		t.Fatalf("update not applied: %+v / %+v", got, repo.updated)
	}
}

func TestTripService_Update_NotFound(t *testing.T) {
	repo := &mockTripRepo{}
	s := NewTripService(repo)
	start := time.Now()
	_, err := s.Update(context.Background(), "u1", "missing", &TripPayload{Title: "x", StartDate: start, EndDate: start})
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}

func TestTripService_Update_Validation(t *testing.T) {
	repo := &mockTripRepo{}
	s := NewTripService(repo)
	_, err := s.Update(context.Background(), "u1", "t1", &TripPayload{Title: ""})
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeValidation {
		t.Fatalf("expected validation, got %v", err)
	}
}

// --- Delete ---

func TestTripService_Delete_NotFound(t *testing.T) {
	repo := &mockTripRepo{deleteFn: func(_ context.Context, _, _ string) error {
		return domain.NewNotFound("trip not found")
	}}
	s := NewTripService(repo)
	err := s.Delete(context.Background(), "u1", "t1")
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}

// --- List ---

func TestTripService_List_Passthrough(t *testing.T) {
	repo := &mockTripRepo{listFn: func(_ context.Context, userID string) ([]*domain.Trip, error) {
		if userID != "u1" {
			t.Fatalf("wrong user %q", userID)
		}
		return []*domain.Trip{newTestTrip(t)}, nil
	}}
	s := NewTripService(repo)
	trips, err := s.List(context.Background(), "u1")
	if err != nil || len(trips) != 1 {
		t.Fatalf("got=%v err=%v", trips, err)
	}
}
