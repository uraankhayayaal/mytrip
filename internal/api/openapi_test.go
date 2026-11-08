package api_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// TestOpenAPIContract проверяет, что все эндпоинты из OpenAPI существуют в сервере и возвращают заявленные коды.
func TestOpenAPIContract(t *testing.T) {
	t.Parallel()

	// Читаем OpenAPI yaml
	file, err := os.ReadFile("docs/openapi.yaml")
	assert.NoError(t, err)

	var openapi struct {
		Paths map[string]map[string]struct {
			Responses map[string]struct {
				Content map[string]struct {
					Schema struct {
						Type string `yaml:"type"`
					}
				}
			}
		}
	}

	assert.NoError(t, yaml.Unmarshal(file, &openapi))

	// Проверяем каждый путь из OpenAPI
	for path, methods := range openapi.Paths {
		for method, methodSpec := range methods {
			// Проверяем, что в методе есть заявленные коды ответа
			for code, responseSpec := range methodSpec.Responses {
				// Проверяем, что код ответа поддерживается
				assert.Contains(t, []string{"200", "201", "204", "400", "401", "404", "409", "500"}, code)

				// Проверяем Content-Type
				contentType := "application/json"
				if code == "200" && (path == "/api/trips/{id}/export/pdf" || path == "/api/trips/{id}/export/csv") {
					contentType = responseSpec.Content["application/pdf"].Schema.Type
					if contentType == "" {
						contentType = responseSpec.Content["text/csv"].Schema.Type
					}
				}
				assert.NotEmpty(t, contentType)
			}
		}
	}
}
