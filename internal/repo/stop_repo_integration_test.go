//go:build integration

package repo

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

func TestStopRepoIntegration(t *testing.T) {
	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Clean up the container after the test is finished

defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Initialize database connection
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Run migrations
	err = runMigrations(pool)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create user repository
	userRepo := repo.NewUserRepo(pool)

	// Create a user first
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create trip repository
	tripRepo := repo.NewTripRepo(pool)

	// Create a trip
	trip := &domain.Trip{
		Title:       "Test Trip",
		Description: "Test Description",
		UserID:      createdUser.ID,
		}
	createdTrip, err := tripRepo.Create(ctx, trip)
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}

	// Create stop repository
	stopRepo := repo.NewStopRepo(pool)

	// Test Create
	t.Run("Create", func(t *testing.T) {
		stop := &domain.Stop{
			TripID:      createdTrip.ID,
			Name:        "Test Stop",
			Description: "Test Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6176,
			},
		}

		createdStop, err := stopRepo.Create(ctx, stop)
		if err != nil {
			t.Fatalf("failed to create stop: %v", err)
		}

		if createdStop.ID == "" {
			t.Error("created stop ID should not be empty")
		}

		// Test GetByID
		retrievedStop, err := stopRepo.GetByID(ctx, createdStop.ID)
		if err != nil {
			t.Fatalf("failed to get stop by ID: %v", err)
		}

		if retrievedStop.ID != createdStop.ID {
			t.Error("retrieved stop ID should match created stop ID")
		}

		// Test ListByTrip
		stops, err := stopRepo.ListByTrip(ctx, createdTrip.ID)
		if err != nil {
			t.Fatalf("failed to list stops by trip: %v", err)
		}

		if len(stops) == 0 {
			t.Error("should have at least one stop")
		}

		// Test Update
		updatedStop := &domain.Stop{
			ID:          createdStop.ID,
			TripID:      createdTrip.ID,
			Name:        "Updated Stop",
			Description: "Updated Description",
			Location: &domain.Location{
				Lat: 55.7558,
				Lng: 37.6176,
			},
		}

		_, err = stopRepo.Update(ctx, updatedStop)
		if err != nil {
			t.Fatalf("failed to update stop: %v", err)
		}

		// Test Delete
		_, err = stopRepo.Delete(ctx, createdStop.ID)
		if err != nil {
			t.Fatalf("failed to delete stop: %v", err)
		}
	})
}