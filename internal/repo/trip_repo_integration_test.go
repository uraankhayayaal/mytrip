package repo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"mytrip/internal/domain"
)

func TestTripRepoIntegration(t *testing.T) {
	t.Parallel()

	// Create test container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(10 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	}()

	// Get database connection
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	connStr := "host=" + host + " port=" + port.Port() + " user=user password=password dbname=testdb sslmode=disable"

	// Create repository with connection
	repo, err := NewTripRepo(connStr)
	if err != nil {
		t.Fatalf("failed to create trip repo: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		trip := &domain.Trip{
			Title:       "Test Trip",
			Description: "Test Description",
			UserID:      "1",
		}

		error := repo.Create(ctx, trip)
		if error != nil {
			t.Fatalf("failed to create trip: %v", error)
		}

		if trip.ID == "" {
			t.Error("trip ID should not be empty")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		trip, error := repo.GetByID(ctx, "1")
		if error != nil {
			t.Fatalf("failed to get trip by ID: %v", error)
		}

		if trip.Title != "Test Trip" {
			t.Error("trip title does not match")
		}
	})

	t.Run("ListByUser", func(t *testing.T) {
		trips, error := repo.ListByUser(ctx, "1")
		if error != nil {
			t.Fatalf("failed to list trips by user: %v", error)
		}

		if len(trips) == 0 {
			t.Error("trips list should not be empty")
		}
	})

	t.Run("Update", func(t *testing.T) {
		trip := &domain.Trip{
			ID:          "1",
			Title:       "Updated Trip",
			Description: "Updated Description",
			UserID:      "1",
		}

		error := repo.Update(ctx, trip)
		if error != nil {
			t.Fatalf("failed to update trip: %v", error)
		}

		updatedTrip, error := repo.GetByID(ctx, "1")
		if error != nil {
			t.Fatalf("failed to get updated trip: %v", error)
		}

		if updatedTrip.Title != "Updated Trip" {
			t.Error("trip title was not updated")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		error := repo.Delete(ctx, "1")
		if error != nil {
			t.Fatalf("failed to delete trip: %v", error)
		}

		_, error = repo.GetByID(ctx, "1")
		if error == nil {
			t.Error("trip should be deleted")
		}
	})
}