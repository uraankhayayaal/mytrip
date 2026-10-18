package repo

// Unit-тесты StopRepo без живой БД (CHAT-07: A7-03).
// Покрывают scanStop (формат строк trip_stops) и чистую логику Reindex
// (no-op при from==to, обращение к пулу — только при реальном сдвиге).

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"mytrip/internal/domain"
)

// --- fake row ---

type fakeStopRow struct {
	vals []any
	err  error
}

func (f *fakeStopRow) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}
	if len(f.vals) != len(dest) {
		return fmt.Errorf("fakeStopRow: want %d dests, got %d", len(f.vals), len(dest))
	}
	for i := range dest {
		switch p := dest[i].(type) {
		case *any:
			*p = f.vals[i]
		case *string:
			s, ok := f.vals[i].(string)
			if !ok {
				return fmt.Errorf("fakeStopRow: dest %d wants string, got %T", i, f.vals[i])
			}
			*p = s
		case *float64:
			n, ok := f.vals[i].(float64)
			if !ok {
				return fmt.Errorf("fakeStopRow: dest %d wants float64, got %T", i, f.vals[i])
			}
			*p = n
		case *int:
			n, ok := f.vals[i].(int)
			if !ok {
				return fmt.Errorf("fakeStopRow: dest %d wants int, got %T", i, f.vals[i])
			}
			*p = n
		case *time.Time:
			tm, ok := f.vals[i].(time.Time)
			if !ok {
				return fmt.Errorf("fakeStopRow: dest %d wants time.Time, got %T", i, f.vals[i])
			}
			*p = tm
		}
	}
	return nil
}

// --- тесты scanStop ---

func TestScanStop_OK(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	row := &fakeStopRow{vals: []any{"id-1", "trip-1", "title", "desc", 55.75, 37.6, ts, 3, ts}}

	s, err := scanStop(row)
	if err != nil {
		t.Fatalf("scanStop: %v", err)
	}
	if s.ID != "id-1" || s.TripID != "trip-1" || s.Title != "title" || s.Description != "desc" {
		t.Fatalf("unexpected stop: %+v", s)
	}
	if s.Lat != 55.75 || s.Lng != 37.6 || s.Order != 3 || !s.VisitDate.Equal(ts) || !s.CreatedAt.Equal(ts) {
		t.Fatalf("unexpected stop: %+v", s)
	}
}

func TestScanStop_ErrNoRows(t *testing.T) {
	row := &fakeStopRow{err: pgx.ErrNoRows}
	if _, err := scanStop(row); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("want ErrNoRows, got %v", err)
	}
}

func TestScanStop_Mismatch(t *testing.T) {
	row := &fakeStopRow{vals: []any{"only-one"}}
	if _, err := scanStop(row); err == nil {
		t.Fatal("want error on column mismatch, got nil")
	}
}

// --- compile-time check: PGTripStopRepo реализует StopRepo (в т.ч. Reindex) ---

var _ StopRepo = (*PGTripStopRepo)(nil)

// domain.Stop — целевой тип scanStop: фиксация полей на границе scan.
var _ = (*domain.Stop)(nil)

// --- Reindex: чистая логика без БД ---
//
// Позиционная корректность SQL (сдвиг строк в интервале) проверяется только
// против живой БД — интеграционным тестом (make e2e в CI).
// Здесь фиксируем: (1) from==to → no-op без обращения к пулу;
// (2) to!=from → обращение к пулу (nil-pool → panic, т.е. SQL действительно
// исполнялся, а не пропущен).

func TestReindex_NoopWhenFromEqualsTo(t *testing.T) {
	r := NewPGStopRepo(nil) // nil-pool допустим: при from==to пул не используется
	if err := r.Reindex(context.Background(), "trip1", 2, 2); err != nil {
		t.Fatalf("Reindex no-op: %v", err)
	}
}

func TestReindex_RealShiftTouchesPool(t *testing.T) {
	r := NewPGStopRepo(nil) // nil-pool → panic при обращении к пулу
	for _, c := range []struct{ from, to int }{{1, 3}, {3, 1}} {
		t.Run(fmt.Sprintf("from%d_to%d", c.from, c.to), func(t *testing.T) {
			defer func() {
				if rec := recover(); rec == nil {
					t.Fatal("want panic on nil pool (pool must be touched), got none")
				}
			}()
			_ = r.Reindex(context.Background(), "trip1", c.from, c.to)
		})
	}
}
