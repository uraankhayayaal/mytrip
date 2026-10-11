package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTripsIntegration(t *testing.T) {
	// Тесты для trips эндпоинтов
	// CRUD trips (201, 200, 404, 400)
	// CRUD stops (201, 200, 404, 400)

	t.Run("create trip", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/trips", nil)
		w := httptest.NewRecorder()
		// Вызов обработчика
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("get trip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/1", nil)
		w := httptest.NewRecorder()
		// Вызов обработчика
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
