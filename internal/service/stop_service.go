package service

// StopService — CRUD точек маршрута (CHAT-07: A7-03).
// Контракт: internal/service/interfaces.go (StopService, StopPayload).
// Правила: title non-blank, lat/lng в пределах, max 100 точек → domain.NewValidation;
// чужая/несуществующая поездка или точка — domain.NewNotFound.

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

const maxStopsPerTrip = 100

// stopID — UUID v4 без внешних зависимостей (RFC 4122, вариант 1).
func stopID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// stopService — реализация StopService.
type stopService struct {
	stops repo.StopRepo
	trips repo.TripRepo
	pool  *pgxpool.Pool
}

var _ StopService = (*stopService)(nil)

// NewStopService — конструктор StopService.
func NewStopService(pool *pgxpool.Pool, stops repo.StopRepo, trips repo.TripRepo) StopService {
	return &stopService{pool: pool, stops: stops, trips: trips}
}

// validateStopPayload — валидация полей точки: title non-blank, lat/lng в пределах.
func validateStopPayload(p *StopPayload) *domain.AppError {
	if p == nil || strings.TrimSpace(p.Title) == "" {
		return domain.NewValidation("title is required", map[string]any{"field": "title"})
	}
	if p.Lat < -90 || p.Lat > 90 {
		return domain.NewValidation("lat must be between -90 and 90", map[string]any{"field": "lat"})
	}
	if p.Lng < -180 || p.Lng > 180 {
		return domain.NewValidation("lng must be between -180 and 180", map[string]any{"field": "lng"})
	}
	return nil
}

// checkTripOwner — проверить, что trip принадлежит userID.
func (s *stopService) checkTripOwner(ctx context.Context, userID, tripID string) error {
	_, err := s.trips.Get(ctx, userID, tripID)
	return err
}

// List — все точки маршрута пользователя.
func (s *stopService) List(ctx context.Context, userID, tripID string) ([]*domain.Stop, error) {
	if err := s.checkTripOwner(ctx, userID, tripID); err != nil {
		return nil, err
	}
	return s.stops.ListByTrip(ctx, tripID)
}

// Get — точка по id; чужая/нет → domain.NewNotFound.
func (s *stopService) Get(ctx context.Context, userID, tripID, stopID string) (*domain.Stop, error) {
	if err := s.checkTripOwner(ctx, userID, tripID); err != nil {
		return nil, err
	}
	return s.stops.Get(ctx, tripID, stopID)
}

// Create — новая точка маршрута; валидация → domain.NewValidation.
func (s *stopService) Create(ctx context.Context, userID, tripID string, p *StopPayload) (*domain.Stop, error) {
	if err := validateStopPayload(p); err != nil {
		return nil, err
	}
	if err := s.checkTripOwner(ctx, userID, tripID); err != nil {
		return nil, err
	}

	// Проверка лимита: < 100 точек на поездку
	stops, err := s.stops.ListByTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if len(stops) >= maxStopsPerTrip {
		return nil, domain.NewValidation("too many stops on this trip (max 100)", nil)
	}

	// order: если 0 — взять MaxOrder+1
	order := p.Order
	if order == 0 {
		maxOrder, err := s.stops.MaxOrder(ctx, tripID)
		if err != nil {
			return nil, err
		}
		order = maxOrder + 1
	}

	sStop := &domain.Stop{
		ID:          stopID(),
		TripID:      tripID,
		Title:       strings.TrimSpace(p.Title),
		Description: p.Description,
		Lat:         p.Lat,
		Lng:         p.Lng,
		VisitDate:   p.VisitDate,
		Order:       order,
	}
	if err := s.stops.Insert(ctx, sStop); err != nil {
		return nil, err
	}
	return sStop, nil
}

// Update — обновление точки; валидация + NotFound из репозитория.
func (s *stopService) Update(ctx context.Context, userID, tripID, stopID string, p *StopPayload) (*domain.Stop, error) {
	if err := validateStopPayload(p); err != nil {
		return nil, err
	}
	if err := s.checkTripOwner(ctx, userID, tripID); err != nil {
		return nil, err
	}

	cur, err := s.stops.Get(ctx, tripID, stopID)
	if err != nil {
		return nil, err
	}

	newOrder := p.Order
	if newOrder == 0 {
		newOrder = cur.Order // не меняем, если не указан
	}

	// Если order изменился — Reindex в транзакции
	if newOrder != cur.Order {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return nil, err
		}
		// Reindex сначала, потом Update
		if err := stopReindexInTx(tx, tripID, cur.Order, newOrder); err != nil {
			_ = tx.Rollback(ctx)
			return nil, err
		}
		cur.Title = strings.TrimSpace(p.Title)
		cur.Description = p.Description
		cur.Lat = p.Lat
		cur.Lng = p.Lng
		cur.VisitDate = p.VisitDate
		cur.Order = newOrder
		if err := stopUpdateInTx(tx, cur); err != nil {
			_ = tx.Rollback(ctx)
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return cur, nil
	}

	// Без изменения order — простой update
	cur.Title = strings.TrimSpace(p.Title)
	cur.Description = p.Description
	cur.Lat = p.Lat
	cur.Lng = p.Lng
	cur.VisitDate = p.VisitDate
	if err := s.stops.Update(ctx, cur); err != nil {
		return nil, err
	}
	return cur, nil
}

// Delete — удаление точки; пересчёт order (атомарно в транзакции).
func (s *stopService) Delete(ctx context.Context, userID, tripID, stopID string) error {
	if err := s.checkTripOwner(ctx, userID, tripID); err != nil {
		return err
	}

	cur, err := s.stops.Get(ctx, tripID, stopID)
	if err != nil {
		return err
	}

	// Транзакция: удалить точку, затем сдвинуть order > deletedOrder
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	if err := stopDeleteInTx(tx, tripID, stopID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	// Сдвинуть все с order > deletedOrder на -1
	_, err = tx.Exec(ctx, `UPDATE trip_stops SET "order" = "order" - 1
		WHERE trip_id = $1::uuid AND "order" > $2`, tripID, cur.Order)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// ---------- tx-хелперы для репозитория в транзакции ----------

// stopReindexInTx выполняет Reindex внутри транзакции.
func stopReindexInTx(tx repo.PgTx, tripID string, from, to int) error {
	if from == to {
		return nil
	}
	if from < to {
		_, err := tx.Exec(context.Background(), `UPDATE trip_stops SET "order" = "order" + 1
			WHERE trip_id = $1::uuid AND "order" >= $2 AND "order" < $3`, tripID, to, from)
		return err
	}
	// from > to
	_, err := tx.Exec(context.Background(), `UPDATE trip_stops SET "order" = "order" - 1
			WHERE trip_id = $1::uuid AND "order" > $2 AND "order" <= $3`, tripID, to, from)
	return err
}

// stopUpdateInTx выполняет Update внутри транзакции.
func stopUpdateInTx(tx repo.PgTx, s *domain.Stop) error {
	tag, err := tx.Exec(context.Background(), `UPDATE trip_stops SET title = $1, description = $2, lat = $3, lng = $4, visit_date = $5, "order" = $6
		WHERE id = $7::uuid AND trip_id = $8::uuid`,
		s.Title, s.Description, s.Lat, s.Lng, s.VisitDate, s.Order, s.ID, s.TripID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFound("stop not found")
	}
	return nil
}

// stopDeleteInTx выполняет Delete внутри транзакции.
func stopDeleteInTx(tx repo.PgTx, tripID, id string) error {
	tag, err := tx.Exec(context.Background(), `DELETE FROM trip_stops WHERE trip_id = $1::uuid AND id = $2::uuid`, tripID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFound("stop not found")
	}
	return nil
}
