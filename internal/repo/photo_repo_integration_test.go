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

func TestPhotoRepoIntegration(t *testing.T) {
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
	repo, err := NewPhotoRepo(connStr)
	if err != nil {
		t.Fatalf("failed to create photo repo: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		photo := &domain.Photo{
			TripID: "1",
			URL:    "https://example.com/photo.jpg",
		}

		error := repo.Create(ctx, photo)
		if error != nil {
			t.Fatalf("failed to create photo: %v", error)
		}

		if photo.ID == "" {
			t.Error("photo ID should not be empty")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		photo, error := repo.GetByID(ctx, "1")
		if error != nil {
			t.Fatalf("failed to get photo by ID: %v", error)
		}

		if photo.URL != "https://example.com/photo.jpg" {
			t.Error("photo URL does not match")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		error := repo.Delete(ctx, "1")
		if error != nil {
			t.Fatalf("failed to delete photo: %v", error)
		}

		_, error = repo.GetByID(ctx, "1")
		if error == nil {
			t.Error("photo should be deleted")
		}
	})
}