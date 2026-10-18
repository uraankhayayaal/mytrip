package repo

// StopRepo — реализация поверх Postgres pgx (таблица trip_stops, 0003_stops).
// Контракт: internal/repo/interfaces.go (CHAT-07: A7-03).
// Reindex: сдвиг order строк в интервале (from, to] при to > from и [to, from)
// при to < from, в одной транзакции (Update выполняется сервисом в той же tx).

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mytrip/internal/domain"
)

// PGTripStopRepo — StopRepo поверх Postgres.
type PGTripStopRepo struct {
	pool *pgxpool.Pool
}

var _ StopRepo = (*PGTripStopRepo)(nil)

// NewPGStopRepo — конструктор StopRepo.
func NewPGStopRepo(pool *pgxpool.Pool) *PGTripStopRepo { return &PGTripStopRepo{pool: pool} }

// scanStop — скан одной строки trip_stops в *domain.Stop.
func scanStop(row interface{ Scan(...any) error }) (*domain.Stop, error) {
	var s domain.Stop
	if err := row.Scan(&s.ID, &s.TripID, &s.Title, &s.Description, &s.Lat, &s.Lng, &s.VisitDate, &s.Order, &s.CreatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

// Insert — INSERT trip_stops; пустой id генерируется (uuid v4), created_at — now().
func (r *PGTripStopRepo) Insert(ctx context.Context, s *domain.Stop) error {
	if s.ID == "" {
		s.ID = newTripID()
	}
	var id string
	err := r.pool.QueryRow(ctx, `INSERT INTO trip_stops (id, trip_id, title, description, lat, lng, visit_date, "order", created_at)
		VALUES ($1, $2::uuid, $3, $4, $5, $6, $7, $8, now()) RETURNING id`,
		s.ID, s.TripID, s.Title, s.Description, s.Lat, s.Lng, s.VisitDate, s.Order,
	).Scan(&id)
	if err != nil {
		return err
	}
	s.ID = id
	return nil
}

// ListByTrip — все точки маршрута поездки, ORDER BY "order" ASC.
func (r *PGTripStopRepo) ListByTrip(ctx context.Context, tripID string) ([]*domain.Stop, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, trip_id, title, description, lat, lng, visit_date, "order", created_at
		FROM trip_stops WHERE trip_id = $1::uuid ORDER BY "order" ASC`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stops []*domain.Stop
	for rows.Next() {
		s, err := scanStop(rows)
		if err != nil {
			return nil, err
		}
		stops = append(stops, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stops, nil
}

// Get — точка по trip_id и id; нет/чужая → domain.NewNotFound.
func (r *PGTripStopRepo) Get(ctx context.Context, tripID, id string) (*domain.Stop, error) {
	s, err := scanStop(r.pool.QueryRow(ctx, `SELECT id, trip_id, title, description, lat, lng, visit_date, "order", created_at
		FROM trip_stops WHERE id = $1::uuid AND trip_id = $2::uuid`, id, tripID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("stop not found")
		}
		return nil, err
	}
	return s, nil
}

// Update — обновление полей точки; нет строк → domain.NewNotFound.
func (r *PGTripStopRepo) Update(ctx context.Context, s *domain.Stop) error {
	tag, err := r.pool.Exec(ctx, `UPDATE trip_stops SET title = $1, description = $2, lat = $3, lng = $4, visit_date = $5, "order" = $6
		WHERE id = $7::uuid AND trip_id = $8::uuid`,
		s.Title, s.Description, s.Lat, s.Lng, s.VisitDate, s.Order, s.ID, s.TripID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFound("stop not found")
	}
	return nil
}

// Delete — удаление точки; нет строк → domain.NewNotFound.
func (r *PGTripStopRepo) Delete(ctx context.Context, tripID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM trip_stops WHERE id = $1::uuid AND trip_id = $2::uuid`, id, tripID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFound("stop not found")
	}
	return nil
}

// MaxOrder — максимальный order в поездке; пустая поездка → 0.
func (r *PGTripStopRepo) MaxOrder(ctx context.Context, tripID string) (int, error) {
	var max int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX("order"), 0) FROM trip_stops WHERE trip_id = $1::uuid`, tripID).Scan(&max)
	if err != nil {
		return 0, err
	}
	return max, nil
}

// Reindex — сдвиг order строк при перемещении точки с позиции from на to:
//   - to > from: строки с order в (from, to] сдвигаются +1;
//   - to < from: строки с order в [to, from) сдвигаются -1.
//
// Вызывается до Update (см. stopService.Update), поэтому перемещаемая строка
// ещё стоит на позиции from и исключается из сдвига явным id <> $4::uuid.
func (r *PGTripStopRepo) Reindex(ctx context.Context, tripID string, from, to int, stopID string) error {
	if from == to {
		return nil
	}
	var q string
	if to > from {
		q = `UPDATE trip_stops SET "order" = "order" + 1
			WHERE trip_id = $1::uuid AND "order" > $2 AND "order" <= $3 AND id <> $4::uuid`
	} else {
		q = `UPDATE trip_stops SET "order" = "order" - 1
			WHERE trip_id = $1::uuid AND "order" >= $2 AND "order" < $3 AND id <> $4::uuid`
	}
	_, err := r.pool.Exec(ctx, q, tripID, from, to, stopID)
	return err
}
