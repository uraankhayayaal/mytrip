package handlers

// ExportHandler — GET /api/v1/trips/{id}/export (CHAT-07: A7-06).
// Контракт: internal/api/handlers/handlers.go (ExportHandler).
//
// Эндпоинт: GET /trips/{id}/export
//   ok : 200, Content-Type: text/csv, Content-Disposition: attachment; filename="<slugified title>.csv"
//        строка 1: "order,title,description,lat,lng,visit_date"
//        строки N: значения остановки; visit_date — RFC3339.
//   err: 404 (чужая/нет), 500 — через middleware.WriteError.
//
// Реализация: CSV формируется поверх StopService.List (ORDER BY order ASC).

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/service"
)

// exportHeader — заголовочная строка CSV (фиксированный контракт).
const exportHeader = "order,title,description,lat,lng,visit_date"

// defaultExportFilename — имя файла по умолчанию, если title пуст/невозможно slug-изировать.
const defaultExportFilename = "trip.csv"

// NewExportHandler — конструктор ExportHandler (реализация контракта из handlers.go).
func NewExportHandler(trips service.TripService, stops service.StopService) *ExportHandler {
	return &ExportHandler{trips: trips, stops: stops}
}

// Mount — регистрация маршрутов на chi router (путь относительно /api/v1).
// Роут вешается ПОД auth-middleware (см. cmd/server/main.go).
func (h *ExportHandler) Mount(r chi.Router) {
	r.Get("/trips/{id}/export", h.export)
}

// export — GET /trips/{id}/export.
func (h *ExportHandler) export(w http.ResponseWriter, r *http.Request) {
	userID := middleware.AuthContext(r.Context())
	tripID := chi.URLParam(r, "id")

	trip, err := h.trips.Get(r.Context(), userID, tripID)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}

	stops, err := h.stops.List(r.Context(), userID, tripID)
	if err != nil {
		middleware.WriteError(w, toAppError(err))
		return
	}

	var buf bytes.Buffer
	buf.WriteString(exportHeader)
	buf.WriteByte('\n')
	for _, s := range stops {
		buf.WriteString(strconv.Itoa(s.Order))
		buf.WriteByte(',')
		buf.WriteString(csvQuote(s.Title))
		buf.WriteByte(',')
		buf.WriteString(csvQuote(s.Description))
		buf.WriteByte(',')
		buf.WriteString(formatFloat(s.Lat))
		buf.WriteByte(',')
		buf.WriteString(formatFloat(s.Lng))
		buf.WriteByte(',')
		buf.WriteString(s.VisitDate.Format("2006-01-02T15:04:05Z07:00"))
		buf.WriteByte('\n')
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+exportFilename(trip.Title)+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// exportFilename — имя файла: slug-изированный title + ".csv"; если пуст — defaultExportFilename.
func exportFilename(title string) string {
	slug := slugify(title)
	if slug == "" {
		return defaultExportFilename
	}
	return slug + ".csv"
}

// slugify — ASCII-slug: заглавные → строчные, non-alnum → "-", последовательные
// дефисы схлопываются, края обрезаются. Пустой ввод → "".
func slugify(s string) string {
	var run strings.Builder
	dash := true
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			run.WriteRune(r)
			dash = false
		case r >= 'A' && r <= 'Z':
			run.WriteRune(r + ('a' - 'A'))
			dash = false
		default:
			if !dash && run.Len() > 0 {
				run.WriteRune('-')
				dash = true
			}
		}
	}
	return strings.TrimRight(run.String(), "-")
}

// formatFloat — float64 → строка без экспоненты (strconv.FormatFloat 'g' может дать 1e-07;
// используем 'f' с precision 6 — для координат достаточно, без хвостовых нулей).
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}

// csvQuote — экранирует строку для CSV: кавычки удваиваются, само поле обёрнуто
// в кавычки, если содержит "," "\n" "\r" или "\""; иначе возвращается как есть.
func csvQuote(s string) string {
	if !strings.ContainsAny(s, `,\"\n\r`) {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
