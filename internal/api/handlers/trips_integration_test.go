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
	"github.com/testcontainers/testcontainers-go/wait"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

func TestTripsIntegration(t *testing.T) {
	// Integration tests for trips endpoints
	// GET /api/v1/trips → 200 [trip]
	// POST /api/v1/trips → 201 trip
	// GET /api/v1/trips/{id} → 200 trip
	// PUT /api/v1/trips/{id} → 200 trip
	// DELETE /api/v1/trips/{id} → 204
	// GET /api/v1/trips/{id}/stops → 200 [stop]
	// POST /api/v1/trips/{id}/stops → 201 stop
	// PUT /api/v1/stops/{stopId} → 200 stop
	// DELETE /api/v1/stops/{stopId} → 204

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

	// Get connection string
	connectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create repositories
	userRepo, err := repo.NewUserRepo(connectionString)
	require.NoError(t, err)
	tripRepo, err := repo.NewTripRepo(connectionString)
	require.NoError(t, err)
	stopRepo, err := repo.NewStopRepo(connectionString)
	require.NoError(t, err)

	// Create test user
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("ListTrips - Success", func(t *testing.T) {
		trips, err := tripRepo.ListByUser(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.NotNil(t, trips)
	})

	t.Run("CreateTrip - Success", func(t *testing.T) {
		trip := &domain.Trip{
			Title:       "Test Trip",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)
		assert.NotEmpty(t, createdTrip.ID)
		assert.Equal(t, trip.Title, createdTrip.Title)
		assert.Equal(t, trip.Description, createdTrip.Description)
	})

	t.Run("CreateTrip - Validation Error", func(t *testing.T) {
		trip := &domain.Trip{
			Title:       "",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		_, err := tripRepo.Create(ctx, trip)
		assert.Error(t, err)
	})

	t.Run("GetTrip - Success", func(t *testing.T) {
		// Create a trip first
		trip := &domain.Trip{
			Title:       "Test Trip 2",
			Description: "Test Description 2",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		gotTrip, err := tripRepo.GetByID(ctx, createdTrip.ID)
		require.NoError(t, err)
		assert.Equal(t, createdTrip.ID, gotTrip.ID)
		assert.Equal(t, trip.Title, gotTrip.Title)
	})

	t.Run("GetTrip - Not Found", func(t *testing.T) {
		_, err := tripRepo.GetByID(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("UpdateTrip - Success", func(t *testing.T) {
		// Create a trip to update
		trip := &domain.Trip{
			Title:       "Test Trip to Update",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		updatedTrip := &domain.Trip{
			ID:          createdTrip.ID,
			Title:       "Updated Title",
			Description: "Updated Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(48 * time.Hour),
			UserID:      createdUser.ID,
		}

		updated, err := tripRepo.Update(ctx, updatedTrip)
		require.NoError(t, err)
		assert.Equal(t, updatedTrip.Title, updated.Title)
		assert.Equal(t, updatedTrip.Description, updated.Description)
	})

	t.Run("UpdateTrip - Validation Error", func(t *testing.T) {
		trip := &domain.Trip{
			Title:       "",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		_, err := tripRepo.Update(ctx, trip)
		assert.Error(t, err)
	})

	t.Run("UpdateTrip - Not Found", func(t *testing.T) {
		trip := &domain.Trip{
			ID:          "non-existent-id",
			Title:       "Updated Title",
			Description: "Updated Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(48 * time.Hour),
			UserID:      createdUser.ID,
		}

		_, err := tripRepo.Update(ctx, trip)
		assert.Error(t, err)
	})

	t.Run("DeleteTrip - Success", func(t *testing.T) {
		// Create a trip to delete
		trip := &domain.Trip{
			Title:       "Test Trip to Delete",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		err = tripRepo.Delete(ctx, createdTrip.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = tripRepo.GetByID(ctx, createdTrip.ID)
		assert.Error(t, err)
	})

	t.Run("DeleteTrip - Not Found", func(t *testing.T) {
		err := tripRepo.Delete(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("ListStops - Success", func(t *testing.T) {
		// Create a trip first
		trip := &domain.Trip{
			Title:       "Test Trip for Stops",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		stops, err := stopRepo.ListByTrip(ctx, createdTrip.ID)
		require.NoError(t, err)
		assert.NotNil(t, stops)
	})

	t.Run("CreateStop - Success", func(t *testing.T) {
		// Create a trip first
		trip := &domain.Trip{
			Title:       "Test Trip for Stops",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "Test Stop",
			Description: "Test Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		createdStop, err := stopRepo.Create(ctx, stop)
		require.NoError(t, err)
		assert.NotEmpty(t, createdStop.ID)
		assert.Equal(t, stop.Title, createdStop.Title)
	})

	t.Run("CreateStop - Validation Error", func(t *testing.T) {
		// Create a trip first
		trip := &domain.Trip{
			Title:       "Test Trip for Stops",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "",
			Description: "Test Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		_, err := stopRepo.Create(ctx, stop)
		assert.Error(t, err)
	})

	t.Run("CreateStop - Not Found", func(t *testing.T) {
		stop := &domain.Stop{
			TripID:      "non-existent-id",
			Title:       "Test Stop",
			Description: "Test Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		_, err := stopRepo.Create(ctx, stop)
		assert.Error(t, err)
	})

	t.Run("UpdateStop - Success", func(t *testing.T) {
		// Create a trip first
		trip := &domain.Trip{
			Title:       "Test Trip for Stops",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Create a stop to update
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "Test Stop to Update",
			Description: "Test Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		createdStop, err := stopRepo.Create(ctx, stop)
		require.NoError(t, err)

		updatedStop := &domain.Stop{
			ID:          createdStop.ID,
			TripID:      createdTrip.ID,
			Title:       "Updated Stop Name",
			Description: "Updated Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		updated, err := stopRepo.Update(ctx, updatedStop)
		require.NoError(t, err)
		assert.Equal(t, updatedStop.Title, updated.Title)
		assert.Equal(t, updatedStop.Description, updated.Description)
	})

	t.Run("UpdateStop - Validation Error", func(t *testing.T) {
		// Create a trip first
		trip := &domain.Trip{
			Title:       "Test Trip for Stops",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Create a stop to update
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "Test Stop to Update",
			Description: "Test Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		createdStop, err := stopRepo.Create(ctx, stop)
		require.NoError(t, err)

		updatedStop := &domain.Stop{
			ID:          createdStop.ID,
			TripID:      createdTrip.ID,
			Title:       "",
			Description: "Updated Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		_, err := stopRepo.Update(ctx, updatedStop)
		assert.Error(t, err)
	})

	t.Run("UpdateStop - Not Found", func(t *testing.T) {
		updatedStop := &domain.Stop{
			ID:          "non-existent-id",
			TripID:      "non-existent-id",
			Title:       "Updated Stop Name",
			Description: "Updated Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		_, err := stopRepo.Update(ctx, updatedStop)
		assert.Error(t, err)
	})

	t.Run("DeleteStop - Success", func(t *testing.T) {
		// Create a trip first
		trip := &domain.Trip{
			Title:       "Test Trip for Stops",
			Description: "Test Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Create a stop to delete
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "Test Stop to Delete",
			Description: "Test Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		createdStop, err := stopRepo.Create(ctx, stop)
		require.NoError(t, err)

		err = stopRepo.Delete(ctx, createdStop.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = stopRepo.GetByID(ctx, createdStop.ID)
		assert.Error(t, err)
	})

	t.Run("DeleteStop - Not Found", func(t *testing.T) {
		err := stopRepo.Delete(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}