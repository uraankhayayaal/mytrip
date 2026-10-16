package handlers

// REST-эндпоинты API (CHAT-07: A7-01..A7-05).
// Контракты: методы, пути, тела JSON и коды ошибок зафиксированы комментариями ниже.
// Правило: публичные сигнатуры и JSON-схемы менять НЕЛЬЗЯ.
//
// Общие правила:
//   - user_id берётся из контекста (задаёт auth middleware, A7-05) — НЕМАЯ параметром из body/path.
//   - ошибки пишутся через middleware.WriteError (internal/api/middleware/error_handler.go),
//     success — через middleware.WriteJSON.
//   - формат ошибки: {"error":{"code","message","details?"}} (domain.AppError).

// AuthHandler — /api/v1/auth/* (CHAT-07: A7-01).
//
// Эндпоинты и контракты:
//
//	POST /auth/register
//	  req : {"email":"string","password":"string(>=8)","name":"string(non-blank)"}
//	  ok  : 201 {"user":{"id","email","name","created_at"}}
//	  err : 400 validation (email не валиден / короткая password), 409 conflict (email занят)
//	POST /auth/login
//	  req : {"email":"string","password":"string"}
//	  ok  : 200 {"access_token":"<jwt>","user":{...}} + Set-Cookie: rt=<jwt>; HttpOnly; Path=/; Max-Age=2592000; SameSite=Lax
//	  err : 401 unauthorized (неверный email или password — единое сообщение)
//	POST /auth/refresh
//	  req : пустое тело; refresh берётся из httpOnly cookie 'rt'
//	  ok  : 200 {"access_token":"<jwt>"} + Set-Cookie: rt=<new jwt> (ротация)
//	  err : 401 unauthorized (токен истёк/подпись неверна/типа access)
//	POST /auth/logout
//	  req : пустое тело; cookie 'rt' — удалить (Set-Cookie: rt=; Max-Age=0)
//	  ok  : 204
//	GET  /users/me
//	  ok  : 200 {"id","email","name","created_at"}
//	  err : 401
//
// Реализация: тела — через service.AuthService; парсинг JWT и Set-Cookie — здесь (handler),
// не в сервисе. Cookie-имя 'rt', путь '/', SameSite=Lax, HttpOnly, Max-Age = refresh TTL.
type AuthHandler struct{}

// TripHandler — /api/v1/trips* (CHAT-07: A7-02).
//
// Эндпоинты и контракты:
//
//	GET    /trips
//	  ok  : 200 [Trip]  (Trip — см. JSON-теги domain.Trip: snake_case, cover_photo_url omitempty)
//	POST   /trips
//	  req : {"title":"non-blank","description":"string","start_date":"RFC3339","end_date":"RFC3339","cover_photo_url?":"string"}
//	  ok  : 201 {Trip}
//	  err : 400 validation (title пуст / end_date < start_date)
//	GET    /trips/{id}
//	  ok  : 200 {Trip}
//	  err : 404 not_found (нет или чужая)
//	PUT    /trips/{id}
//	  req : {TripPayload как у POST}
//	  ok  : 200 {Trip}; 404 not_found; 400 validation
//	DELETE /trips/{id}
//	  ok  : 204; 404 not_found
//
// Реализация: тела — через service.TripService. ID из chi-параметра {id}. JSON-поля даты — RFC3339 (time.Time default).
type TripHandler struct{}

// StopHandler — /api/v1/trips/{id}/stops* (CHAT-07: A7-03).
//
// Эндпоинты и контракты:
//
//	GET    /trips/{id}/stops
//	  ok  : 200 [StopJSON]; 404 (нет trip)
//	POST   /trips/{id}/stops
//	  req : {"title":"non-blank","description":"string","location":{"lat":"number(|lat|<=90)","lng":"number(|lng|<=180)"},"visit_date":"RFC3339","order":"int(0=авто)"}
//	  ok  : 201 {StopJSON}; 404 (нет trip); 400 validation
//	GET    /trips/{id}/stops/{stopId}
//	  ok  : 200 {StopJSON}; 404
//	PUT    /trips/{id}/stops/{stopId}
//	  req : {StopPayload как у POST}
//	  ok  : 200 {StopJSON}; 404; 400
//	DELETE /trips/{id}/stops/{stopId}
//	  ok  : 204; 404
//
// StopJSON — DTO на границе API (frontend-тип Stop использует location:{lat,lng},
// а domain.Stop — плоские Lat/Lng): см. StopJSON ниже.
//
// Реализация: тела — через service.StopService; маппинг Stop<->StopJSON — в хендлере (или DTO-помощники в этом пакете).
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
//
// Эндпоинты и контракты:
//
//	GET    /trips/{id}/photos
//	  ok  : 200 [Photo] (domain.Photo: id, trip_id, url, created_at); 404
//	POST   /trips/{id}/photos
//	  req : {"object_key":"string","url":"string","cover?":false}
//	  ok  : 201 {Photo}; 404; 400 (object_key пуст)
//	DELETE /trips/{id}/photos/{photoId}
//	  ok  : 204; 404
//	GET    /trips/{id}/photos/presign  (или POST /trips/{id}/photos/presign)
//	  req : {"object_key":"string"}
//	  ok  : 200 {"url":"<presigned put url>","expires_at":"RFC3339"}  (frontend-тип PhotoUrlResponse)
//	  err : 404; 400 (object_key пуст)
//
// Реализация: тела — через service.PhotoService + service.StoragePort (PresignPut).
// Порядок presign: сначала проверка владения trip (404), затем presign.
type PhotoHandler struct{}

// ExportHandler — GET /api/v1/trips/{id}/export (уже смонтирован в cmd/server/main.go).
// Контракт (не менять):
//
//	GET /trips/{id}/export
//	  ok  : 200, Content-Type: text/csv; charset=utf-8, Content-Disposition: attachment; filename="<trip-title>.csv"
//	  тело: CSV: заголовок "order;title;description;lat;lng;visit_date" (или , — выбрать и зафиксировать в реализации),
//	        по строке на stop в порядке order, visit_date — RFC3339.
//	  err : 404 (нет trip), 401 (нет auth)
//
// Реализация: service.TripService.Get + service.StopService.List; форматирование CSV — здесь.
type ExportHandler struct{}
