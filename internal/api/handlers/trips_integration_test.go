//go:build integration
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

	// Test trip repository CRUD operations
	t.Run("CreateTrip - Success", func(t *testing.T) {
		// Create user repo first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "tripuser@example.com",
			Name:     "Trip User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create trip
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
		assert.Equal(t, createdUser.ID, createdTrip.UserID)
		assert.NotEmpty(t, createdTrip.CreatedAt)
	})

	t.Run("ListTrips - Success", func(t *testing.T) {
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "listuser@example.com",
			Name:     "List User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create some trips
		trip1 := &domain.Trip{
			Title:       "Trip 1",
			Description: "Description 1",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		trip2 := &domain.Trip{
			Title:       "Trip 2",
			Description: "Description 2",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		_, err = tripRepo.Create(ctx, trip1)
		require.NoError(t, err)
		_, err = tripRepo.Create(ctx, trip2)
		require.NoError(t, err)

		// List trips
		trips, err := tripRepo.ListByUser(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, trips)
		assert.Len(t, trips, 2)
	})

	t.Run("GetTrip - Success", func(t *testing.T) {
		// Create user repo first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "getuser@example.com",
			Name:     "Get User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create a trip
		trip := &domain.Trip{
			Title:       "Get Trip",
			Description: "Get Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Get trip by ID
		gotTrip, err := tripRepo.GetByID(ctx, createdTrip.ID)
		require.NoError(t, err)
		assert.Equal(t, createdTrip.ID, gotTrip.ID)
		assert.Equal(t, trip.Title, gotTrip.Title)
	})

	t.Run("UpdateTrip - Success", func(t *testing.T) {
		// Create user repo first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "updateuser@example.com",
			Name:     "Update User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create a trip to update
		trip := &domain.Trip{
			Title:       "Trip to Update",
			Description: "Original Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Update trip
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
		assert.NotEqual(t, createdTrip.UpdatedAt, updated.UpdatedAt)
	})

	t.Run("DeleteTrip - Success", func(t *testing.T) {
		// Create user repo first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "deleteuser@example.com",
			Name:     "Delete User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create a trip to delete
		trip := &domain.Trip{
			Title:       "Trip to Delete",
			Description: "Delete Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Delete trip
		err = tripRepo.Delete(ctx, createdTrip.ID)
		require.NoError(t, err)

		// Verify it's deleted
		gotTrip, err := tripRepo.GetByID(ctx, createdTrip.ID)
		assert.Error(t, err)
		assert.Nil(t, gotTrip)
	})

	t.Run("ListStops - Success", func(t *testing.T) {
		// Create user repo first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "stopuser@example.com",
			Name:     "Stop User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create a trip
		trip := &domain.Trip{
			Title:       "Trip with Stops",
			Description: "Trip Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Create stop repo
		stopRepo, err := repo.NewStopRepo(connectionString)
		require.NoError(t, err)

		// Create some stops
		stop1 := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "Stop 1",
			Description: "Description 1",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		stop2 := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "Stop 2",
			Description: "Description 2",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       2,
		}

		_, err = stopRepo.Create(ctx, stop1)
		require.NoError(t, err)
		_, err = stopRepo.Create(ctx, stop2)
		require.NoError(t, err)

		// List stops
		stops, err := stopRepo.ListByTrip(ctx, createdTrip.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, stops)
		assert.Len(t, stops, 2)
	})

	t.Run("CreateStop - Success", func(t *testing.T) {
		// Create user repo first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "stopcreateuser@example.com",
			Name:     "Stop Create User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create a trip
		trip := &domain.Trip{
			Title:       "Trip to Create Stop",
			Description: "Trip Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Create stop repo
		stopRepo, err := repo.NewStopRepo(connectionString)
		require.NoError(t, err)

		// Create a stop
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "New Stop",
			Description: "Stop Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		createdStop, err := stopRepo.Create(ctx, stop)
		require.NoError(t, err)
		assert.NotEmpty(t, createdStop.ID)
		assert.Equal(t, stop.Title, createdStop.Title)
		assert.Equal(t, createdTrip.ID, createdStop.TripID)
	})

	t.Run("UpdateStop - Success", func(t *testing.T) {
		// Create user repo first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "stopupdateuser@example.com",
			Name:     "Stop Update User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create a trip
		trip := &domain.Trip{
			Title:       "Trip to Update Stop",
			Description: "Trip Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Create stop repo
		stopRepo, err := repo.NewStopRepo(connectionString)
		require.NoError(t, err)

		// Create a stop to update
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "Stop to Update",
			Description: "Original Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		createdStop, err := stopRepo.Create(ctx, stop)
		require.NoError(t, err)

		// Update stop
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
		assert.NotEqual(t, createdStop.UpdatedAt, updated.UpdatedAt)
	})

	t.Run("DeleteStop - Success", func(t *testing.T) {
		// Create user repo first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "stopdeleteuser@example.com",
			Name:     "Stop Delete User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create trip repo
		tripRepo, err := repo.NewTripRepo(connectionString)
		require.NoError(t, err)

		// Create a trip
		trip := &domain.Trip{
			Title:       "Trip to Delete Stop",
			Description: "Trip Description",
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		require.NoError(t, err)

		// Create stop repo
		stopRepo, err := repo.NewStopRepo(connectionString)
		require.NoError(t, err)

		// Create a stop to delete
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Title:       "Stop to Delete",
			Description: "Stop Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
		}

		createdStop, err := stopRepo.Create(ctx, stop)
		require.NoError(t, err)

		// Delete stop
		err = stopRepo.Delete(ctx, createdStop.ID)
		require.NoError(t, err)

		// Verify it's deleted
		gotStop, err := stopRepo.GetByID(ctx, createdStop.ID)
		assert.Error(t, err)
		assert.Nil(t, gotStop)
	})
}
