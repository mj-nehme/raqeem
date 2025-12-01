package docs

import (
	"strings"
	"testing"
)

func TestGetDocTemplate(t *testing.T) {
	// Test that the SwaggerInfo is initialized with template
	if SwaggerInfo == nil {
		t.Fatal("SwaggerInfo is not initialized")
	}

	template := SwaggerInfo.SwaggerTemplate

	if template == "" {
		t.Error("Expected non-empty template, got empty string")
	}

	// Check that it contains expected OpenAPI 3.0 content
	if !strings.Contains(template, "openapi") {
		t.Error("Expected template to contain 'openapi'")
	}

	if !strings.Contains(template, "servers") {
		t.Error("Expected template to contain 'servers'")
	}

	if !strings.Contains(template, "components") {
		t.Error("Expected template to contain 'components'")
	}
}

func TestSwaggerInfoInitialization(t *testing.T) {
	if SwaggerInfo == nil {
		t.Error("Expected SwaggerInfo to be initialized")
	}

	if SwaggerInfo.Title == "" {
		t.Error("Expected non-empty title")
	}

	if SwaggerInfo.Version == "" {
		t.Error("Expected non-empty version")
	}

	if SwaggerInfo.SwaggerTemplate == "" {
		t.Error("Expected non-empty swagger template")
	}
}

func TestTemplateContainsExpectedPaths(t *testing.T) {
	if SwaggerInfo == nil {
		t.Fatal("SwaggerInfo is not initialized")
	}

	template := SwaggerInfo.SwaggerTemplate

	// Check for expected API paths
	expectedPaths := []string{"/devices", "/activities", "alerts", "metrics"}

	for _, path := range expectedPaths {
		if !strings.Contains(template, path) {
			t.Errorf("Expected template to contain path '%s'", path)
		}
	}
}

func TestTemplateContainsExpectedModels(t *testing.T) {
	if SwaggerInfo == nil {
		t.Fatal("SwaggerInfo is not initialized")
	}

	template := SwaggerInfo.SwaggerTemplate

	// Check for expected model definitions (OpenAPI 3.0 uses components/schemas without models. prefix)
	expectedModels := []string{"Device", "DeviceAlert", "DeviceMetric"}

	for _, model := range expectedModels {
		if !strings.Contains(template, model) {
			t.Errorf("Expected template to contain model '%s'", model)
		}
	}
}
