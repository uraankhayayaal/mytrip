package repo

// TripRepo — реализация поверх Postgres pgx (таблица trips, 0002_trips).
// Контракт: internal/repo/interfaces.go (CHAT-07: A7-02).

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mytrip/internal/domain"
)

// PGTripRepo — TripRepo поверх Postgres.
type PGTripRepo struct {
	pool *pgxpool.Pool
}

var _ TripRepo = (*PGTripRepo)(nil)

// NewPGTripRepo — конструктор.
func NewPGTripRepo(pool *pgxpool.Pool) *PGTripRepo { return &PGTripRepo{pool: pool} }

// newTripID — UUID v4 без внешних зависимостей (RFC 4122, вариант 1).
func newTripID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// scanTrip — скан одной строки trips в *domain.Trip.
func scanTrip(row interface{ Scan(...any) error }) (*domain.Trip, error) {
	var (
		t     domain.Trip
		uid   string
		cover *string
	)
	err := row.Scan(&t.ID, &uid, &t.Title, &t.Description, &t.StartDate, &t.EndDate, &cover, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	t.UserID = uid
	t.CoverPhotoURL = cover
	return &t, nil
}

// Insert — INSERT trips; пустой id генерируется (uuid v4), timestamps — now().
func (r *PGTripRepo) Insert(ctx context.Context, t *domain.Trip) error {
	if t.ID == "" {
		t.ID = newTripID()
	}
	var id string
	err := r.pool.QueryRow(ctx, `INSERT INTO trips (id, user_id, title, description, start_date, end_date, cover_photo_url, created_at, updated_at)
		VALUES ($1, $2::uuid, $3, $4, $5, $6, $7, now(), now()) RETURNING id`,
		t.ID, t.UserID, t.Title, t.Description, t.StartDate, t.EndDate, t.CoverPhotoURL,
	).Scan(&id)
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

// ListByUser — все поездки пользователя, ORDER BY start_date DESC.
func (r *PGTripRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Trip, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, user_id, title, description, start_date, end_date, cover_photo_url, created_at, updated_at
		FROM trips WHERE user_id = $1::uuid ORDER BY start_date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*domain.Trip
	for rows.Next() {
		t, err := scanTrip(rows)
		if err != nil {
			return nil, err
		}
		trips = append(trips, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return trips, nil
}

// Get — поездка по id и owner; нет/чужая → domain.NewNotFound.
func (r *PGTripRepo) Get(ctx context.Context, userID, id string) (*domain.Trip, error) {
	t, err := scanTrip(r.pool.QueryRow(ctx, `SELECT id, user_id, title, description, start_date, end_date, cover_photo_url, created_at, updated_at
		FROM trips WHERE id = $1::uuid AND user_id = $2::uuid`, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("trip not found")
		}
		return nil, err
	}
	return t, nil
}

// Update — обновление полей поездки; чужая/нет → domain.NewNotFound.
func (r *PGTripRepo) Update(ctx context.Context, t *domain.Trip) error {
	tag, err := r.pool.Exec(ctx, `UPDATE trips SET title = $1, description = $2, start_date = $3, end_date = $4, cover_photo_url = $5, updated_at = now()
		WHERE id = $6::uuid AND user_id = $7::uuid`,
		t.Title, t.Description, t.StartDate, t.EndDate, t.CoverPhotoURL, t.ID, t.UserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFound("trip not found")
	}
	return nil
}

// Delete — удаление поездки; нет строк → domain.NewNotFound.
func (r *PGTripRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM trips WHERE id = $1::uuid AND user_id = $2::uuid`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFound("trip not found")
	}
	return nil
}

// SetCover — смена cover_photo_url; нет строк → domain.NewNotFound.
func (r *PGTripRepo) SetCover(ctx context.Context, userID, tripID, coverURL string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE trips SET cover_photo_url = $1, updated_at = now()
		WHERE id = $2::uuid AND user_id = $3::uuid`, coverURL, tripID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFound("trip not found")
	}
	return nil
}
