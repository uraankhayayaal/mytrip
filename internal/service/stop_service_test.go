package service

// Unit-тесты StopService (CHAT-07: A7-03).
// Контракт: internal/service/interfaces.go (StopService, StopPayload).
// Моки: in-memory repo.StopRepo и repo.TripRepo.

import (
	"context"
	"errors"
	"testing"
	"time"

	"mytrip/internal/domain"
)

// --- моки ---

type mockTripRepo struct {
	trips map[string]*domain.Trip // id → trip
}

func newMockTripRepo() *mockTripRepo {
	return &mockTripRepo{trips: map[string]*domain.Trip{}}
}

func (m *mockTripRepo) Insert(ctx context.Context, t *domain.Trip) error {
	if t.ID == "" {
		t.ID = "trip-id-1"
	}
	m.trips[t.ID] = t
	return nil
}

func (m *mockTripRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Trip, error) {
	var out []*domain.Trip
	for _, t := range m.trips {
		if t.UserID == userID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (m *mockTripRepo) Get(ctx context.Context, userID, id string) (*domain.Trip, error) {
	t, ok := m.trips[id]
	if !ok || t.UserID != userID {
		return nil, domain.NewNotFound("trip not found")
	}
	return t, nil
}

func (m *mockTripRepo) Update(ctx context.Context, t *domain.Trip) error {
	if _, ok := m.trips[t.ID]; !ok {
		return domain.NewNotFound("trip not found")
	}
	m.trips[t.ID] = t
	return nil
}

func (m *mockTripRepo) Delete(ctx context.Context, userID, id string) error {
	if _, ok := m.trips[id]; !ok {
		return domain.NewNotFound("trip not found")
	}
	delete(m.trips, id)
	return nil
}

func (m *mockTripRepo) SetCover(ctx context.Context, userID, tripID, coverURL string) error {
	t, ok := m.trips[tripID]
	if !ok || t.UserID != userID {
		return domain.NewNotFound("trip not found")
	}
	t.CoverPhotoURL = &coverURL
	return nil
}

type mockStopRepo struct {
	stops       map[string]*domain.Stop // id → stop
	maxOrder    int
	reindexLog  [][2]int // (from, to)
	updateCalls int
}

func newMockStopRepo() *mockStopRepo {
	return mockStopRepo{stops: map[string]*domain.Stop{}}
}

func (m *mockStopRepo) Insert(ctx context.Context, s *domain.Stop) error {
	if s.ID == "" {
		s.ID = "stop-id-1"
	}
	m.stops[s.ID] = s
	return nil
}

func (m *mockStopRepo) ListByTrip(ctx context.Context, tripID string) ([]*domain.Stop, error) {
	var out []*domain.Stop
	for _, s := range m.stops {
		if s.TripID == tripID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *mockStopRepo) Get(ctx context.Context, tripID, id string) (*domain.Stop, error) {
	s, ok := m.stops[id]
	if !ok || s.TripID != tripID {
		return nil, domain.NewNotFound("stop not found")
	}
	return s, nil
}

func (m *mockStopRepo) Update(ctx context.Context, s *domain.Stop) error {
	m.updateCalls++
	if _, ok := m.stops[s.ID]; !ok {
		return domain.NewNotFound("stop not found")
	}
	m.stops[s.ID] = s
	return nil
}

func (m *mockStopRepo) Delete(ctx context.Context, tripID, id string) error {
	if _, ok := m.stops[id]; !ok {
		return domain.NewNotFound("stop not found")
	}
	delete(m.stops, id)
	return nil
}

func (m *mockStopRepo) MaxOrder(ctx context.Context, tripID string) (int, error) {
	return m.maxOrder, nil
}

func (m *mockStopRepo) Reindex(ctx context.Context, tripID string, from, to int) error {
	m.reindexLog = append(m.reindexLog, [2]int{from, to})
	return nil
}

// --- тесты ---

func TestStopService_Create_Validation(t *testing.T) {
	tc := []struct {
		name    string
		payload *StopPayload
		wantErr string // код ошибки, пустое — без ошибки
	}{
		{name: "empty title", payload: &StopPayload{Title: "  ", Lat: 1, Lng: 2}, wantErr: domain.ErrCodeValidation},
		{name: "lat out of range", payload: &StopPayload{Title: "a", Lat: 91, Lng: 0}, wantErr: domain.ErrCodeValidation},
		{name: "lat -91 out of range", payload: &StopPayload{Title: "a", Lat: -91, Lng: 0}, wantErr: domain.ErrCodeValidation},
		{name: "lng out of range", payload: &StopPayload{Title: "a", Lat: 0, Lng: 181}, wantErr: domain.ErrCodeValidation},
		{name: "lng -181 out of range", payload: &StopPayload{Title: "a", Lat: 0, Lng: -181}, wantErr: domain.ErrCodeValidation},
		{name: "valid payload", payload: &StopPayload{Title: "a", Lat: 55.75, Lng: 37.6}, wantErr: ""},
		{name: "boundary lat -90", payload: &StopPayload{Title: "a", Lat: -90, Lng: 0}, wantErr: ""},
		{name: "boundary lng 180", payload: &StopPayload{Title: "a", Lat: 0, Lng: 180}, wantErr: ""},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			trips := newMockTripRepo()
			trips.trips["trip1"] = &domain.Trip{ID: "trip1", UserID: "u1"}
			svc := NewStopService(mockStopRepo{}, trips)
			_, err := svc.Create(context.Background(), "u1", "trip1", c.payload)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var ae *domain.AppError
			if !errors.As(err, &ae) || ae.Code != c.wantErr {
				t.Fatalf("want err code %s, got %v", c.wantErr, err)
			}
		})
	}
}

func TestStopService_Create_OrderZero(t *testing.T) {
	trips := newMockTripRepo()
	trips.trips["trip1"] = &domain.Trip{ID: "trip1", UserID: "u1"}
	mr := mockStopRepo{maxOrder: 5}
	svc := NewStopService(mr, trips)

	st, err := svc.Create(context.Background(), "u1", "trip1", &StopPayload{Title: "a", Lat: 0, Lng: 0, Order: 0})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if st.Order != 6 {
		t.Fatalf("want order 6 (max 5 + 1), got %d", st.Order)
	}
}

func TestStopService_Create_ExplicitOrder(t *testing.T) {
	trips := newMockTripRepo()
	trips.trips["trip1"] = &domain.Trip{ID: "trip1", UserID: "u1"}
	mr := mockStopRepo{maxOrder: 5}
	svc := NewStopService(mr, trips)

	st, err := svc.Create(context.Background(), "u1", "trip1", &StopPayload{Title: "a", Lat: 0, Lng: 0, Order: 3})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if st.Order != 3 {
		t.Fatalf("want order 3, got %d", st.Order)
	}
}

func TestStopService_List_UnknownTrip(t *testing.T) {
	trips := newMockTripRepo()
	svc := NewStopService(mockStopRepo{}, trips)
	_, err := svc.List(context.Background(), "u1", "missing")
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeNotFound {
		t.Fatalf("want not_found, got %v", err)
	}
}

func TestStopService_Get_ForeignTrip(t *testing.T) {
	trips := newMockTripRepo()
	trips.trips["t1"] = &domain.Trip{ID: "t1", UserID: "owner"}
	svc := NewStopService(mockStopRepo{}, trips)
	_, err := svc.Get(context.Background(), "other", "t1", "s1")
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeNotFound {
		t.Fatalf("want not_found (foreign trip), got %v", err)
	}
}

func TestStopService_Update_ReindexOnOrderChange(t *testing.T) {
	trips := newMockTripRepo()
	trips.trips["trip1"] = &domain.Trip{ID: "trip1", UserID: "u1"}
	mr := mockStopRepo{stops: map[string]*domain.Stop{
		"s1": {ID: "s1", TripID: "trip1", Title: "old", Order: 1},
	}}
	svc := NewStopService(mr, trips)

	// Смена order с 1 на 3 → Reindex(1, 3) должен вызваться.
	st, err := svc.Update(context.Background(), "u1", "trip1", "s1",
		&StopPayload{Title: "new", Lat: 1, Lng: 1, Order: 3})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if st.Order != 3 {
		t.Fatalf("want order 3, got %d", st.Order)
	}
	if len(mr.reindexLog) != 1 || mr.reindexLog[0] != [2]int{1, 3} {
		t.Fatalf("want Reindex(1,3), got %v", mr.reindexLog)
	}
	if mr.updateCalls != 1 {
		t.Fatalf("want 1 Update call, got %d", mr.updateCalls)
	}
}

func TestStopService_Update_NoReindexSameOrder(t *testing.T) {
	trips := newMockTripRepo()
	trips.trips["trip1"] = &domain.Trip{ID: "trip1", UserID: "u1"}
	mr := mockStopRepo{stops: map[string]*domain.Stop{
		"s1": {ID: "s1", TripID: "trip1", Title: "old", Order: 2},
	}}
	svc := NewStopService(mr, trips)

	_, err := svc.Update(context.Background(), "u1", "trip1", "s1",
		&StopPayload{Title: "new", Lat: 0, Lng: 0, Order: 2})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(mr.reindexLog) != 0 {
		t.Fatalf("Reindex should not be called for same order, got %v", mr.reindexLog)
	}
}

func TestStopService_Update_ZeroOrderKeepsCurrent(t *testing.T) {
	trips := newMockTripRepo()
	trips.trips["trip1"] = &domain.Trip{ID: "trip1", UserID: "u1"}
	mr := mockStopRepo{stops: map[string]*domain.Stop{
		"s1": {ID: "s1", TripID: "trip1", Title: "old", Order: 4},
	}}
	svc := NewStopService(mr, trips)

	st, err := svc.Update(context.Background(), "u1", "trip1", "s1",
		&StopPayload{Title: "new", Lat: 0, Lng: 0, Order: 0})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if st.Order != 4 {
		t.Fatalf("want order kept 4, got %d", st.Order)
	}
	if len(mr.reindexLog) != 0 {
		t.Fatalf("Reindex should not be called for order==0, got %v", mr.reindexLog)
	}
}

func TestStopService_Delete_NotFound(t *testing.T) {
	trips := newMockTripRepo()
	trips.trips["trip1"] = &domain.Trip{ID: "trip1", UserID: "u1"}
	svc := NewStopService(mockStopRepo{}, trips)
	err := svc.Delete(context.Background(), "u1", "trip1", "missing")
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeNotFound {
		t.Fatalf("want not_found, got %v", err)
	}
}

func TestStopService_Create_ForeignTrip(t *testing.T) {
	trips := newMockTripRepo()
	trips.trips["t1"] = &domain.Trip{ID: "t1", UserID: "owner"}
	svc := NewStopService(mockStopRepo{}, trips)
	_, err := svc.Create(context.Background(), "other", "t1", &StopPayload{Title: "a", Lat: 0, Lng: 0})
	var ae *domain.AppError
	if !errors.As(err, &ae) || ae.Code != domain.ErrCodeNotFound {
		t.Fatalf("want not_found (foreign trip), got %v", err)
	}
}

var _ = time.Now // keep import (used in future tests)
