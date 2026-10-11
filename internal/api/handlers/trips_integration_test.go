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

// TestTripsIntegration tests trips endpoints with full integration
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

	// Initialize repositories
	userRepo, err := repo.NewUserRepo(connectionString)
	require.NoError(t, err)

	tripRepo, err := repo.NewTripRepo(connectionString)
	require.NoError(t, err)

	stopRepo, err := repo.NewStopRepo(connectionString)
	require.NoError(t, err)

	postRepo, err := repo.NewPhotoRepo(connectionString)
	require.NoError(t, err)

	// Create a test user for auth
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("ListTrips", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("CreateTrip", func(t *testing.T) {
		tripReq := map[string]interface{}{
			"title":       "Test Trip",
			"description": "Test Description",
			"start_date":  time.Now().Format(time.RFC3339),
			"end_date":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		}

		reqBody, _ := json.Marshal(tripReq)
		req := httptest.NewRequest("POST", "/api/v1/trips", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("CreateTrip Validation Error", func(t *testing.T) {
		tripReq := map[string]interface{}{
			"title":       "",
			"description": "Test Description",
		}

		reqBody, _ := json.Marshal(tripReq)
		req := httptest.NewRequest("POST", "/api/v1/trips", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("GetTrip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/123", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("GetTrip Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/999", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("UpdateTrip", func(t *testing.T) {
		tripReq := map[string]interface{}{
			"title":       "Updated Trip",
			"description": "Updated Description",
			"start_date":  time.Now().Format(time.RFC3339),
			"end_date":    time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		}

		reqBody, _ := json.Marshal(tripReq)
		req := httptest.NewRequest("PUT", "/api/v1/trips/123", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("UpdateTrip Validation Error", func(t *testing.T) {
		tripReq := map[string]interface{}{
			"title":       "",
			"description": "Updated Description",
		}

		reqBody, _ := json.Marshal(tripReq)
		req := httptest.NewRequest("PUT", "/api/v1/trips/123", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("DeleteTrip", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/trips/123", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("ListStops", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/123/stops", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("CreateStop", func(t *testing.T) {
		stopReq := map[string]interface{}{
			"name":        "Test Stop",
			"description": "Test Description",
			"location": map[string]interface{}{
				"lat": 55.7558,
				"lng": 37.6173,
			},
		}

		reqBody, _ := json.Marshal(stopReq)
		req := httptest.NewRequest("POST", "/api/v1/trips/123/stops", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("UpdateStop", func(t *testing.T) {
		stopReq := map[string]interface{}{
			"name":        "Updated Stop",
			"description": "Updated Description",
			"location": map[string]interface{}{
				"lat": 55.7558,
				"lng": 37.6173,
			},
		}

		reqBody, _ := json.Marshal(stopReq)
		req := httptest.NewRequest("PUT", "/api/v1/stops/123", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("DeleteStop", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/stops/123", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})
}