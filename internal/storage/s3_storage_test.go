package storage

// Unit-тесты S3Storage (CHAT-07: A7-04).
// PresignPut — реальным s3.Client (pre-signing — локальная операция, объект
// не отправляется; endpoint можно задать произвольно).
// DeleteObject — через mock s3ClientAPI (узкий интерфейс в s3_storage.go).

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

// s3ClientMock — мок s3ClientAPI: только DeleteObject (используется в тестах).
type s3ClientMock struct {
	deleteFn func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	lastKey  string
	lastBkt  string
}

func (m *s3ClientMock) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.lastKey = aws.ToString(params.Key)
	m.lastBkt = aws.ToString(params.Bucket)
	return m.deleteFn(ctx, params, optFns...)
}

func TestNewS3Storage_DefaultTTL(t *testing.T) {
	s := NewS3Storage(nil, "bkt", 0)
	if s.ttl != presignTTL {
		t.Fatalf("ttl = %s, want %s", s.ttl, presignTTL)
	}
	s2 := NewS3Storage(nil, "bkt", -time.Second)
	if s2.ttl != presignTTL {
		t.Fatalf("ttl (negative) = %s, want %s", s2.ttl, presignTTL)
	}
}

func TestPresignPut_ReturnsSignedURL(t *testing.T) {
	cfg := aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("AK", "SK", ""),
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.BaseEndpoint = aws.String("http://127.0.0.1:9") })
	s := NewS3Storage(client, "photos", time.Minute)

	url, expiresAt, err := s.PresignPut(context.Background(), "trip1/pic.jpg")
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	if !strings.HasPrefix(url, "http://127.0.0.1:9/photos/trip1%2Fpic.jpg") ||
		!strings.Contains(url, "X-Amz-Signature") ||
		!strings.Contains(url, "X-Amz-Expires=60") {
		t.Fatalf("unexpected presigned URL: %s", url)
	}
	if time.Until(expiresAt) <= 0 || time.Until(expiresAt) > time.Minute+time.Second {
		t.Fatalf("expiresAt out of range: %s (now+1m expected)", expiresAt)
	}
}

func TestDeleteObject_Success(t *testing.T) {
	m := &s3ClientMock{deleteFn: func(ctx context.Context, p *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
		return &s3.DeleteObjectOutput{}, nil
	}}
	s := &S3Storage{api: m, bucket: "photos"}
	if err := s.DeleteObject(context.Background(), "trip1/a.jpg"); err != nil {
		t.Fatalf("DeleteObject: %v", err)
	}
	if m.lastKey != "trip1/a.jpg" || m.lastBkt != "photos" {
		t.Fatalf("unexpected args: key=%q bucket=%q", m.lastKey, m.lastBkt)
	}
}

func TestDeleteObject_NoSuchKeyIsIdempotent(t *testing.T) {
	m := &s3ClientMock{deleteFn: func(ctx context.Context, p *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
		return nil, &smithy.GenericAPIError{Code: "NoSuchKey", Message: "no such key"}
	}}
	s := &S3Storage{api: m, bucket: "photos"}
	if err := s.DeleteObject(context.Background(), "missing.jpg"); err != nil {
		t.Fatalf("want nil for NoSuchKey, got: %v", err)
	}
}

func TestDeleteObject_OtherErrorPropagates(t *testing.T) {
	boom := errors.New("connection refused")
	m := &s3ClientMock{deleteFn: func(ctx context.Context, p *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
		return nil, boom
	}}
	s := &S3Storage{api: m, bucket: "photos"}
	err := s.DeleteObject(context.Background(), "x.jpg")
	if !errors.Is(err, boom) {
		t.Fatalf("want propagated error, got: %v", err)
	}
}
