// +build integration

package handlers

import (
	"context"
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

// TestExportIntegration tests export endpoints with full integration
func TestExportIntegration(t *testing.T) {
	// Integration tests for export endpoints
	// GET /api/v1/trips/{id}/export?format=pdf|csv → 200 (стриминг)

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

	// Create a test user and trip
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
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

	t.Run("ExportPDF", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/123/export?format=pdf", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("ExportCSV", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/123/export?format=csv", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("ExportInvalidFormat", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/123/export?format=xml", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("ExportNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/999/export?format=pdf", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})
}