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

// S3Storage — реализация StoragePort поверх s3.Client.
type S3Storage struct {
	client  *s3.Client
	bucket  string
	ttl     time.Duration
}

var _ service.StoragePort = (*S3Storage)(nil)

// NewS3Storage — конструктор. ttl <= 0 → presignTTL.
func NewS3Storage(client *s3.Client, bucket string, ttl time.Duration) *S3Storage {
	if ttl <= 0 {
		ttl = presignTTL
	}
	return &S3Storage{client: client, bucket: bucket, ttl: ttl}
}

// PresignPut — pre-signed PUT URL для objectKey; срок действия = ttl от момента вызова.
func (s *S3Storage) PresignPut(ctx context.Context, objectKey string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.ttl)
	presigner := s3.NewPresignClient(s.client)
	out, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}, func(o *s3.PresignOptions) { o.Expires = s.ttl })
	if err != nil {
		return "", time.Time{}, err
	}
	return out.URL, expiresAt, nil
}

// DeleteObject — удаление объекта; NoSuchKey и аналогичные → не ошибка (идемпотентность).
func (s *S3Storage) DeleteObject(ctx context.Context, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err == nil {
		return nil
	}
	// S3: удаление несуществующего объекта — не ошибка (PUT-объект мог не дойти).
	var apiErr *types.NoSuchKey
	var nf *types.NotFound
	if errors.As(err, &apiErr) || errors.As(err, &nf) {
		return nil
	}
	var ae smithy.APIError
	if errors.As(err, &ae) {
		code := ae.ErrorCode()
		if code == "NoSuchKey" || code == "NotFound" {
			return nil
		}
	}
	return err
}
