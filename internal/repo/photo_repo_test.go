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

func TestPhotoRepo_Create(t *testing.T) {
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
	photoRepo := NewPhotoRepo(connStr)

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

	// Create photo
	photo := &domain.Photo{
		TripID:   createdTrip.ID,
		Name:     "test-photo.jpg",
		URL:      "http://example.com/test-photo.jpg",
		Size:     1024,
		MimeType: "image/jpeg",
	}

	// Act
	createdPhoto, err := photoRepo.Create(ctx, photo)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, createdPhoto)
	assert.NotEmpty(t, createdPhoto.ID)
	assert.Equal(t, photo.Name, createdPhoto.Name)
}

func TestPhotoRepo_GetByID(t *testing.T) {
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
	photoRepo := NewPhotoRepo(connStr)

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

	// Create photo
	photo := &domain.Photo{
		TripID:   createdTrip.ID,
		Name:     "test-photo.jpg",
		URL:      "http://example.com/test-photo.jpg",
		Size:     1024,
		MimeType: "image/jpeg",
	}
	createdPhoto, err := photoRepo.Create(ctx, photo)
	require.NoError(t, err)

	// Act
	foundPhoto, err := photoRepo.GetByID(ctx, createdPhoto.ID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, foundPhoto)
	assert.Equal(t, createdPhoto.ID, foundPhoto.ID)
}

func TestPhotoRepo_Delete(t *testing.T) {
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
	photoRepo := NewPhotoRepo(connStr)

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

	// Create photo
	photo := &domain.Photo{
		TripID:   createdTrip.ID,
		Name:     "test-photo.jpg",
		URL:      "http://example.com/test-photo.jpg",
		Size:     1024,
		MimeType: "image/jpeg",
	}
	createdPhoto, err := photoRepo.Create(ctx, photo)
	require.NoError(t, err)

	// Act
	err := photoRepo.Delete(ctx, createdPhoto.ID)

	// Assert
	require.NoError(t, err)

	// Verify photo is deleted
	_, err = photoRepo.GetByID(ctx, createdPhoto.ID)
	assert.Error(t, err)
}