package service

// StopService — CRUD точек маршрута поездки (CHAT-07: A7-03).
// Контракт: internal/service/interfaces.go (StopService, StopPayload).
// Правила: валидация (|lat|<=90, |lng|<=180, title non-blank) → domain.NewValidation;
// order==0 при Create → MaxOrder+1; смена order при Update → Reindex
// (в той же транзакции, что и Update).

import (
	"context"
	"strings"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

// stopService — реализация StopService.
type stopService struct {
	stops repo.StopRepo
	trips repo.TripRepo
}

var _ StopService = (*stopService)(nil)

// NewStopService — конструктор StopService.
// trips — для проверки владения поездкой (A7-02).
func NewStopService(stops repo.StopRepo, trips repo.TripRepo) StopService {
	return &stopService{stops: stops, trips: trips}
}

// validateStopPayload — валидация полей точки: title non-blank, |lat|<=90, |lng|<=180.
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

// List — все точки маршрута поездки (в порядке order ASC).
func (s *stopService) List(ctx context.Context, userID, tripID string) ([]*domain.Stop, error) {
	// Проверка владения поездкой (иначе чужие stops доступны по trip_id).
	if _, err := s.trips.Get(ctx, userID, tripID); err != nil {
		return nil, err
	}
	return s.stops.ListByTrip(ctx, tripID)
}

// Get — точка по id; чужая/нет → domain.NewNotFound.
func (s *stopService) Get(ctx context.Context, userID, tripID, stopID string) (*domain.Stop, error) {
	if _, err := s.trips.Get(ctx, userID, tripID); err != nil {
		return nil, err
	}
	return s.stops.Get(ctx, tripID, stopID)
}

// Create — новая точка; при order==0 → MaxOrder+1.
func (s *stopService) Create(ctx context.Context, userID, tripID string, p *StopPayload) (*domain.Stop, error) {
	if err := validateStopPayload(p); err != nil {
		return nil, err
	}
	if _, err := s.trips.Get(ctx, userID, tripID); err != nil {
		return nil, err
	}
	order := p.Order
	if order == 0 {
		max, err := s.stops.MaxOrder(ctx, tripID)
		if err != nil {
			return nil, err
		}
		order = max + 1
	}
	st := &domain.Stop{
		ID:          newStopID(),
		TripID:      tripID,
		Title:       strings.TrimSpace(p.Title),
		Description: p.Description,
		Lat:         p.Lat,
		Lng:         p.Lng,
		VisitDate:   p.VisitDate,
		Order:       order,
	}
	if err := s.stops.Insert(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

// Update — обновление точки; при смене order — Reindex ДО Update.
func (s *stopService) Update(ctx context.Context, userID, tripID, stopID string, p *StopPayload) (*domain.Stop, error) {
	if err := validateStopPayload(p); err != nil {
		return nil, err
	}
	if _, err := s.trips.Get(ctx, userID, tripID); err != nil {
		return nil, err
	}
	cur, err := s.stops.Get(ctx, tripID, stopID)
	if err != nil {
		return nil, err
	}
	newOrder := p.Order
	if newOrder == 0 {
		newOrder = cur.Order
	}
	if newOrder != cur.Order {
		if err := s.stops.Reindex(ctx, tripID, cur.Order, newOrder); err != nil {
			return nil, err
		}
	}
	cur.Title = strings.TrimSpace(p.Title)
	cur.Description = p.Description
	cur.Lat = p.Lat
	cur.Lng = p.Lng
	cur.VisitDate = p.VisitDate
	cur.Order = newOrder
	if err := s.stops.Update(ctx, cur); err != nil {
		return nil, err
	}
	return cur, nil
}

// Delete — удаление точки; чужая/нет → domain.NewNotFound.
func (s *stopService) Delete(ctx context.Context, userID, tripID, stopID string) error {
	if _, err := s.trips.Get(ctx, userID, tripID); err != nil {
		return err
	}
	return s.stops.Delete(ctx, tripID, stopID)
}
