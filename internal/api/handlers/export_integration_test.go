//go:build integration
// +build integration

package handlers

import (
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

func TestExportIntegration(t *testing.T) {
	// Integration tests for export endpoints
	// GET /api/v1/trips/{id}/export?format=pdf|csv → 200 (стриминг)

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

	// Create test trip
	trip := &domain.Trip{
		Title:       "Test Trip for Export",
		Description: "Test Description",
		StartAt:     time.Now(),
		EndAt:       time.Now().Add(24 * time.Hour),
		UserID:      createdUser.ID,
	}
	createdTrip, err := tripRepo.Create(ctx, trip)
	require.NoError(t, err)

	// Create some stops for export
	stop := &domain.Stop{
		TripID:      createdTrip.ID,
		Title:       "Test Stop",
		Description: "Test Description",
		Lat:         55.7558,
		Lng:         37.6173,
		VisitDate:   time.Now(),
		Order:       1,
	}
	_, err = stopRepo.Create(ctx, stop)
	require.NoError(t, err)

	t.Run("ExportPDF - Success", func(t *testing.T) {
		// In a real implementation this would test PDF export
		// This is a placeholder test
		t.Skip("PDF export not implemented in this test")
	})

	t.Run("ExportCSV - Success", func(t *testing.T) {
		// In a real implementation this would test CSV export
		// This is a placeholder test
		t.Skip("CSV export not implemented in this test")
	})

	t.Run("ExportInvalidFormat - Error", func(t *testing.T) {
		// Test export with invalid format
		// In a real implementation this would return 400
		t.Skip("Format validation not implemented in this test")
	})
}
