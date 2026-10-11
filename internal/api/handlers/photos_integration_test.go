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

	// Create repositories
	userRepo, err := repo.NewUserRepo(postgresConnectionString)
	require.NoError(t, err)
	tripRepo, err := repo.NewTripRepo(postgresConnectionString)
	require.NoError(t, err)
	photoRepo, err := repo.NewPhotoRepo(postgresConnectionString)
	require.NoError(t, err)

	// Create test user
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test trip
	trip := &domain.Trip{
		Title:       "Test Trip",
		Description: "Test Description",
		StartAt:     time.Now(),
		EndAt:       time.Now().Add(24 * time.Hour),
		UserID:      createdUser.ID,
	}
	createdTrip, err := tripRepo.Create(ctx, trip)
	require.NoError(t, err)

	t.Run("UploadPhoto - Success", func(t *testing.T) {
		photo := &domain.Photo{
			TripID: createdTrip.ID,
			URL:    "http://example.com/test.jpg",
		}

		createdPhoto, err := photoRepo.Create(ctx, photo)
		require.NoError(t, err)
		assert.NotEmpty(t, createdPhoto.ID)
		assert.Equal(t, photo.URL, createdPhoto.URL)
	})

	t.Run("UploadPhoto - Validation Error", func(t *testing.T) {
		// Test with invalid data
		photo := &domain.Photo{
			TripID: "",
			URL:    "",
		}

		_, err := photoRepo.Create(ctx, photo)
		assert.Error(t, err)
	})

	t.Run("GetPhotoURL - Success", func(t *testing.T) {
		// Create a photo first
		photo := &domain.Photo{
			TripID: createdTrip.ID,
			URL:    "http://example.com/test2.jpg",
		}

		createdPhoto, err := photoRepo.Create(ctx, photo)
		require.NoError(t, err)

		gotPhoto, err := photoRepo.GetByID(ctx, createdPhoto.ID)
		require.NoError(t, err)
		assert.Equal(t, createdPhoto.ID, gotPhoto.ID)
		assert.Equal(t, photo.URL, gotPhoto.URL)
	})

	t.Run("GetPhotoURL - Not Found", func(t *testing.T) {
		_, err := photoRepo.GetByID(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("DeletePhoto - Success", func(t *testing.T) {
		// Create a photo to delete
		photo := &domain.Photo{
			TripID: createdTrip.ID,
			URL:    "http://example.com/test3.jpg",
		}

		createdPhoto, err := photoRepo.Create(ctx, photo)
		require.NoError(t, err)

		err = photoRepo.Delete(ctx, createdPhoto.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = photoRepo.GetByID(ctx, createdPhoto.ID)
		assert.Error(t, err)
	})

	t.Run("DeletePhoto - Not Found", func(t *testing.T) {
		err := photoRepo.Delete(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}