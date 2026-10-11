//go:build integration

package repo

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

func TestUserRepoIntegration(t *testing.T) {
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

	// Test Create
	t.Run("Create", func(t *testing.T) {
		user := &domain.User{
			Email:     "test@example.com",
			Name:      "Test User",
			Password:  "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		if createdUser.ID == "" {
			t.Error("created user ID should not be empty")
		}

		// Test GetByEmail
		retrievedUser, err := userRepo.GetByEmail(ctx, "test@example.com")
		if err != nil {
			t.Fatalf("failed to get user by email: %v", err)
		}

		if retrievedUser.ID != createdUser.ID {
			t.Error("retrieved user ID should match created user ID")
		}

		// Test GetByID
		retrievedUser, err = userRepo.GetByID(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("failed to get user by ID: %v", err)
		}

		if retrievedUser.ID != createdUser.ID {
			t.Error("retrieved user ID should match created user ID")
		}
	})

	// Test conflict on duplicate email
	t.Run("CreateWithDuplicateEmail", func(t *testing.T) {
		user := &domain.User{
			Email:     "test@example.com",
			Name:      "Test User 2",
			Password:  "password123",
		}

		_, err := userRepo.Create(ctx, user)
		if err == nil {
			t.Error("should have returned error for duplicate email")
		}
	})
}

func runMigrations(pool *pgxpool.Pool) error {
	// In a real implementation, you would run migrations here
	// For now, we'll assume migrations are already set up
	return nil
}