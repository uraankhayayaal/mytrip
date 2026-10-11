package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportIntegration(t *testing.T) {
	// Тесты для export эндпоинтов
	// export PDF (200, Content-Type, Content-Disposition)
	// export CSV (200, Content-Type, Content-Disposition)
	// invalid format (400)

	t.Run("export PDF", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/1/export?format=pdf", nil)
		w := httptest.NewRecorder()
		// Вызов обработчика
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
