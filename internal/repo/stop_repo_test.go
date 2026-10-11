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

// TestStopRepo tests the stop repository implementation
func TestStopRepo(t *testing.T) {
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
	// Исправлено: добавлено defer с вызовом функции
	defer func() {
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// Get connection string
	connectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create repositories
	db, err := NewStopRepo(connectionString)
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
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Name:        "Test Stop",
			Description: "Test Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		createdStop, err := db.Create(ctx, stop)
		require.NoError(t, err)
		assert.NotEmpty(t, createdStop.ID)
		assert.Equal(t, stop.Name, createdStop.Name)
		assert.Equal(t, stop.Description, createdStop.Description)
		assert.Equal(t, createdTrip.ID, createdStop.TripID)
		assert.NotEmpty(t, createdStop.CreatedAt)
	})

	t.Run("GetByID", func(t *testing.T) {
		// First create a stop to get
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Name:        "Test Stop 2",
			Description: "Test Description 2",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		createdStop, err := db.Create(ctx, stop)
		require.NoError(t, err)

		gotStop, err := db.GetByID(ctx, createdStop.ID)
		require.NoError(t, err)
		assert.Equal(t, createdStop.ID, gotStop.ID)
		assert.Equal(t, stop.Name, gotStop.Name)
	})

	t.Run("ListByTrip", func(t *testing.T) {
		stops, err := db.ListByTrip(ctx, createdTrip.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, stops)
	})

	t.Run("Update", func(t *testing.T) {
		// Create a stop to update
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Name:        "Test Stop to Update",
			Description: "Test Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		createdStop, err := db.Create(ctx, stop)
		require.NoError(t, err)

		updatedStop := &domain.Stop{
			ID:          createdStop.ID,
			TripID:      createdTrip.ID,
			Name:        "Updated Stop Name",
			Description: "Updated Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		updated, err := db.Update(ctx, updatedStop)
		require.NoError(t, err)
		assert.Equal(t, updatedStop.Name, updated.Name)
		assert.Equal(t, updatedStop.Description, updated.Description)
		assert.NotEqual(t, createdStop.UpdatedAt, updated.UpdatedAt)
	})

	t.Run("Delete", func(t *testing.T) {
		// Create a stop to delete
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Name:        "Test Stop to Delete",
			Description: "Test Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6173,
			},
		}

		createdStop, err := db.Create(ctx, stop)
		require.NoError(t, err)

		err = db.Delete(ctx, createdStop.ID)
		require.NoError(t, err)

		// Verify it's deleted
		gotStop, err := db.GetByID(ctx, createdStop.ID)
		assert.Error(t, err)
		assert.Nil(t, gotStop)
	})
}