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

func TestPhotoRepoIntegration(t *testing.T) {
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

	// Create photo repository
	photoRepo := repo.NewPhotoRepo(pool)

	// Test Create
	t.Run("Create", func(t *testing.T) {
		photo := &domain.Photo{
			TripID: createdTrip.ID,
			URL:    "http://example.com/photo.jpg",
		}

		createdPhoto, err := photoRepo.Create(ctx, photo)
		if err != nil {
			t.Fatalf("failed to create photo: %v", err)
		}

		if createdPhoto.ID == "" {
			t.Error("created photo ID should not be empty")
		}

		// Test GetByID
		retrievedPhoto, err := photoRepo.GetByID(ctx, createdPhoto.ID)
		if err != nil {
			t.Fatalf("failed to get photo by ID: %v", err)
		}

		if retrievedPhoto.ID != createdPhoto.ID {
			t.Error("retrieved photo ID should match created photo ID")
		}

		// Test Delete
		_, err = photoRepo.Delete(ctx, createdPhoto.ID)
		if err != nil {
			t.Fatalf("failed to delete photo: %v", err)
		}
	})
}