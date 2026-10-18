package service

// PhotoService — управление фотографиями поездки (CHAT-07: A7-04).
// Контракт: internal/service/interfaces.go (PhotoService, StoragePort).
// Правила: владение поездкой проверяется через TripRepo.Get (чужая/нет → domain.NewNotFound);
// при cover=true — TripRepo.SetCover; при Delete — удаление объекта из хранилища
// (ошибка хранилища не блокирует удаление записи).

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

// photoID — UUID v4 без внешних зависимостей (RFC 4122, вариант 1).
func photoID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// photoService — реализация PhotoService.
type photoService struct {
	trips   repo.TripRepo
	photos  repo.PhotoRepo
	storage StoragePort
}

var _ PhotoService = (*photoService)(nil)

// NewPhotoService — конструктор PhotoService.
func NewPhotoService(trips repo.TripRepo, photos repo.PhotoRepo, storage StoragePort) PhotoService {
	return &photoService{trips: trips, photos: photos, storage: storage}
}

// List — фото поездки; чужая/нет → domain.NewNotFound из TripRepo.Get.
func (s *photoService) List(ctx context.Context, userID, tripID string) ([]*domain.Photo, error) {
	if _, err := s.trips.Get(ctx, userID, tripID); err != nil {
		return nil, err
	}
	return s.photos.ListByTrip(ctx, tripID)
}

// Create — фото поездки; владение проверяется через TripRepo.Get.
// При cover=true — TripRepo.SetCover(userID, tripID, url) ДО вставки фото
// (SetCover вернёт NotFound, если поездка вдруг исчезнет).
func (s *photoService) Create(ctx context.Context, userID, tripID, objectKey, url string, cover bool) (*domain.Photo, error) {
	if _, err := s.trips.Get(ctx, userID, tripID); err != nil {
		return nil, err
	}
	if cover {
		if err := s.trips.SetCover(ctx, userID, tripID, url); err != nil {
			return nil, err
		}
	}
	p := &domain.Photo{
		ID:        photoID(),
		TripID:    tripID,
		URL:       url,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.photos.Insert(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete — удаление фото: запись из БД + объект из хранилища.
// Ошибка DeleteObject не блокирует удаление записи (объект мог уже отсутствовать;
// сироту уберёт ливер-джоб).
func (s *photoService) Delete(ctx context.Context, userID, tripID, photoID string) error {
	if _, err := s.trips.Get(ctx, userID, tripID); err != nil {
		return err
	}
	// URL объекта берём из БД ДО удаления записи.
	photos, err := s.photos.ListByTrip(ctx, tripID)
	if err != nil {
		return err
	}
	var url string
	found := false
	for _, p := range photos {
		if p.ID == photoID {
			url = p.URL
			found = true
			break
		}
	}
	if !found {
		return domain.NewNotFound("photo not found")
	}
	if err := s.photos.Delete(ctx, tripID, photoID); err != nil {
		return err
	}
	if s.storage != nil {
		if err := s.storage.DeleteObject(ctx, url); err != nil {
			// Не критично: логирование недоступно в слое сервиса; сироту уберёт ливер-джоб.
			_ = err
		}
	}
	return nil
}
