// Package repo contains repository implementations and tests
package repo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"mytrip/internal/domain"
)

// TestUserRepo tests the user repository implementation
func TestUserRepo(t *testing.T) {
	t.Parallel()

	// Create a test database container
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1).WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}

	// Get connection string
	connectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create repository
	db, err := NewUserRepo(connectionString)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		user := &domain.User{
			Email:     "test@example.com",
			Name:      "Test User",
			Password:  "password123",
		}

		createdUser, err := db.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, createdUser.ID)
		assert.Equal(t, user.Email, createdUser.Email)
		assert.Equal(t, user.Name, createdUser.Name)
		assert.NotEmpty(t, createdUser.CreatedAt)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		user, err := db.GetByEmail(ctx, "test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
	})

	t.Run("GetByID", func(t *testing.T) {
		// First create a user to get
		user := &domain.User{
			Email:     "test2@example.com",
			Name:      "Test User 2",
			Password:  "password123",
		}

		createdUser, err := db.Create(ctx, user)
		require.NoError(t, err)

		nonExistentUser, err := db.GetByID(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Nil(t, nonExistentUser)

		gotUser, err := db.GetByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, gotUser.ID)
		assert.Equal(t, user.Email, gotUser.Email)
	})

	t.Run("GetByEmail conflict", func(t *testing.T) {
		// Try to create user with existing email
		user := &domain.User{
			Email:     "test@example.com",
			Name:      "Another User",
			Password:  "password123",
		}

		_, err := db.Create(ctx, user)
		assert.Error(t, err)
	})
}