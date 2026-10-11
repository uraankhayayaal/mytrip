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

func TestUserRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Запуск тестового PostgreSQL
	postgresContainer, err := postgres.RunContainer(ctx, testcontainers.WithImage("postgres:16-alpine"))
	require.NoError(t, err)
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	// Подключение к БД
	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Создание репозитория
	repo, err := NewUserRepo(connStr)
	require.NoError(t, err)

	// Создание пользователя
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	createdUser, err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotZero(t, createdUser.ID)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.Equal(t, user.Name, createdUser.Name)
}

func TestUserRepo_GetByEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Запуск тестового PostgreSQL
	postgresContainer, err := postgres.RunContainer(ctx, testcontainers.WithImage("postgres:16-alpine"))
	require.NoError(t, err)
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	// Подключение к БД
	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Создание репозитория
	repo, err := NewUserRepo(connStr)
	require.NoError(t, err)

	// Создание пользователя
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	createdUser, err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Поиск пользователя по email
	foundUser, err := repo.GetByEmail(ctx, user.Email)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, user.Email, foundUser.Email)
}

func TestUserRepo_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Запуск тестового PostgreSQL
	postgresContainer, err := postgres.RunContainer(ctx, testcontainers.WithImage("postgres:16-alpine"))
	require.NoError(t, err)
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	// Подключение к БД
	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Создание репозитория
	repo, err := NewUserRepo(connStr)
	require.NoError(t, err)

	// Создание пользователя
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	createdUser, err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Поиск пользователя по ID
	foundUser, err := repo.GetByID(ctx, createdUser.ID)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, user.Email, foundUser.Email)
}