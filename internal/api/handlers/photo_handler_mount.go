package handlers

// Фото-маршрут "/{tail...}" требует хотя бы один сегмент, поэтому точный
// базовый путь /trips/{id}/photos (без хвоста) регистрируется отдельно.
// Маршруты /presign и /{photoId} — конкретнее wildcard'а и не пересекаются.
//
//	GET    /trips/{id}/photos        → 200 [PhotoJSON]
//	POST   /trips/{id}/photos        → 201 PhotoJSON; body {object_key, url, cover?}
//	DELETE /trips/{id}/photos/{photoId} → 204
//	POST   /trips/{id}/photos/presign   → 200 {url, expires_at} (frontend PhotoUrlResponse)

// mountPhotoRoutes — регистрация фото-маршрутов на chi router.
func mountPhotoRoutes(r chi.Router, impl *photoHandler) {
	r.Post("/trips/{id}/photos/presign", impl.presign)
	r.Delete("/trips/{id}/photos/{photoId}", impl.delete)
	r.Get("/trips/{id}/photos", impl.list)
	r.Post("/trips/{id}/photos", impl.create)
	r.Get("/trips/{id}/photos/{tail...}", impl.list)
	r.Post("/trips/{id}/photos/{tail...}", impl.create)
}

// Mount — регистрация маршрутов на chi router (путь относительно /api/v1).
// Роуты вешаются ПОД auth-middleware (см. cmd/server/main.go).
func (h *PhotoHandler) Mount(r chi.Router) {
	mountPhotoRoutes(r, h.impl())
}

// impl — приватный доступ к реализации по identity-ключу.
func (h *PhotoHandler) impl() *photoHandler {
	if impl, ok := photoHandlers[h]; ok {
		return impl
	}
	return &photoHandler{photos: nil, svc: nil}
}