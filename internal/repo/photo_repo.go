package repo

// PhotoRepo — реализация поверх Postgres pgx (таблица trip_photos, 0004_photos).
// Контракт: internal/repo/interfaces.go (CHAT-07: A7-04).

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"mytrip/internal/domain"
)

// PGPhotoRepo — PhotoRepo поверх Postgres.
type PGPhotoRepo struct {
	pool *pgxpool.Pool
}

var _ PhotoRepo = (*PGPhotoRepo)(nil)

// NewPGPhotoRepo — конструктор.
func NewPGPhotoRepo(pool *pgxpool.Pool) *PGPhotoRepo { return &PGPhotoRepo{pool: pool} }

// scanPhoto — скан одной строки trip_photos в *domain.Photo.
func scanPhoto(row interface{ Scan(...any) error }) (*domain.Photo, error) {
	var p domain.Photo
	if err := row.Scan(&p.ID, &p.TripID, &p.URL, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

// Insert — INSERT trip_photos; пустой id генерируется (uuid v4), created_at — now().
func (r *PGPhotoRepo) Insert(ctx context.Context, p *domain.Photo) error {
	if p.ID == "" {
		p.ID = newTripID()
	}
	var id string
	err := r.pool.QueryRow(ctx, `INSERT INTO trip_photos (id, trip_id, url, created_at)
		VALUES ($1, $2::uuid, $3, now()) RETURNING id`,
		p.ID, p.TripID, p.URL,
	).Scan(&id)
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

// ListByTrip — все фото поездки, ORDER BY created_at DESC.
func (r *PGPhotoRepo) ListByTrip(ctx context.Context, tripID string) ([]*domain.Photo, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, trip_id, url, created_at FROM trip_photos WHERE trip_id = $1::uuid ORDER BY created_at DESC`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []*domain.Photo
	for rows.Next() {
		p, err := scanPhoto(rows)
		if err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return photos, nil
}

// Delete — удаление фото; нет строк → domain.NewNotFound.
func (r *PGPhotoRepo) Delete(ctx context.Context, tripID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM trip_photos WHERE id = $1::uuid AND trip_id = $2::uuid`, id, tripID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFound("photo not found")
	}
	return nil
}
