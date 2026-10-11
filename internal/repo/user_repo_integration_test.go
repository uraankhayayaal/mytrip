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

func TestUserRepoIntegration(t *testing.T) {
	t.Parallel()

	// Create test container
n	ctx := context.Background()
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
	repo, err := NewUserRepo(connStr)
	if err != nil {
		t.Fatalf("failed to create user repo: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		user := &domain.User{
			Email:     "test@example.com",
			Name:      "Test User",
			Password:  "password123",
		}

		error := repo.Create(ctx, user)
		if error != nil {
			t.Fatalf("failed to create user: %v", error)
		}

		if user.ID == "" {
			t.Error("user ID should not be empty")
		}
	})

	t.Run("GetByEmail", func(t *testing.T) {
		user, error := repo.GetByEmail(ctx, "test@example.com")
		if error != nil {
			t.Fatalf("failed to get user by email: %v", error)
		}

		if user.Email != "test@example.com" {
			t.Error("user email does not match")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		user, error := repo.GetByID(ctx, "1")
		if error != nil {
			t.Fatalf("failed to get user by ID: %v", error)
		}

		if user.Email != "test@example.com" {
			t.Error("user email does not match")
		}
	})
}