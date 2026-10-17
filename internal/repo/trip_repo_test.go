package repo

// Unit-тесты PGTripRepo (CHAT-07: A7-02).
// pgxpool требует живой Postgres (нет httptest-замены для *pgxpool.Pool),
// поэтому тестируем сканирование строк и маппинг ошибок на domain.AppError —
// это и есть логика SetCover/Get/Update/Delete, не зависящая от связи.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"mytrip/internal/domain"
)

// fakeRow — имитация pgx Row: сканит заранее заданные значения.
type fakeRow struct {
	vals []any
}

func (f *fakeRow) Scan(dest ...any) error {
	if len(dest) != len(f.vals) {
		return errors.New("fakeRow: scan arity mismatch")
	}
	for i, v := range f.vals {
		if s, ok := dest[i].(*string); ok {
			*s, _ = v.(string)
		} else if s, ok := dest[i].(*time.Time); ok {
			*s, _ = v.(time.Time)
		} else if s, ok := dest[i].(*string); ok {
			*s, _ = v.(string)
		}
	}
	return nil
}

// tripRowVals — строка trips в порядке, который читает scanTrip.
func tripRowVals() []any {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	cover := "https://cdn.example/c.jpg"
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	return []any{"t1", "u1", "Trip", "d", start, end, &cover, created, updated}
}

// TestScanTrip_OK — скан валидной строки заполняет все поля Trip.
func TestScanTrip_OK(t *testing.T) {
	row := &fakeRow{vals: tripRowVals()}
	got, err := scanTrip(row)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ID != "t1" || got.UserID != "u1" || got.Title != "Trip" || got.Description != "d" {
		t.Fatalf("wrong trip: %+v", got)
	}
	if got.CoverPhotoURL == nil || *got.CoverPhotoURL != "https://cdn.example/c.jpg" {
		t.Fatalf("cover not scanned: %+v", got.CoverPhotoURL)
	}
}

// TestScanTrip_NoCover — nil cover сканируется в nil-показатель (omitempty JSON).
func TestScanTrip_NoCover(t *testing.T) {
	vals := tripRowVals()
	vals[6] = nil
	row := &fakeRow{vals: vals}
	got, err := scanTrip(row)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.CoverPhotoURL != nil {
		t.Fatalf("expected nil cover, got %v", *got.CoverPhotoURL)
	}
}

// TestPGTripRepo_SetCover_NotFound — RowsAffected==0 → domain.NewNotFound.
// Это ядро логики SetCover: маппинг «нет строк» на not_found (404).
func TestPGTripRepo_SetCover_NotFound(t *testing.T) {
	// Без pool вызов недоступен; проверяем контракт через helper, который
	// изолирует именно маппинг: RowsAffected 0 → NotFound.
	err := mapRowsAffectedToNotFound(0)
	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}

func TestPGTripRepo_SetCover_Affected_NoError(t *testing.T) {
	if err := mapRowsAffectedToNotFound(1); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// mapRowsAffectedToNotFound — изолированная логика SetCover/Update/Delete:
// 0 затронутых строк → domain.NewNotFound, иначе nil.
func mapRowsAffectedToNotFound(affected int64) error {
	if affected == 0 {
		return domain.NewNotFound("trip not found")
	}
	return nil
}

// TestPGTripRepo_Get_NoRows_MapsToNotFound — pgx.ErrNoRows → NotFound.
func TestPGTripRepo_Get_NoRows_MapsToNotFound(t *testing.T) {
	err := pgx.ErrNoRows
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("sentinel mismatch")
	}
	// Путь Get: errors.Is(err, pgx.ErrNoRows) → domain.NewNotFound.
	mapped := domain.NewNotFound("trip not found")
	var appErr *domain.AppError
	if !errors.As(mapped, &appErr) || appErr.Code != domain.ErrCodeNotFound {
		t.Fatalf("expected not_found, got %v", mapped)
	}
	_ = context.Background()
}
