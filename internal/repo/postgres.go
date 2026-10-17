package repo

// Реализации репозиториев (CHAT-07: A7-01) — Postgres pgx.
// Контракты: UserRepo, RefreshTokenRepo — internal/repo/interfaces.go.

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"mytrip/internal/domain"
)

// PGUserRepo — UserRepo поверх Postgres (таблица users, 0001_init).
type PGUserRepo struct {
	pool *pgxpool.Pool
}

var _ UserRepo = (*PGUserRepo)(nil)

// NewPGUserRepo — конструктор.
func NewPGUserRepo(pool *pgxpool.Pool) *PGUserRepo { return &PGUserRepo{pool: pool} }

// Insert — INSERT users; дубль email → domain.NewConflict.
func (r *PGUserRepo) Insert(ctx context.Context, u *domain.User, passwordHash string) error {
	var id string
	err := r.pool.QueryRow(ctx, `INSERT INTO users (id, email, name, password_hash, created_at)
		VALUES ($1, $2, $3, $4, now()) RETURNING id`,
		u.ID, u.Email, u.Name, passwordHash,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.NewConflict("email already registered")
		}
		return err
	}
	u.ID = id
	return nil
}

// ByEmail — выборка по email; нет → domain.NewNotFound.
func (r *PGUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	var u domain.User
	hash := ""
	err := r.pool.QueryRow(ctx, `SELECT id, email, name, password_hash, created_at FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.Name, &hash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", domain.NewNotFound("user not found")
		}
		return nil, "", err
	}
	return &u, hash, nil
}

// ByID — выборка по id; нет → domain.NewNotFound.
func (r *PGUserRepo) ByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx, `SELECT id, email, name, created_at FROM users WHERE id = $1::uuid`, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFound("user not found")
		}
		return nil, err
	}
	return &u, nil
}

// PGRefreshTokenRepo — RefreshTokenRepo поверх Postgres (таблица refresh_tokens, 0001_init).
// Хранится hash (SHA-256 hex) — НЕ сам JWT.
type PGRefreshTokenRepo struct {
	pool *pgxpool.Pool
}

var _ RefreshTokenRepo = (*PGRefreshTokenRepo)(nil)

// NewPGRefreshTokenRepo — конструктор.
func NewPGRefreshTokenRepo(pool *pgxpool.Pool) *PGRefreshTokenRepo {
	return &PGRefreshTokenRepo{pool: pool}
}

// Insert — INSERT refresh_tokens (token_hash UNIQUE).
func (r *PGRefreshTokenRepo) Insert(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at) VALUES (gen_random_uuid(), $1::uuid, $2, $3, now())`,
		userID, tokenHash, expiresAt,
	)
	return err
}

// DeleteByHash — удаление по hash (нет строки — не ошибка).
func (r *PGRefreshTokenRepo) DeleteByHash(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

// DeleteByUser — удаление всех токенов пользователя.
func (r *PGRefreshTokenRepo) DeleteByUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1::uuid`, userID)
	return err
}

// DeleteByHashWithResult — удаление по hash с количеством удалённых строк (для ротации).
func (r *PGRefreshTokenRepo) DeleteByHashWithResult(ctx context.Context, tokenHash string) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// isUniqueViolation — PostgreSQL unique_violation (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
