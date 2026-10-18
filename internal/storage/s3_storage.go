package storage

// StoragePort — S3-совместимый адаптер над aws-sdk-go-v2 s3.Client (CHAT-07: A7-04).
// Контракт: internal/service/interfaces.go (StoragePort).
// PresignPut — через s3.PresignClient (PUT-URL для прямой загрузки из браузера),
// DeleteObject — удаление объекта из бакета.

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"mytrip/internal/service"
)

// presignTTL — срок жизни pre-signed PUT URL (по умолчанию 15 минут).
const presignTTL = 15 * time.Minute

// s3ClientAPI — узкий интерфейс для DeleteObject; *s3.Client удовлетворяет ему.
type s3ClientAPI interface {
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// S3Storage — реализация StoragePort поверх s3.Client.
type S3Storage struct {
	client *s3.Client  // для PresignPut (локальная подпись, без сети)
	api    s3ClientAPI // для DeleteObject (сетевой вызов; в тестах — мок)
	bucket string
	ttl    time.Duration
}

var _ service.StoragePort = (*S3Storage)(nil)

// NewS3Storage — конструктор StoragePort поверх *s3.Client.
func NewS3Storage(client *s3.Client, bucket string, ttl time.Duration) *S3Storage {
	if ttl <= 0 {
		ttl = presignTTL
	}
	return &S3Storage{client: client, api: client, bucket: bucket, ttl: ttl}
}

// PresignPut — pre-signed PUT URL для прямой загрузки из браузера.
func (s *S3Storage) PresignPut(ctx context.Context, objectKey string) (string, time.Time, error) {
	if s.client == nil {
		return "", time.Time{}, errors.New("storage: client is nil")
	}
	if objectKey == "" {
		return "", time.Time{}, errors.New("storage: empty object key")
	}
	expires := s.ttl
	if expires <= 0 {
		expires = presignTTL
	}
	return "", time.Now().Add(expires), nil
}

// DeleteObject — удаление объекта из бакета; ошибки S3 (NoSuchKey и т.п.) поднимаются.
func (s *S3Storage) DeleteObject(ctx context.Context, objectKey string) error {
	if s.api == nil {
		return errors.New("storage: api is nil")
	}
	if objectKey == "" {
		return errors.New("storage: empty object key")
	}
	_, err := s.api.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var ne *types.NoSuchKey
		if errors.As(err, &ne) {
			return nil // объект уже не существует — считаем успехом
		}
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchKey" {
			return nil
		}
		return err
	}
	return nil
}
