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

func TestTripRepoIntegration(t *testing.T) {
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

	// Test Create
	t.Run("Create", func(t *testing.T) {
		trip := &domain.Trip{
			Title:       "Test Trip",
			Description: "Test Description",
			UserID:      createdUser.ID,
		}

		createdTrip, err := tripRepo.Create(ctx, trip)
		if err != nil {
			t.Fatalf("failed to create trip: %v", err)
		}

		if createdTrip.ID == "" {
			t.Error("created trip ID should not be empty")
		}

		// Test GetByID
		retrievedTrip, err := tripRepo.GetByID(ctx, createdTrip.ID)
		if err != nil {
			t.Fatalf("failed to get trip by ID: %v", err)
		}

		if retrievedTrip.ID != createdTrip.ID {
			t.Error("retrieved trip ID should match created trip ID")
		}

		// Test ListByUser
		trips, err := tripRepo.ListByUser(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("failed to list trips by user: %v", err)
		}

		if len(trips) == 0 {
			t.Error("should have at least one trip")
		}

		// Test Update
		updatedTrip := &domain.Trip{
			ID:          createdTrip.ID,
			Title:       "Updated Trip",
			Description: "Updated Description",
			UserID:      createdUser.ID,
		}

		_, err = tripRepo.Update(ctx, updatedTrip)
		if err != nil {
			t.Fatalf("failed to update trip: %v", err)
		}

		// Test Delete
		_, err = tripRepo.Delete(ctx, createdTrip.ID)
		if err != nil {
			t.Fatalf("failed to delete trip: %v", err)
		}
	})
}