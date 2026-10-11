// +build integration

package handlers

import (
	"testing"
)

func TestPhotosIntegration(t *testing.T) {
	// Integration tests for photos endpoints
	// POST /api/v1/trips/{id}/photos → 201 {photo_url}
	// GET /api/v1/photos/{id}/url → 200 {url, expires_at}
	// DELETE /api/v1/photos/{id} → 204
}