//go:build unit_test

package repo

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mytrip/internal/domain"
)

func TestTripRepo_Create(t *testing.T) {
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
	repo := NewTripRepo(connStr)

	// Create user first
	userRepo := NewUserRepo(connStr)
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create trip
	trip := &domain.Trip{
		Name:        "Test Trip",
		Description: "Test Description",
		UserID:      createdUser.ID,
	}

	// Act
	createdTrip, err := repo.Create(ctx, trip)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, createdTrip)
	assert.NotEmpty(t, createdTrip.ID)
	assert.Equal(t, trip.Name, createdTrip.Name)
}

func TestTripRepo_GetByID(t *testing.T) {
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
	repo := NewTripRepo(connStr)

	// Create user first
	userRepo := NewUserRepo(connStr)
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create trip
	trip := &domain.Trip{
		Name:        "Test Trip",
		Description: "Test Description",
		UserID:      createdUser.ID,
	}
	createdTrip, err := repo.Create(ctx, trip)
	require.NoError(t, err)

	// Act
	foundTrip, err := repo.GetByID(ctx, createdTrip.ID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, foundTrip)
	assert.Equal(t, createdTrip.ID, foundTrip.ID)
}

func TestTripRepo_ListByUser(t *testing.T) {
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

	// Initialize repos
	userRepo := NewUserRepo(connStr)
	tripRepo := NewTripRepo(connStr)

	// Create user
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create trips
	trip1 := &domain.Trip{
		Name:        "Test Trip 1",
		Description: "Test Description 1",
		UserID:      createdUser.ID,
	}
	trip2 := &domain.Trip{
		Name:        "Test Trip 2",
		Description: "Test Description 2",
		UserID:      createdUser.ID,
	}

	_, err = tripRepo.Create(ctx, trip1)
	require.NoError(t, err)
	_, err = tripRepo.Create(ctx, trip2)
	require.NoError(t, err)

	// Act
	trips, err := tripRepo.ListByUser(ctx, createdUser.ID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, trips, 2)
}