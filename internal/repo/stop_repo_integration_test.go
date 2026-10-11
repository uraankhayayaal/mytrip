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

func TestStopRepoIntegration(t *testing.T) {
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
	repo, err := NewStopRepo(connStr)
	if err != nil {
		t.Fatalf("failed to create stop repo: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		stop := &domain.Stop{
			TripID:      "1",
			Name:        "Test Stop",
			Description: "Test Description",
		}

		error := repo.Create(ctx, stop)
		if error != nil {
			t.Fatalf("failed to create stop: %v", error)
		}

		if stop.ID == "" {
			t.Error("stop ID should not be empty")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		stop, error := repo.GetByID(ctx, "1")
		if error != nil {
			t.Fatalf("failed to get stop by ID: %v", error)
		}

		if stop.Name != "Test Stop" {
			t.Error("stop name does not match")
		}
	})

	t.Run("ListByTrip", func(t *testing.T) {
		stops, error := repo.ListByTrip(ctx, "1")
		if error != nil {
			t.Fatalf("failed to list stops by trip: %v", error)
		}

		if len(stops) == 0 {
			t.Error("stops list should not be empty")
		}
	})

	t.Run("Update", func(t *testing.T) {
		stop := &domain.Stop{
			ID:          "1",
			TripID:      "1",
			Name:        "Updated Stop",
			Description: "Updated Description",
		}

		error := repo.Update(ctx, stop)
		if error != nil {
			t.Fatalf("failed to update stop: %v", error)
		}

		updatedStop, error := repo.GetByID(ctx, "1")
		if error != nil {
			t.Fatalf("failed to get updated stop: %v", error)
		}

		if updatedStop.Name != "Updated Stop" {
			t.Error("stop name was not updated")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		error := repo.Delete(ctx, "1")
		if error != nil {
			t.Fatalf("failed to delete stop: %v", error)
		}

		_, error = repo.GetByID(ctx, "1")
		if error == nil {
			t.Error("stop should be deleted")
		}
	})
}