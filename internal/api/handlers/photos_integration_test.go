//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/wait"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

func TestPhotosIntegration(t *testing.T) {
	// Integration tests for photos endpoints
	// POST /api/v1/trips/{id}/photos → 201 {photo_url}
	// GET /api/v1/photos/{id}/url → 200 {url, expires_at}
	// DELETE /api/v1/photos/{id} → 204

	// Create a test database container
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1).WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// Create a test minio container
	minioContainer, err := minio.RunContainer(ctx,
		testcontainers.WithImage("minio/minio:latest"),
		minio.WithUsername("minioadmin"),
		minio.WithPassword("minioadmin"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Started MinIO"),
		),
	)
	require.NoError(t, err)
	defer func() {
		err := minioContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// Get connection strings
	postgresConnectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	minioURL, err := minioContainer.Endpoint(ctx)
	require.NoError(t, err)

	// For now we'll test the structure
	
	t.Run("UploadPhoto - Success", func(t *testing.T) {
		t.Skip("Not implemented yet")
	})

	t.Run("UploadPhoto - Validation Error", func(t *testing.T) {
		t.Skip("Not implemented yet")
	})

	t.Run("GetPhotoURL - Success", func(t *testing.T) {
		t.Skip("Not implemented yet")
	})

	t.Run("GetPhotoURL - Not Found", func(t *testing.T) {
		t.Skip("Not implemented yet")
	})

	t.Run("DeletePhoto - Success", func(t *testing.T) {
		t.Skip("Not implemented yet")
	})

	t.Run("DeletePhoto - Not Found", func(t *testing.T) {
		t.Skip("Not implemented yet")
	})
}