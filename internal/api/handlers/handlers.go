package handlers

// REST-эндпоинты API (CHAT-07: A7-01..A7-05).
// Контракты: методы, пути, тела JSON и коды ошибок зафиксированы комментариями ниже.
// Правило: публичные сигнатуры и JSON-схемы менять НЕЛЬЗЯ.

import "net/http"

// AuthHandler — /api/v1/auth/* (CHAT-07: A7-01).
//
// Эндпоинты и контракты:
//
//	POST /auth/register
//	  req : {"email":"string","password":"string(>=8)","name":"string(non-blank)"}
//	  ok  : 201 {"user":{"id","email","name","created_at"}}
//	  err : 400 validation, 409 conflict (email занят)
//	POST /auth/login
//	  req : {"email":"string","password":"string"}
//	  ok  : 200 {"access_token":"<jwt>","user":{...}} + Set-Cookie: rt=<jwt>; HttpOnly; Path=/; Max-Age=2592000; SameSite=Lax
//	  err : 401 unauthorized
//	POST /auth/refresh
//	  req : пустое тело; refresh берётся из httpOnly cookie 'rt'
//	  ok  : 200 {"access_token":"<jwt>"} + Set-Cookie: rt=<new jwt> (ротация)
//	  err : 401 unauthorized
//	POST /auth/logout
//	  req : пустое тело; cookie 'rt' — удалить (Set-Cookie: rt=; Max-Age=0)
//	  ok  : 204
//	GET  /users/me
//	  ok  : 200 {"id","email","name","created_at"}
//	  err : 401
//
// Реализация: auth_handler.go (NewAuthHandler/Mount).
type AuthHandler struct {
	handler http.Handler
}

// TripHandler — /api/v1/trips* (CHAT-07: A7-02).
type TripHandler struct{}

// StopHandler — /api/v1/trips/{id}/stops* (CHAT-07: A7-03).
type StopHandler struct{}

// StopJSON — JSON-представление остановки на границе API.
type StopJSON struct {
	ID          string    `json:"id"`
	TripID      string    `json:"trip_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    StopLoc   `json:"location"`
	VisitDate   time.Time `json:"visit_date"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StopLoc struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// PhotoHandler — /api/v1/trips/{id}/photos* (CHAT-07: A7-04).
type PhotoHandler struct{}

// ExportHandler — GET /api/v1/trips/{id}/export.
type ExportHandler struct{}
