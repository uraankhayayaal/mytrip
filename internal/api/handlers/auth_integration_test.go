//go:build integration
// +build integration

package handlers

import (
	"bytes"
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

	// Create repositories
	userRepo, err := repo.NewUserRepo(connectionString)
	require.NoError(t, err)

	t.Run("Register - Valid", func(t *testing.T) {
		// Create test user
		user := &domain.User{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		// Test registration
		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, createdUser.ID)
		assert.Equal(t, user.Email, createdUser.Email)
		assert.Equal(t, user.Name, createdUser.Name)
	})

	t.Run("Register - Validation Error", func(t *testing.T) {
		// Test registration with invalid data
		user := &domain.User{
			Email:    "",
			Name:     "",
			Password: "",
		}

		_, err := userRepo.Create(ctx, user)
		assert.Error(t, err)
	})

	t.Run("Register - Conflict Error", func(t *testing.T) {
		// Try to register user with existing email
		user := &domain.User{
			Email:    "test@example.com",
			Name:     "Another User",
			Password: "password123",
		}

		_, err := userRepo.Create(ctx, user)
		assert.Error(t, err)
	})

	t.Run("Login - Success", func(t *testing.T) {
		// This would test actual auth flow with JWT
		t.Skip("JWT handling not implemented in this test")
	})

	t.Run("Login - Unauthorized", func(t *testing.T) {
		t.Skip("JWT handling not implemented in this test")
	})

	t.Run("Refresh - Success", func(t *testing.T) {
		t.Skip("JWT handling not implemented in this test")
	})

	t.Run("Refresh - Unauthorized", func(t *testing.T) {
		t.Skip("JWT handling not implemented in this test")
	})

	t.Run("Logout - Success", func(t *testing.T) {
		t.Skip("JWT handling not implemented in this test")
	})

	t.Run("GetUser - Success", func(t *testing.T) {
		t.Skip("JWT handling not implemented in this test")
	})

	t.Run("GetUser - Unauthorized", func(t *testing.T) {
		t.Skip("JWT handling not implemented in this test")
	})
}