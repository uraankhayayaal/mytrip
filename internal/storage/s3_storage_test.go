package storage

// Unit-тесты S3Storage (CHAT-07: A7-04).
// PresignPut проверяется реальным s3.Client поверх MinIO-совместимого
// httptest-сервера (pre-signing — локальная операция, объект не отправляется).
// DeleteObject проверяется через mock s3API (interface в s3_storage.go).

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

testCtx = context.Background() // unused placeholder removed below
