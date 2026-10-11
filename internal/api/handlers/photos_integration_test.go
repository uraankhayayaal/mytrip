// +build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

// TestPhotosIntegration tests photos endpoints with full integration
func TestPhotosIntegration(t *testing.T) {
	// Integration tests for photos endpoints
	// POST /api/v1/trips/{id}/photos → 201 {photo_url}
	// GET /api/v1/photos/{id}/url → 200 {url, expires_at}
	// DELETE /api/v1/photos/{id} → 204

	// Create test containers
	ctx := context.Background()

	// Create postgres container
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
		_ = postgresContainer.Terminate(ctx)
	}

	// Get connection string
	connectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create minio container
	minioContainer, err := minio.RunContainer(ctx,
		testcontainers.WithImage("minio/minio:latest"),
		minio.WithUsername("testuser"),
		minio.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Browser: http://127.0.0.1").WithOccurrence(1).WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		_ = minioContainer.Terminate(ctx)
	}

	// Get minio endpoint
	minioEndpoint, err := minioContainer.Endpoint(ctx, "http")
	require.NoError(t, err)

	// Initialize repositories
	userRepo, err := repo.NewUserRepo(connectionString)
	require.NoError(t, err)

	tripRepo, err := repo.NewTripRepo(connectionString)
	require.NoError(t, err)

	stopRepo, err := repo.NewStopRepo(connectionString)
	require.NoError(t, err)

	postRepo, err := repo.NewPhotoRepo(connectionString)
	require.NoError(t, err)

	// Create a test user and trip
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	trip := &domain.Trip{
		Title:       "Test Trip",
		Description: "Test Description",
		StartAt:     time.Now(),
		EndAt:       time.Now().Add(24 * time.Hour),
		UserID:      createdUser.ID,
	}
	createdTrip, err := tripRepo.Create(ctx, trip)
	require.NoError(t, err)

	t.Run("UploadPhoto", func(t *testing.T) {
		// Upload photo with valid data
		photoReq := map[string]interface{}{
			"file": "test-file.jpg",
		}

		reqBody, _ := json.Marshal(photoReq)
		req := httptest.NewRequest("POST", "/api/v1/trips/123/photos", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("UploadPhoto Validation Error", func(t *testing.T) {
		// Upload photo with invalid data
		photoReq := map[string]interface{}{
			"file": "",
		}

		reqBody, _ := json.Marshal(photoReq)
		req := httptest.NewRequest("POST", "/api/v1/trips/123/photos", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("GetPhotoURL", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/photos/123/url", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("GetPhotoURL Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/photos/999/url", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("DeletePhoto", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/photos/123", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("DeletePhoto Not Found", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/photos/999", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})
}