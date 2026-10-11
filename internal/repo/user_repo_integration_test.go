package repo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"mytrip/internal/domain"
)

func TestUserRepoIntegration(t *testing.T) {
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

	// Create user repo
	repo := NewUserRepo(db)

	t.Run("CreateUser", func(t *testing.T) {
		user := &domain.User{
			ID:        uuid.New().String(),
			Email:     "test@example.com",
			Name:      "Test User",
			CreatedAt: time.Now(),
		}

		newUser, err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		if newUser.ID != user.ID {
			t.Errorf("expected user ID %s, got %s", user.ID, newUser.ID)
		}

		if newUser.Email != user.Email {
			t.Errorf("expected user email %s, got %s", user.Email, newUser.Email)
		}
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "test@example.com")
		if err != nil {
			t.Fatalf("failed to get user by email: %v", err)
		}

		if user.Email != "test@example.com" {
			t.Errorf("expected user email test@example.com, got %s", user.Email)
		}
	})

	t.Run("GetUserByID", func(t *testing.T) {
		user, err := repo.GetByID(ctx, "1")
		if err != nil {
			t.Fatalf("failed to get user by ID: %v", err)
		}

		if user.ID != "1" {
			t.Errorf("expected user ID 1, got %s", user.ID)
		}
	})
}

func runMigrations(db *pgxpool.Pool) error {
	// In a real implementation, you would run actual migrations here
	return nil
}
