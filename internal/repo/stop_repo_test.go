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

// TestStopRepo tests the stop repository implementation with testcontainers
func TestStopRepo(t *testing.T) {
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

	stopRepo, err := repo.NewStopRepo(connStr)
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

		// Create a stop
		stop := &domain.Stop{
			TripID:      trip.ID,
			Name:        "Test Stop",
			Description: "Test Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		error = stopRepo.Create(ctx, stop)
		assert.NoError(t, error)
		assert.NotEmpty(t, stop.ID)

		// Verify stop was created
		foundStop, err := stopRepo.GetByID(ctx, stop.ID)
		assert.NoError(t, err)
		assert.Equal(t, stop.Name, foundStop.Name)
		assert.Equal(t, stop.Description, foundStop.Description)
		assert.Equal(t, stop.TripID, foundStop.TripID)
		assert.Equal(t, stop.Location.Lat, foundStop.Location.Lat)
		assert.Equal(t, stop.Location.Lng, foundStop.Location.Lng)
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

		// Create a stop
		stop := &domain.Stop{
			TripID:      trip.ID,
			Name:        "Test Stop 2",
			Description: "Test Description 2",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		error = stopRepo.Create(ctx, stop)
		assert.NoError(t, error)

		// Get stop by ID
		foundStop, err := stopRepo.GetByID(ctx, stop.ID)
		assert.NoError(t, err)
		assert.Equal(t, stop.Name, foundStop.Name)
		assert.Equal(t, stop.Description, foundStop.Description)
		assert.Equal(t, stop.TripID, foundStop.TripID)
	})

	t.Run("ListByTrip", func(t *testing.T) {
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

		// Create stops for this trip
		stop1 := &domain.Stop{
			TripID:      trip.ID,
			Name:        "Stop 1",
			Description: "Description 1",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		stop2 := &domain.Stop{
			TripID:      trip.ID,
			Name:        "Stop 2",
			Description: "Description 2",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		error = stopRepo.Create(ctx, stop1)
		assert.NoError(t, error)

		error = stopRepo.Create(ctx, stop2)
		assert.NoError(t, error)

		// List stops by trip
		stops, err := stopRepo.ListByTrip(ctx, trip.ID)
		assert.NoError(t, err)
		assert.Len(t, stops, 2)
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
			Title:       "Test Trip 4",
			Description: "Test Description 4",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 1),
			UserID:      user.ID,
		}

		error = tripRepo.Create(ctx, trip)
		assert.NoError(t, error)

		// Create a stop
		stop := &domain.Stop{
			TripID:      trip.ID,
			Name:        "Test Stop",
			Description: "Test Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		error = stopRepo.Create(ctx, stop)
		assert.NoError(t, error)

		// Update stop
		stop.Name = "Updated Stop Name"
		stop.Description = "Updated Description"

		error = stopRepo.Update(ctx, stop)
		assert.NoError(t, error)

		// Verify update
		foundStop, err := stopRepo.GetByID(ctx, stop.ID)
		assert.NoError(t, err)
		assert.Equal(t, stop.Name, foundStop.Name)
		assert.Equal(t, stop.Description, foundStop.Description)
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
			Title:       "Test Trip 5",
			Description: "Test Description 5",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 0, 1),
			UserID:      user.ID,
		}

		error = tripRepo.Create(ctx, trip)
		assert.NoError(t, error)

		// Create a stop
		stop := &domain.Stop{
			TripID:      trip.ID,
			Name:        "Test Stop",
			Description: "Test Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		error = stopRepo.Create(ctx, stop)
		assert.NoError(t, error)

		// Delete stop
		error = stopRepo.Delete(ctx, stop.ID)
		assert.NoError(t, error)

		// Verify deletion
		_, err := stopRepo.GetByID(ctx, stop.ID)
		assert.Error(t, err)
	})
}