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

// TestTripRepo tests the trip repository implementation
func TestTripRepo(t *testing.T) {
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

	// Create repository
	db, err := NewTripRepo(connectionString)
	require.NoError(t, err)

	// Create a test user first
	userRepo, err := NewUserRepo(connectionString)
	require.NoError(t, err)
	
	user := &domain.User{
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		trip := &domain.Trip{
			Title:       "Test Trip",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := db.Create(ctx, trip)
		require.NoError(t, err)
		assert.NotEmpty(t, createdTrip.ID)
		assert.Equal(t, trip.Title, createdTrip.Title)
		assert.Equal(t, trip.Description, createdTrip.Description)
		assert.Equal(t, createdUser.ID, createdTrip.UserID)
		assert.NotEmpty(t, createdTrip.CreatedAt)
	})

	t.Run("GetByID", func(t *testing.T) {
		// First create a trip to get
		trip := &domain.Trip{
			Title:       "Test Trip 2",
			Description: "Test Description 2",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := db.Create(ctx, trip)
		require.NoError(t, err)

		gotTrip, err := db.GetByID(ctx, createdTrip.ID)
		require.NoError(t, err)
		assert.Equal(t, createdTrip.ID, gotTrip.ID)
		assert.Equal(t, trip.Title, gotTrip.Title)
	})

	t.Run("ListByUser", func(t *testing.T) {
		trips, err := db.ListByUser(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, trips)
	})

	t.Run("Update", func(t *testing.T) {
		// Create a trip to update
		trip := &domain.Trip{
			Title:       "Test Trip to Update",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := db.Create(ctx, trip)
		require.NoError(t, err)

		updatedTrip := &domain.Trip{
			ID:          createdTrip.ID,
			Title:       "Updated Title",
			Description: "Updated Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(48 * time.Hour),
			UserID:      createdUser.ID,
		}

		updated, err := db.Update(ctx, updatedTrip)
		require.NoError(t, err)
		assert.Equal(t, updatedTrip.Title, updated.Title)
		assert.Equal(t, updatedTrip.Description, updated.Description)
		assert.NotEqual(t, createdTrip.UpdatedAt, updated.UpdatedAt)
	})

	t.Run("Delete", func(t *testing.T) {
		// Create a trip to delete
		trip := &domain.Trip{
			Title:       "Test Trip to Delete",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := db.Create(ctx, trip)
		require.NoError(t, err)

		err = db.Delete(ctx, createdTrip.ID)
		require.NoError(t, err)

		// Verify it's deleted
		gotTrip, err := db.GetByID(ctx, createdTrip.ID)
		assert.Error(t, err)
		assert.Nil(t, gotTrip)
	})
}