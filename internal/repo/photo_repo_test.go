// Package repo contains repository implementations and tests
package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"mytrip/internal/domain"
)

// TestPhotoRepo tests the photo repository implementation
func TestPhotoRepo(t *testing.T) {
	t.Parallel()

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
		_ = postgresContainer.Terminate(ctx)
	}

	// Get connection string
	connectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create repositories
	db, err := NewPhotoRepo(connectionString)
	require.NoError(t, err)

	userRepo, err := NewUserRepo(connectionString)
	require.NoError(t, err)

	tripRepo, err := NewTripRepo(connectionString)
	require.NoError(t, err)

	// Create a test user and trip
	user := &domain.User{
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "password123",
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

	t.Run("Create", func(t *testing.T) {
		photo := &domain.Photo{
			TripID: createdTrip.ID,
			URL:    "http://example.com/test.jpg",
		}

		createdPhoto, err := db.Create(ctx, photo)
		require.NoError(t, err)
		assert.NotEmpty(t, createdPhoto.ID)
		assert.Equal(t, photo.URL, createdPhoto.URL)
		assert.Equal(t, createdTrip.ID, createdPhoto.TripID)
		assert.NotEmpty(t, createdPhoto.CreatedAt)
	})

	t.Run("GetByID", func(t *testing.T) {
		// First create a photo to get
		photo := &domain.Photo{
			TripID: createdTrip.ID,
			URL:    "http://example.com/test2.jpg",
		}

		createdPhoto, err := db.Create(ctx, photo)
		require.NoError(t, err)

		gotPhoto, err := db.GetByID(ctx, createdPhoto.ID)
		require.NoError(t, err)
		assert.Equal(t, createdPhoto.ID, gotPhoto.ID)
		assert.Equal(t, photo.URL, gotPhoto.URL)
	})

	t.Run("Delete", func(t *testing.T) {
		// Create a photo to delete
		photo := &domain.Photo{
			TripID: createdTrip.ID,
			URL:    "http://example.com/test3.jpg",
		}

		createdPhoto, err := db.Create(ctx, photo)
		require.NoError(t, err)

		err = db.Delete(ctx, createdPhoto.ID)
		require.NoError(t, err)

		// Verify it's deleted
		gotPhoto, err := db.GetByID(ctx, createdPhoto.ID)
		assert.Error(t, err)
		assert.Nil(t, gotPhoto)
	})
}