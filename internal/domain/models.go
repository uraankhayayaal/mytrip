package domain

import "time"

// User — пользователь приложения.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Trip — поездка пользователя.
type Trip struct {
	ID            string    `json:"id"`
	UserID        string    `json:"-"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	CoverPhotoURL *string   `json:"cover_photo_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Stop — точка маршрута внутри поездки.
type Stop struct {
	ID          string    `json:"id"`
	TripID      string    `json:"trip_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	VisitDate   time.Time `json:"visit_date"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
}

// Photo — фотография поездки.
type Photo struct {
	ID        string    `json:"id"`
	TripID    string    `json:"trip_id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// Location — географическая точка.
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
