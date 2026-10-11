package repo

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"mytrip/internal/domain"
)

func TestPhotoRepoIntegration(t *testing.T) {
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
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Clean up the container after the test is finished
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	})

	// Create MinIO container
	minioContainer, err := minio.RunContainer(ctx,
		testcontainers.WithImage("minio/minio:latest"),
		minio.WithUsername("minioadmin"),
		minio.WithPassword("minioadmin"),
	)
	if err != nil {
		t.Fatalf("failed to start minio container: %v", err)
	}

	// Clean up the container after the test is finished
	t.Cleanup(func() {
		if err := minioContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate minio container: %v", err)
		}
	})

	// Get connection strings
	postgresConnStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	minioConnStr, err := minioContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get minio connection string: %v", err)
	}

	// Connect to database
	db, err := pgxpool.New(ctx, postgresConnStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create photo repo
	repo := NewPhotoRepo(db)

	t.Run("CreatePhoto", func(t *testing.T) {
		photo := &domain.Photo{
			ID:        uuid.New().String(),
			TripID:    "1",
			URL:       "http://example.com/photo.jpg",
			CreatedAt: time.Now(),
		}

		newPhoto, err := repo.Create(ctx, photo)
		if err != nil {
			t.Fatalf("failed to create photo: %v", err)
		}

		if newPhoto.ID != photo.ID {
			t.Errorf("expected photo ID %s, got %s", photo.ID, newPhoto.ID)
		}
	})

	t.Run("GetPhotoByID", func(t *testing.T) {
		photo, err := repo.GetByID(ctx, "1")
		if err != nil {
			t.Fatalf("failed to get photo by ID: %v", err)
		}

		if photo.ID != "1" {
			t.Errorf("expected photo ID 1, got %s", photo.ID)
		}
	})

	t.Run("DeletePhoto", func(t *testing.T) {
		err := repo.Delete(ctx, "1")
		if err != nil {
			t.Fatalf("failed to delete photo: %v", err)
		}

		// Verify deletion worked
		_, err = repo.GetByID(ctx, "1")
		if err == nil {
			t.Error("expected photo to be deleted")
		}
	})
}
