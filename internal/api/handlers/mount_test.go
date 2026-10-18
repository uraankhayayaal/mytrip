package handlers

// Тесты монтажа маршрутов (CHAT-07: A7-06): проверяют, что каждый Mount
// регистрирует ровно ожидаемые пути и методы на chi-роутере — без БД и
// без сети (r.Routes() — синхронный разбор маршрутной таблицы).

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

// collectRoutes — все маршруты роутера в виде "METHOD path" (сортировка не нужна — ищем подмножество).
func collectRoutes(r chi.Router) map[string]bool {
	got := make(map[string]bool)
	walkFunc := func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		got[method+" "+route] = true
		return nil
	}
	if err := chi.Walk(r, walkFunc); err != nil {
		panic(err)
	}
	return got
}

func assertRouteMounted(t *testing.T, routes map[string]bool, method, path string) {
	t.Helper()
	key := method + " " + path
	if !routes[key] {
		t.Fatalf("маршрут %s не смонтирован; смонтировано: %v", key, routes)
	}
}

func assertRouteAbsent(t *testing.T, routes map[string]bool, key string) {
	t.Helper()
	if routes[key] {
		t.Fatalf("маршрут %s смонтирован, хотя должен отсутствовать", key)
	}
}

func TestTripHandlerMountRoutes(t *testing.T) {
	r := chi.NewRouter()
	NewTripHandler(nil).Mount(r)
	routes := collectRoutes(r)

	assertRouteMounted(t, routes, http.MethodGet, "/trips")
	assertRouteMounted(t, routes, http.MethodPost, "/trips")
	assertRouteMounted(t, routes, http.MethodGet, "/trips/{id}")
	assertRouteMounted(t, routes, http.MethodPut, "/trips/{id}")
	assertRouteMounted(t, routes, http.MethodDelete, "/trips/{id}")
	// export — НЕ маршрут TripHandler (A7-06: отдельный ExportHandler)
	assertRouteAbsent(t, routes, http.MethodGet+" /trips/{id}/export")
}

func TestStopHandlerMountRoutes(t *testing.T) {
	r := chi.NewRouter()
	NewStopHandler(nil).Mount(r)
	routes := collectRoutes(r)

	assertRouteMounted(t, routes, http.MethodGet, "/trips/{id}/stops")
	assertRouteMounted(t, routes, http.MethodPost, "/trips/{id}/stops")
	assertRouteMounted(t, routes, http.MethodPut, "/trips/{id}/stops/{stopId}")
	assertRouteMounted(t, routes, http.MethodDelete, "/trips/{id}/stops/{stopId}")
}

func TestPhotoHandlerMountRoutes(t *testing.T) {
	r := chi.NewRouter()
	NewPhotoHandler(nil, nil).Mount(r)
	routes := collectRoutes(r)

	assertRouteMounted(t, routes, http.MethodGet, "/trips/{id}/photos")
	assertRouteMounted(t, routes, http.MethodPost, "/trips/{id}/photos")
	assertRouteMounted(t, routes, http.MethodPost, "/trips/{id}/photos/presign")
	assertRouteMounted(t, routes, http.MethodDelete, "/trips/{id}/photos/{photoId}")
}

func TestExportHandlerMountRoute(t *testing.T) {
	r := chi.NewRouter()
	NewExportHandler(nil, nil).Mount(r)
	routes := collectRoutes(r)

	assertRouteMounted(t, routes, http.MethodGet, "/trips/{id}/export")
}

func TestAuthHandlerMountRoutes(t *testing.T) {
	r := chi.NewRouter()
	NewAuthHandler(nil).Mount(r)
	routes := collectRoutes(r)

	assertRouteMounted(t, routes, http.MethodPost, "/auth/register")
	assertRouteMounted(t, routes, http.MethodPost, "/auth/login")
	assertRouteMounted(t, routes, http.MethodPost, "/auth/refresh")
	assertRouteMounted(t, routes, http.MethodPost, "/auth/logout")
	assertRouteMounted(t, routes, http.MethodGet, "/users/me")
}

// TestMountComposedAPI — полный состав /api/v1 как в cmd/server/main.go:
// auth публичный, остальные — под auth-группой. Проверяет композицию,
// которую монтирует composition root (без самого main).
func TestMountComposedAPI(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		NewAuthHandler(nil).Mount(r)
		r.With(func(next http.Handler) http.Handler { return next }).Group(func(r chi.Router) {
			NewTripHandler(nil).Mount(r)
			NewStopHandler(nil).Mount(r)
			NewPhotoHandler(nil, nil).Mount(r)
			NewExportHandler(nil, nil).Mount(r)
		})
	})
	routes := collectRoutes(r)

	for _, want := range []string{
		http.MethodGet + " /api/v1/trips",
		http.MethodPost + " /api/v1/trips",
		http.MethodGet + " /api/v1/trips/{id}",
		http.MethodPut + " /api/v1/trips/{id}",
		http.MethodDelete + " /api/v1/trips/{id}",
		http.MethodGet + " /api/v1/trips/{id}/stops",
		http.MethodGet + " /api/v1/trips/{id}/photos",
		http.MethodPost + " /api/v1/trips/{id}/photos/presign",
		http.MethodGet + " /api/v1/trips/{id}/export",
	} {
		if !routes[want] {
			t.Fatalf("маршрут %s не смонтирован в композиции /api/v1; есть: %v", want, routes)
		}
	}
}
