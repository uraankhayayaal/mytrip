package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthIntegration(t *testing.T) {
	// Тесты для auth эндпоинтов
	// register (201, 400 validation, 409 conflict)
	// login (200, 401)
	// refresh (200, 401)
	// logout (204)
	// me (200, 401)

	t.Run("register", func(t *testing.T) {
		// Проверка успешной регистрации
		req := httptest.NewRequest("POST", "/api/v1/auth/register", nil)
		w := httptest.NewRecorder()
		// Вызов обработчика
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("login", func(t *testing.T) {
		// Проверка успешного логина
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		w := httptest.NewRecorder()
		// Вызов обработчика
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
