package service

// Слой сервисов (Clean Architecture: application-слой).
// Задачи CHAT-07: реализации пишут рядовые разработчики, контракты зафиксированы здесь.
// Правило: публичные сигнатуры менять НЕЛЬЗЯ.

import (
	"context"
	"time"

	"mytrip/internal/domain"
)

// AuthService — аутентификация (register/login/refresh/logout/me).
// Контракт API — см. internal/api/handlers/auth.go (CHAT-07: A7-01).
//
// Реализация (TODO A7-01):
//   - register: bcrypt-хэш пароля, INSERT users; дубль email → domain.NewConflict.
//   - login: проверка bcrypt, выдача access JWT (sub=user_id, exp=cfg.JWTAccessTTL) и refresh JWT (exp=cfg.JWTRefreshTTL).
//     refresh записывается в httpOnly cookie 'rt' (Set-Cookie) — это делает handler, не сервис; сервис возвращает строку refresh.
//   - refresh: принимает refresh JWT, валидация (подпись, exp, тип 'refresh'), ротация: новый refresh, старый hash — удалить.
//   - logout: удалить refresh-токен (hash) из БД.
//   - me: пользователь по user_id из контекста (задаётся auth middleware).
//
// DB: users (0001_init), refresh_tokens (0001_init: token_hash UNIQUE, expires_at).
type AuthService interface {
	Register(ctx context.Context, email, password, name string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (user *domain.User, accessToken string, refreshToken string, err error)
	// Refresh: refreshToken — строка JWT из httpOnly cookie 'rt'.
	// Возвращает новый access и новый refresh (ротация); истёкший/чужой — domain.NewUnauthorized.
	Refresh(ctx context.Context, refreshToken string) (accessToken string, newRefreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, userID string) (*domain.User, error)
}

// TripService — CRUD поездок (CHAT-07: A7-02).
//
// Реализация (TODO A7-02):
//   - List: по userID, ORDER BY created_at DESC; чужие поездки — недоступны (фильтр user_id, а не 404).
//   - Get/Update/Delete: чужая или несуществующая → domain.NewNotFound (404, не 403).
//   - валидация: title обязателен (non-blank), end_date >= start_date → domain.NewValidation (400).
//
// DB: trips (0002_trips). Валидация дат — time.Time сравнивается напрямую.
type TripService interface {
	List(ctx context.Context, userID string) ([]*domain.Trip, error)
	Get(ctx context.Context, userID, tripID string) (*domain.Trip, error)
	Create(ctx context.Context, userID string, p *domain.TripPayload) (*domain.Trip, error)
	Update(ctx context.Context, userID, tripID string, p *domain.TripPayload) (*domain.Trip, error)
	Delete(ctx context.Context, userID, tripID string) error
}

// TripPayload — входные данные создания/обновления поездки (без ID/timestamps).
type TripPayload struct {
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	CoverPhotoURL *string   `json:"cover_photo_url,omitempty"`
}

// StopService — CRUD точек маршрута (CHAT-07: A7-03).
//
// Реализация (TODO A7-03):
//   - List: по tripID, ORDER BY order ASC; сначала проверка владения trip (владельца нет → NewNotFound).
//   - Create: если order не задан (0) — order = max(order)+1 по trip.
//   - Update: при смене order — реиндексация остальных (сдвиг); валидация: |lat|<=90, |lng|<=180, title non-blank → NewValidation.
//
// DB: stops (0003_stops). Контракт JSON — см. internal/api/handlers/stops.go (location {lat,lng}).
type StopService interface {
	List(ctx context.Context, userID, tripID string) ([]*domain.Stop, error)
	Get(ctx context.Context, userID, tripID, stopID string) (*domain.Stop, error)
	Create(ctx context.Context, userID, tripID string, p *domain.StopPayload) (*domain.Stop, error)
	Update(ctx context.Context, userID, tripID, stopID string, p *domain.StopPayload) (*domain.Stop, error)
	Delete(ctx context.Context, userID, tripID, stopID string) error
}

// StopPayload — входные данные создания/обновления точки (без ID/timestamps).
type StopPayload struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	VisitDate   time.Time `json:"visit_date"`
	Order       int       `json:"order"`
}

// PhotoService — управление фотографиями поездки (CHAT-07: A7-04).
//
// Реализация (TODO A7-04):
//   - List: по tripID (владение проверяется — чужой trip → NewNotFound).
//   - Create: после PUT объекта в S3 (через storage-порт): INSERT photos (url = key/публичный URL),
//     при cover=true — обновить trips.cover_photo_url.
//   - Delete: DELETE из БД + DELETE объекта из S3 (ошибки S3 — log, не 500: запись в БД уже удалена).
//
// DB: photos (0004_photos); S3 через порт StoragePort.
type PhotoService interface {
	List(ctx context.Context, userID, tripID string) ([]*domain.Photo, error)
	// Create: objectKey — ключ объекта в бакете (после PUT-пресайна), url — итоговый URL записи, cover — сделать обложкой.
	Create(ctx context.Context, userID, tripID, objectKey, url string, cover bool) (*domain.Photo, error)
	Delete(ctx context.Context, userID, tripID, photoID string) error
}

// StoragePort — порт доступа к объектному хранилищу (S3-совместимый).
// Реализация (TODO A7-04): адаптер над aws-sdk-go-v2 s3.Client (уже создаётся в cmd/server/main.go).
type StoragePort interface {
	// PresignPut — URL предподписанного PUT объекта (frontend загружает фото напрямую в S3).
	PresignPut(ctx context.Context, objectKey string) (url string, expiresAt time.Time, err error)
	DeleteObject(ctx context.Context, objectKey string) error
}
