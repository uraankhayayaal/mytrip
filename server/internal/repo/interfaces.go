package repo

// Слой репозиториев (Clean Architecture: infrastructure-слой, Postgres pgx).
// Задачи CHAT-07: реализации пишут рядовые разработчики, контракты зафиксированы здесь.
// Правило: публичные сигнатуры менять НЕЛЬЗЯ.
//
// Общий стиль:
//   - pgx pool (*pgxpool.Pool) — из cmd/server/main.go.
//   - UUID — string (cast в SQL: $1::uuid).
//   - Ошибки «нет строки» → domain.NewNotFound, «дубль» → domain.NewConflict.
//   - Контекст — всегда первый аргумент; ctx.Err() — propagate.

import (
	"context"
	"time"

	"mytrip/internal/domain"
)

// UserRepo — пользователи (таблица users, 0001_init).
// TODO(A7-01): реализация по задаче A7-01.
type UserRepo interface {
	Insert(ctx context.Context, u *domain.User, passwordHash string) error
	ByEmail(ctx context.Context, email string) (*domain.User, string /*passwordHash*/, error)
	ByID(ctx context.Context, id string) (*domain.User, error)
}

// RefreshTokenRepo — refresh-токены (таблица refresh_tokens, 0001_init).
// Хранится hash (SHA-256 hex) — НЕ сам JWT.
// TODO(A7-01): реализация по задаче A7-01.
type RefreshTokenRepo interface {
	Insert(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	DeleteByHash(ctx context.Context, tokenHash string) error
	DeleteByUser(ctx context.Context, userID string) error
}

// TripRepo — поездки (таблица trips, 0002_trips).
// TODO(A7-02): реализация по задаче A7-02.
type TripRepo interface {
	Insert(ctx context.Context, t *domain.Trip) error
	ListByUser(ctx context.Context, userID string) ([]*domain.Trip, error)
	Get(ctx context.Context, userID, id string) (*domain.Trip, error)
	Update(ctx context.Context, t *domain.Trip) error
	Delete(ctx context.Context, userID, id string) error
	SetCover(ctx context.Context, userID, tripID, coverURL string) error
}

// StopRepo — точки маршрута (таблица stops, 0003_stops).
// TODO(A7-03): реализация по задаче A7-03.
type StopRepo interface {
	Insert(ctx context.Context, s *domain.Stop) error
	ListByTrip(ctx context.Context, tripID string) ([]*domain.Stop, error) // ORDER BY order ASC
	Get(ctx context.Context, tripID, id string) (*domain.Stop, error)
	Update(ctx context.Context, s *domain.Stop) error
	Delete(ctx context.Context, tripID, id string) error
	MaxOrder(ctx context.Context, tripID string) (int, error)
	// Reindex — сдвиг order после Update с изменением order (в одной транзакции с Update в сервисе).
	Reindex(ctx context.Context, tripID string, from, to int) error
}

// PhotoRepo — фото (таблица photos, 0004_photos).
// TODO(A7-04): реализация по задаче A7-04.
type PhotoRepo interface {
	Insert(ctx context.Context, p *domain.Photo) error
	ListByTrip(ctx context.Context, tripID string) ([]*domain.Photo, error)
	Delete(ctx context.Context, tripID, id string) error
}
