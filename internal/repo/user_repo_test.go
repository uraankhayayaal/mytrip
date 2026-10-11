// Package repo contains repository implementations and tests
package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
)

// TestUserRepo tests the user repository implementation with testcontainers
func TestUserRepo(t *testing.T) {
	t.Parallel()

	// Start a test PostgreSQL container
	ctx := context.Background()
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1).WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, postgresContainer.Terminate(ctx))
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create repository
	userRepo, err := repo.NewUserRepo(connStr)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()

		user := &domain.User{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		error := userRepo.Create(ctx, user)
		assert.NoError(t, error)
		assert.NotEmpty(t, user.ID)
		
		// Verify user was created
		foundUser, err := userRepo.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Name, foundUser.Name)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		ctx := context.Background()

		// Create a user first
		user := &domain.User{
			Email:    "test2@example.com",
			Password: "password123",
			Name:     "Test User 2",
		}

		error := userRepo.Create(ctx, user)
		assert.NoError(t, error)

		// Get user by email
		foundUser, err := userRepo.GetByEmail(ctx, user.Email)
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Name, foundUser.Name)
	})

	t.Run("GetByID", func(t *testing.T) {
		ctx := context.Background()

		// Create a user first
		user := &domain.User{
			Email:    "test3@example.com",
			Password: "password123",
			Name:     "Test User 3",
		}

		error := userRepo.Create(ctx, user)
		assert.NoError(t, error)

		// Get user by ID
		foundUser, err := userRepo.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Name, foundUser.Name)
	})

	t.Run("CreateDuplicateEmail", func(t *testing.T) {
		ctx := context.Background()

		// Create first user
		user1 := &domain.User{
			Email:    "duplicate@example.com",
			Password: "password123",
			Name:     "Test User 1",
		}

		error := userRepo.Create(ctx, user1)
		assert.NoError(t, error)

		// Try to create user with same email
		user2 := &domain.User{
			Email:    "duplicate@example.com",
			Password: "password456",
			Name:     "Test User 2",
		}

		error = userRepo.Create(ctx, user2)
		assert.Error(t, error)
		// You might want to check for specific error type
	})
}