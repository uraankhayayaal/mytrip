//go:build unit_test

package repo

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.postgresql.org/pg/v4"

	"mytrip/internal/domain"
)

func TestUserRepo_Create(t *testing.T) {
	// Arrange
	ctx := context.Background()
	
	// Create test container
	container, err := postgres.RunContainer(ctx, testcontainers.WithImage("postgres:16-alpine"))
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	// Initialize repo
	repo := NewUserRepo(connStr)

	// Create user
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	// Act
	createdUser, err := repo.Create(ctx, user)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.NotEmpty(t, createdUser.ID)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.Equal(t, user.Name, createdUser.Name)
}

func TestUserRepo_GetByEmail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	
	// Create test container
	container, err := postgres.RunContainer(ctx, testcontainers.WithImage("postgres:16-alpine"))
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	// Initialize repo
	repo := NewUserRepo(connStr)

	// Create user first
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Act
	foundUser, err := repo.GetByEmail(ctx, user.Email)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, createdUser.ID, foundUser.ID)
}

func TestUserRepo_GetByID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	
	// Create test container
	container, err := postgres.RunContainer(ctx, testcontainers.WithImage("postgres:16-alpine"))
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	// Initialize repo
	repo := NewUserRepo(connStr)

	// Create user
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Act
	foundUser, err := repo.GetByID(ctx, createdUser.ID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, createdUser.ID, foundUser.ID)
}