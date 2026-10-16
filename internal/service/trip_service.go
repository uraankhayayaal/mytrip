package service

// TripService — CRUD поездок (CHAT-07: A7-02).
// Контракт: internal/service/interfaces.go (TripService, TripPayload).
// Правила: title non-blank, end_date >= start_date → domain.NewValidation;
// чужая/несуществующая поездка — domain.NewNotFound (поднимается из репозитория).

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

// tripID — UUID v4 без внешних зависимостей (RFC 4122, вариант 1).
// Аналог newUserID (internal/service/auth_service.go).
func tripID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// tripService — реализация TripService.
type tripService struct {
	trips repo.TripRepo
}

var _ TripService = (*tripService)(nil)

// NewTripService — конструктор TripService.
func NewTripService(trips repo.TripRepo) TripService {
	return &tripService{trips: trips}
}

// validatePayload — валидация полей поездки: title non-blank, end_date >= start_date.
func validatePayload(p *TripPayload) *domain.AppError {
	if p == nil || strings.TrimSpace(p.Title) == "" {
		return domain.NewValidation("title is required", map[string]any{"field": "title"})
	}
	if p.EndDate.Before(p.StartDate) {
		return domain.NewValidation("end_date must be after start_date",
			map[string]any{"field": "end_date"})
	}
	return nil
}

// List — все поездки пользователя.
func (s *tripService) List(ctx context.Context, userID string) ([]*domain.Trip, error) {
	return s.trips.ListByUser(ctx, userID)
}

// Get — поездка по id; чужая/нет → domain.NewNotFound.
func (s *tripService) Get(ctx context.Context, userID, tripID string) (*domain.Trip, error) {
	return s.trips.Get(ctx, userID, tripID)
}

// Create — новая поездка; валидация → domain.NewValidation.
func (s *tripService) Create(ctx context.Context, userID string, p *TripPayload) (*domain.Trip, error) {
	if err := validatePayload(p); err != nil {
		return nil, err
	}
	t := &domain.Trip{
		ID:            tripID(),
		UserID:        userID,
		Title:         strings.TrimSpace(p.Title),
		Description:   p.Description,
		StartDate:     p.StartDate,
		EndDate:       p.EndDate,
		CoverPhotoURL: p.CoverPhotoURL,
	}
	if err := s.trips.Insert(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// Update — обновление поездки; валидация + NotFound из репозитория.
func (s *tripService) Update(ctx context.Context, userID, tripID string, p *TripPayload) (*domain.Trip, error) {
	if err := validatePayload(p); err != nil {
		return nil, err
	}
	cur, err := s.trips.Get(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	cur.Title = strings.TrimSpace(p.Title)
	cur.Description = p.Description
	cur.StartDate = p.StartDate
	cur.EndDate = p.EndDate
	cur.CoverPhotoURL = p.CoverPhotoURL
	if err := s.trips.Update(ctx, cur); err != nil {
		return nil, err
	}
	return cur, nil
}

// Delete — удаление поездки; чужая/нет → domain.NewNotFound.
func (s *tripService) Delete(ctx context.Context, userID, tripID string) error {
	return s.trips.Delete(ctx, userID, tripID)
}
