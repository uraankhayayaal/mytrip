package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotosIntegration(t *testing.T) {
	// Тесты для photos эндпоинтов
	// upload (201, 400)
	// presigned URL (200, 404)
	// delete (204, 404)

	t.Run("upload photo", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/trips/1/photos", nil)
		w := httptest.NewRecorder()
		// Вызов обработчика
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
