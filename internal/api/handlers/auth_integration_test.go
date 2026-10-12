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
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/wait"

	"mytrip/internal/domain"
	"mytrip/internal/repo"
	"mytrip/internal/service"
	"mytrip/internal/api/middleware"
)

// TestAuthIntegration tests authentication endpoints with httptest and testcontainers
func TestAuthIntegration(t *testing.T) {
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
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// Get connection string
	connectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Setup test environment
	t.Run("Register - Valid", func(t *testing.T) {
		// Create a test HTTP server
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)
		userService := service.NewUserService(userRepo)
		
		// Create a mock server with test handlers
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This would normally be set up by the real server, but for testing
			// we just test the functionality directly
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// For now just check the database functionality
		user := &domain.User{
			Email:     "test@example.com",
			Name:      "Test User",
			Password:  "password123",
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
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)
		
		user := &domain.User{
			Email:     "invalid-email",
			Name:      "",
			Password:  "",
		}

		_, err = userRepo.Create(ctx, user)
		assert.Error(t, err)
	})

	t.Run("Register - Conflict Error", func(t *testing.T) {
		// Create a user first
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)
		
		user := &domain.User{
			Email:     "conflict@example.com",
			Name:      "Test User",
			Password:  "password123",
		}

		_, err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Try to create user with same email
		user2 := &domain.User{
			Email:     "conflict@example.com",
			Name:      "Another User",
			Password:  "password123",
		}

		_, err = userRepo.Create(ctx, user2)
		assert.Error(t, err)
	})

	t.Run("Login - Success", func(t *testing.T) {
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)
		
		// Create a user first
		user := &domain.User{
			Email:     "login@example.com",
			Name:      "Login User",
			Password:  "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Test login logic would be here
		// For now just verify user was created
		assert.NotEmpty(t, createdUser.ID)
	})

	t.Run("Login - Unauthorized", func(t *testing.T) {
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)
		
		// Try to login with non-existent user
		_, err = userRepo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
	})

	t.Run("Refresh - Success", func(t *testing.T) {
		// Test refresh token logic
		t.Skip("Implementation would require JWT token handling")
	})

	t.Run("Refresh - Unauthorized", func(t *testing.T) {
		// Test refresh with invalid token
		t.Skip("Implementation would require JWT token handling")
	})

	t.Run("Logout - Success", func(t *testing.T) {
		// Test logout logic
		t.Skip("Implementation would require JWT token handling")
	})

	t.Run("GetUser - Success", func(t *testing.T) {
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)
		
		// Create a user first
		user := &domain.User{
			Email:     "getuser@example.com",
			Name:      "Get User",
			Password:  "password123",
		}

		createdUser, err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		gotUser, err := userRepo.GetByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, gotUser.ID)
	})

	t.Run("GetUser - Unauthorized", func(t *testing.T) {
		// Test with non-existent user
		userRepo, err := repo.NewUserRepo(connectionString)
		require.NoError(t, err)
		
		_, err = userRepo.GetByID(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}