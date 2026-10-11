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

// TestTripRepo tests the trip repository implementation with testcontainers
func TestTripRepo(t *testing.T) {
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

	// Create repository
	userRepo, err := repo.NewUserRepo(connStr)
	require.NoError(t, err)

	tripRepo, err := repo.NewTripRepo(connStr)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()

		user := &domain.User{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		error := userRepo.Create(ctx, user)
		assert.NoError(t, error)

		trip := &domain.Trip{
			Title:       "Test Trip",
			Description: "Test Description",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 1),
			UserID:      user.ID,
		}

		error = tripRepo.Create(ctx, trip)
		assert.NoError(t, error)
		assert.NotEmpty(t, trip.ID)

		// Verify trip was created
		foundTrip, err := tripRepo.GetByID(ctx, trip.ID)
		assert.NoError(t, err)
		assert.Equal(t, trip.Title, foundTrip.Title)
		assert.Equal(t, trip.Description, foundTrip.Description)
		assert.Equal(t, trip.UserID, foundTrip.UserID)
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

		// Get trip by ID
		foundTrip, err := tripRepo.GetByID(ctx, trip.ID)
		assert.NoError(t, err)
		assert.Equal(t, trip.Title, foundTrip.Title)
		assert.Equal(t, trip.Description, foundTrip.Description)
		assert.Equal(t, trip.UserID, foundTrip.UserID)
	})

	t.Run("ListByUser", func(t *testing.T) {
		ctx := context.Background()

		// Create a user first
		user := &domain.User{
			Email:    "test3@example.com",
			Password: "password123",
			Name:     "Test User 3",
		}

		error := userRepo.Create(ctx, user)
		assert.NoError(t, error)

		// Create trips for this user
		trip1 := &domain.Trip{
			Title:       "Test Trip 1",
			Description: "Test Description 1",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 1),
			UserID:      user.ID,
		}

		trip2 := &domain.Trip{
			Title:       "Test Trip 2",
			Description: "Test Description 2",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 1),
			UserID:      user.ID,
		}

		error = tripRepo.Create(ctx, trip1)
		assert.NoError(t, error)

		error = tripRepo.Create(ctx, trip2)
		assert.NoError(t, error)

		// List trips by user
		trips, err := tripRepo.ListByUser(ctx, user.ID)
		assert.NoError(t, err)
		assert.Len(t, trips, 2)
	})

	t.Run("Update", func(t *testing.T) {
		ctx := context.Background()

		// Create a user first
		user := &domain.User{
			Email:    "test4@example.com",
			Password: "password123",
			Name:     "Test User 4",
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

		// Update trip
		trip.Title = "Updated Trip Title"
		trip.Description = "Updated Description"

		error = tripRepo.Update(ctx, trip)
		assert.NoError(t, error)

		// Verify update
		foundTrip, err := tripRepo.GetByID(ctx, trip.ID)
		assert.NoError(t, err)
		assert.Equal(t, trip.Title, foundTrip.Title)
		assert.Equal(t, trip.Description, foundTrip.Description)
	})

	t.Run("Delete", func(t *testing.T) {
		ctx := context.Background()

		// Create a user first
		user := &domain.User{
			Email:    "test5@example.com",
			Password: "password123",
			Name:     "Test User 5",
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

		// Delete trip
		error = tripRepo.Delete(ctx, trip.ID)
		assert.NoError(t, error)

		// Verify deletion
		_, err := tripRepo.GetByID(ctx, trip.ID)
		assert.Error(t, err)
	})
}