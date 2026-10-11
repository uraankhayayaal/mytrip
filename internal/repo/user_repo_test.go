package repo

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mytrip/internal/domain"
)

func TestUserRepo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Запуск testcontainers PostgreSQL
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	require.NoError(t, err)
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	// Подключение к базе данных
	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	// Создание репозитория
	repo := NewUserRepo(pool)

	// Тесты
	t.Run("Create", func(t *testing.T) {
		user := &domain.User{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("GetByID", func(t *testing.T) {
		user, err := repo.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
	})
}

func TestUserRepoConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Запуск testcontainers PostgreSQL
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	require.NoError(t, err)
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	// Подключение к базе данных
	connStr, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	// Создание репозитория
	repo := NewUserRepo(pool)

	// Создаем пользователя
	user := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Пытаемся создать пользователя с тем же email
	user2 := &domain.User{
		Email:    "test@example.com",
		Name:     "Test User 2",
		Password: "password123",
	}

	err = repo.Create(ctx, user2)
	assert.Error(t, err)
}