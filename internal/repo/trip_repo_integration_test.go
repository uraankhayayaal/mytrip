package repo

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"mytrip/internal/domain"
)

func TestTripRepoIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	// Clean up the container after the test is finished
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	})

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Connect to database
	db, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create trip repo
	repo := NewTripRepo(db)

	t.Run("CreateTrip", func(t *testing.T) {
		trip := &domain.Trip{
			ID:          uuid.New().String(),
			UserID:      "1",
			Title:       "Test Trip",
			Description: "Test Description",
			StartDate:   time.Now(),
			EndDate:     time.Now().Add(24 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		newTrip, err := repo.Create(ctx, trip)
		if err != nil {
			t.Fatalf("failed to create trip: %v", err)
		}

		if newTrip.ID != trip.ID {
			t.Errorf("expected trip ID %s, got %s", trip.ID, newTrip.ID)
		}
	})

	t.Run("GetTripByID", func(t *testing.T) {
		trip, err := repo.GetByID(ctx, "1")
		if err != nil {
			t.Fatalf("failed to get trip by ID: %v", err)
		}

		if trip.ID != "1" {
			t.Errorf("expected trip ID 1, got %s", trip.ID)
		}
	})

	t.Run("ListTripsByUser", func(t *testing.T) {
		trips, err := repo.ListByUser(ctx, "1")
		if err != nil {
			t.Fatalf("failed to list trips by user: %v", err)
		}

		if len(trips) == 0 {
			t.Error("expected at least one trip")
		}
	})

	t.Run("UpdateTrip", func(t *testing.T) {
		updatedTrip := &domain.Trip{
			ID:          "1",
			UserID:      "1",
			Title:       "Updated Test Trip",
			Description: "Updated Description",
			StartDate:   time.Now(),
			EndDate:     time.Now().Add(48 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err := repo.Update(ctx, updatedTrip)
		if err != nil {
			t.Fatalf("failed to update trip: %v", err)
		}

		// Verify update worked
		trip, err := repo.GetByID(ctx, "1")
		if err != nil {
			t.Fatalf("failed to get updated trip: %v", err)
		}

		if trip.Title != "Updated Test Trip" {
			t.Errorf("expected updated title Updated Test Trip, got %s", trip.Title)
		}
	})

	t.Run("DeleteTrip", func(t *testing.T) {
		err := repo.Delete(ctx, "1")
		if err != nil {
			t.Fatalf("failed to delete trip: %v", err)
		}

		// Verify deletion worked
		_, err = repo.GetByID(ctx, "1")
		if err == nil {
			t.Error("expected trip to be deleted")
		}
	})
}
