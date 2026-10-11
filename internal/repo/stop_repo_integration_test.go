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

func TestStopRepoIntegration(t *testing.T) {
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

	// Create stop repo
	repo := NewStopRepo(db)

	t.Run("CreateStop", func(t *testing.T) {
		stop := &domain.Stop{
			ID:          uuid.New().String(),
			TripID:      "1",
			Title:       "Test Stop",
			Description: "Test Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
			CreatedAt:   time.Now(),
		}

		newStop, err := repo.Create(ctx, stop)
		if err != nil {
			t.Fatalf("failed to create stop: %v", err)
		}

		if newStop.ID != stop.ID {
			t.Errorf("expected stop ID %s, got %s", stop.ID, newStop.ID)
		}
	})

	t.Run("GetStopByID", func(t *testing.T) {
		stop, err := repo.GetByID(ctx, "1")
		if err != nil {
			t.Fatalf("failed to get stop by ID: %v", err)
		}

		if stop.ID != "1" {
			t.Errorf("expected stop ID 1, got %s", stop.ID)
		}
	})

	t.Run("ListStopsByTrip", func(t *testing.T) {
		stops, err := repo.ListByTrip(ctx, "1")
		if err != nil {
			t.Fatalf("failed to list stops by trip: %v", err)
		}

		if len(stops) == 0 {
			t.Error("expected at least one stop")
		}
	})

	t.Run("UpdateStop", func(t *testing.T) {
		updatedStop := &domain.Stop{
			ID:          "1",
			TripID:      "1",
			Title:       "Updated Test Stop",
			Description: "Updated Description",
			Lat:         55.7558,
			Lng:         37.6173,
			VisitDate:   time.Now(),
			Order:       1,
			CreatedAt:   time.Now(),
		}

		_, err := repo.Update(ctx, updatedStop)
		if err != nil {
			t.Fatalf("failed to update stop: %v", err)
		}

		// Verify update worked
		stop, err := repo.GetByID(ctx, "1")
		if err != nil {
			t.Fatalf("failed to get updated stop: %v", err)
		}

		if stop.Title != "Updated Test Stop" {
			t.Errorf("expected updated title Updated Test Stop, got %s", stop.Title)
		}
	})

	t.Run("DeleteStop", func(t *testing.T) {
		err := repo.Delete(ctx, "1")
		if err != nil {
			t.Fatalf("failed to delete stop: %v", err)
		}

		// Verify deletion worked
		_, err = repo.GetByID(ctx, "1")
		if err == nil {
			t.Error("expected stop to be deleted")
		}
	})
}
