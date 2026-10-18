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
	"strings"

	"github.com/go-chi/chi/v5"

	"mytrip/internal/api/middleware"
	"mytrip/internal/domain"
	"mytrip/internal/service"
)

// ExportHeader — заголовочная строка CSV (фиксированный контракт).
const ExportHeader = "order,title,description,lat,lng,visit_date"

// DefaultExportFilename — имя файла по умолчанию, если title пуст/невозможно slug-изировать.
const DefaultExportFilename = "trip.csv"

// ExportHandler — реализация контракта из handlers.go.
type ExportHandler struct {
	trips service.TripService
	stops service.StopService
}

// NewExportHandler — конструктор ExportHandler.
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
	buf.WriteString(ExportHeader)
	buf.WriteByte('\n')
	for _, s := range stops {
		buf.WriteString(csvField(s.Order))
		buf.WriteByte(',')
		buf.WriteString(csvQuote(s.Title))
		buf.WriteByte(',')
		buf.WriteString(csvQuote(s.Description))
		buf.WriteByte(',')
		buf.WriteString(floatField(s.Location.Lat))
		buf.WriteByte(',')
		buf.WriteString(floatField(s.Location.Lng))
		buf.WriteByte(',')
		buf.WriteString(s.VisitDate.Format("2006-01-02T15:04:05Z07:00"))
		buf.WriteByte('\n')
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+exportFilename(trip.Title)+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// exportFilename — имя файла: slug-изированный title + ".csv"; если пуст — DefaultExportFilename.
func exportFilename(title string) string {
	slug := slugify(title)
	if slug == "" {
		return DefaultExportFilename
	}
	return slug + ".csv"
}

// slugify — транслитерация ASCII-совместимого slug'а: заглавные в нижние, пробелы/не-alnum → "-",
// последовательные дефисы схлопываются, края обрезаются. Пустой ввод → "".
func slugify(s string) string {
	run := strings.Builder{}
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
	out := strings.TrimRight(run.String(), "-")
	return out
}

// csvField — int → строка (без кавычек).
func csvField(n int) string {
	b := make([]byte, 0, 12)
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

// floatField — float64 → строка в минимальном каноническом виде (без экспоненты, без хвостового ".0").
func floatField(f float64) string {
	b := make([]byte, 0, 24)
	neg := f < 0
	if neg {
		f = -f
	}
	ip := int64(f)
	fr := f - float64(ip)
	// Округляем дробную часть до 12 знаков.
	const digits = 12
	frac := make([]byte, 0, digits)
	for i := 0; i < digits; i++ {
		fr *= 10
		d := int64(fr)
		if d < 0 {
			d = 0
		}
		frac = append(frac, byte('0'+d))
		fr -= float64(d)
		if fr < 1e-9 {
			break
		}
	}
	// Сокращаем хвостовые нули.
	end := len(frac)
	for end > 0 && frac[end-1] == '0' {
		end--
	}
	if neg {
		b = append(b, '-')
	}
	for _, c := range strconvItoa(ip) {
		b = append(b, c)
	}
	if end > 0 {
		b = append(b, '.')
		b = append(b, frac[:end]...)
	}
	return string(b)
}

// strconvItoa — int64 → []byte (без импорта strconv ради компактности).
func strconvItoa(n int64) []byte {
	if n == 0 {
		return []byte("0")
	}
	neg := n < 0
	if neg {
		n = -n
	}
	b := make([]byte, 0, 20)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return b
}

// csvQuote — экранирует строку для CSV: кавычки удваиваются, само поле
// обёрнуто в кавычки если содержит "," "\n" "\r" или "\""; иначе возвращается как есть.
func csvQuote(s string) string {
	if !strings.ContainsAny(s, `,\"\n\r`) {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// toAppError — оборачивает произвольную ошибку в *domain.AppError (500 при не-AppError).
func toAppError(err error) *domain.AppError {
	if ae, ok := err.(*domain.AppError); ok {
		return ae
	}
	return domain.NewInternal()
}
