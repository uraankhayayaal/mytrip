//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestAuthIntegration(t *testing.T) {
	// Integration tests for auth endpoints
	// POST /api/v1/auth/register → 201 {user}
	// POST /api/v1/auth/login → 200 {access_token, user} + Set-Cookie rt
	// POST /api/v1/auth/refresh → 200 {access_token}
	// POST /api/v1/auth/logout → 204
	// GET /api/v1/users/me → 200 {user}

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
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// Get connection string
	connectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Create test server with actual handlers and repository
	// For now we'll simulate the structure
	
	t.Run("Register - Valid", func(t *testing.T) {
		// Create user repository
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Test registration
		user := &domain.User{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, createdUser.ID)
		assert.Equal(t, user.Email, createdUser.Email)
		assert.Equal(t, user.Name, createdUser.Name)
		assert.NotEmpty(t, createdUser.CreatedAt)
	})

	t.Run("Register - Validation Error", func(t *testing.T) {
		// Test with invalid data
		user := &domain.User{
			Email:    "", // Invalid email
			Name:     "Test User",
			Password: "password123",
		}

		// This would be tested through the API endpoint
		// For now we test at repository level
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)
		_, err = userRepo.Create(ctx, user)
		assert.Error(t, err)
	})

	t.Run("Register - Conflict Error", func(t *testing.T) {
		// Try to register user with existing email
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		user := &domain.User{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		// Create first user
		_, err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Try to create user with same email
		_, err = userRepo.Create(ctx, user)
		assert.Error(t, err)
	})

	t.Run("Login - Success", func(t *testing.T) {
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "login@example.com",
			Name:     "Login User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Test login would be via API handler but we verify at repo level
		gotUser, err := userRepo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, gotUser.ID)
	})

	t.Run("GetUser - Success", func(t *testing.T) {
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)

		// Create a test user
		user := &domain.User{
			Email:    "me@example.com",
			Name:     "Me User",
			Password: "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		gotUser, err := userRepo.GetByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, gotUser.ID)
	})
}
