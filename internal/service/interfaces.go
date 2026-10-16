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
// Контракт API — см. internal/api/handlers/handlers.go (CHAT-07: A7-01).
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
type TripService interface {
	List(ctx context.Context, userID string) ([]*domain.Trip, error)
	Get(ctx context.Context, userID, tripID string) (*domain.Trip, error)
	Create(ctx context.Context, userID string, p *TripPayload) (*domain.Trip, error)
	Update(ctx context.Context, userID, tripID string, p *TripPayload) (*domain.Trip, error)
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
type StopService interface {
	List(ctx context.Context, userID, tripID string) ([]*domain.Stop, error)
	Get(ctx context.Context, userID, tripID, stopID string) (*domain.Stop, error)
	Create(ctx context.Context, userID, tripID string, p *StopPayload) (*domain.Stop, error)
	Update(ctx context.Context, userID, tripID, stopID string, p *StopPayload) (*domain.Stop, error)
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
type PhotoService interface {
	List(ctx context.Context, userID, tripID string) ([]*domain.Photo, error)
	Create(ctx context.Context, userID, tripID, objectKey, url string, cover bool) (*domain.Photo, error)
	Delete(ctx context.Context, userID, tripID, photoID string) error
}

// StoragePort — порт доступа к объектному хранилищу (S3-совместимый).
type StoragePort interface {
	PresignPut(ctx context.Context, objectKey string) (url string, expiresAt time.Time, err error)
	DeleteObject(ctx context.Context, objectKey string) error
}
