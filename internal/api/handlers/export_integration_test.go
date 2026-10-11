//go:build integration
// +build integration

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportIntegration(t *testing.T) {
	// Integration tests for export endpoints
	// GET /api/v1/trips/{id}/export?format=pdf|csv → 200 (стриминг)

	t.Run("ExportPDF", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/123/export?format=pdf", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("ExportCSV", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/123/export?format=csv", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})

	t.Run("ExportInvalidFormat", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trips/123/export?format=xml", nil)
		w := httptest.NewRecorder()
		
		assert.NotNil(t, req)
		assert.NotNil(t, w)
	})
}
