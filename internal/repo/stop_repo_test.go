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

func TestStopRepo_Create(t *testing.T) {
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
	stopRepo := NewStopRepo(connStr)

	// Create user
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
	createdTrip, err := tripRepo.Create(ctx, trip)
	require.NoError(t, err)

	// Create stop
	stop := &domain.Stop{
		Name:     "Test Stop",
		TripID:   createdTrip.ID,
		Location: &domain.Location{
			Lat: 10.0,
			Lng: 20.0,
		},
	}

	// Act
	createdStop, err := stopRepo.Create(ctx, stop)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, createdStop)
	assert.NotEmpty(t, createdStop.ID)
	assert.Equal(t, stop.Name, createdStop.Name)
}

func TestStopRepo_GetByID(t *testing.T) {
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
	stopRepo := NewStopRepo(connStr)

	// Create user
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
	createdTrip, err := tripRepo.Create(ctx, trip)
	require.NoError(t, err)

	// Create stop
	stop := &domain.Stop{
		Name:     "Test Stop",
		TripID:   createdTrip.ID,
		Location: &domain.Location{
			Lat: 10.0,
			Lng: 20.0,
		},
	}
	createdStop, err := stopRepo.Create(ctx, stop)
	require.NoError(t, err)

	// Act
	foundStop, err := stopRepo.GetByID(ctx, createdStop.ID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, foundStop)
	assert.Equal(t, createdStop.ID, foundStop.ID)
}

func TestStopRepo_ListByTrip(t *testing.T) {
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
	stopRepo := NewStopRepo(connStr)

	// Create user
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
	createdTrip, err := tripRepo.Create(ctx, trip)
	require.NoError(t, err)

	// Create stops
	stop1 := &domain.Stop{
		Name:     "Test Stop 1",
		TripID:   createdTrip.ID,
		Location: &domain.Location{
			Lat: 10.0,
			Lng: 20.0,
		},
	}
	stop2 := &domain.Stop{
		Name:     "Test Stop 2",
		TripID:   createdTrip.ID,
		Location: &domain.Location{
			Lat: 15.0,
			Lng: 25.0,
		},
	}

	_, err = stopRepo.Create(ctx, stop1)
	require.NoError(t, err)
	_, err = stopRepo.Create(ctx, stop2)
	require.NoError(t, err)

	// Act
	stops, err := stopRepo.ListByTrip(ctx, createdTrip.ID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, stops, 2)
}