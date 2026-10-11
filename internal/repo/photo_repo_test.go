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
	"mytrip/internal/repo"
)

// TestPhotoRepo tests the photo repository implementation with testcontainers
func TestPhotoRepo(t *testing.T) {
	t.Parallel()

	// Start a test PostgreSQL container
	ctx := context.Background()
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1).WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, postgresContainer.Terminate(ctx))
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create repositories
	userRepo, err := repo.NewUserRepo(connStr)
	require.NoError(t, err)

	tripRepo, err := repo.NewTripRepo(connStr)
	require.NoError(t, err)

	photoRepo, err := repo.NewPhotoRepo(connStr)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()

		// Create a user first
		user := &domain.User{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		error := userRepo.Create(ctx, user)
		assert.NoError(t, error)

		// Create a trip
		trip := &domain.Trip{
			Title:       "Test Trip",
			Description: "Test Description",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 1),
			UserID:      user.ID,
		}

		error = tripRepo.Create(ctx, trip)
		assert.NoError(t, error)

		// Create a photo
		photo := &domain.Photo{
			TripID: trip.ID,
			URL:    "http://example.com/photo.jpg",
		}

		error = photoRepo.Create(ctx, photo)
		assert.NoError(t, error)
		assert.NotEmpty(t, photo.ID)

		// Verify photo was created
		foundPhoto, err := photoRepo.GetByID(ctx, photo.ID)
		assert.NoError(t, err)
		assert.Equal(t, photo.URL, foundPhoto.URL)
		assert.Equal(t, photo.TripID, foundPhoto.TripID)
	})

	t.Run("GetByID", func(t *testing.T) {
		ctx := context.Background()

		// Create a user first
		user := &domain.User{
			Email:    "test2@example.com",
			Password: "password123",
			Name:     "Test User 2",
		}

		error := userRepo.Create(ctx, user)
		assert.NoError(t, error)

		// Create a trip
		trip := &domain.Trip{
			Title:       "Test Trip 2",
			Description: "Test Description 2",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 1),
			UserID:      user.ID,
		}

		error = tripRepo.Create(ctx, trip)
		assert.NoError(t, error)

		// Create a photo
		photo := &domain.Photo{
			TripID: trip.ID,
			URL:    "http://example.com/photo2.jpg",
		}

		error = photoRepo.Create(ctx, photo)
		assert.NoError(t, error)

		// Get photo by ID
		foundPhoto, err := photoRepo.GetByID(ctx, photo.ID)
		assert.NoError(t, err)
		assert.Equal(t, photo.URL, foundPhoto.URL)
		assert.Equal(t, photo.TripID, foundPhoto.TripID)
	})

	t.Run("Delete", func(t *testing.T) {
		ctx := context.Background()

		// Create a user first
		user := &domain.User{
			Email:    "test3@example.com",
			Password: "password123",
			Name:     "Test User 3",
		}

		error := userRepo.Create(ctx, user)
		assert.NoError(t, error)

		// Create a trip
		trip := &domain.Trip{
			Title:       "Test Trip 3",
			Description: "Test Description 3",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 1),
			UserID:      user.ID,
		}

		error = tripRepo.Create(ctx, trip)
		assert.NoError(t, error)

		// Create a photo
		photo := &domain.Photo{
			TripID: trip.ID,
			URL:    "http://example.com/photo3.jpg",
		}

		error = photoRepo.Create(ctx, photo)
		assert.NoError(t, error)

		// Delete photo
		error = photoRepo.Delete(ctx, photo.ID)
		assert.NoError(t, error)

		// Verify deletion
		_, err := photoRepo.GetByID(ctx, photo.ID)
		assert.Error(t, err)
	})
}