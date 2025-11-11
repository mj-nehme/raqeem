package docs

import (
	"strings"
	"testing"
)

func TestGetDocTemplate(t *testing.T) {
	// Test that the template loads successfully
	template := getDocTemplate()

	if template == "" {
		t.Error("Expected non-empty template, got empty string")
	}

	// Check that it contains expected swagger content
	if !strings.Contains(template, "swagger") {
		t.Error("Expected template to contain 'swagger'")
	}

	if !strings.Contains(template, "schemes") {
		t.Error("Expected template to contain 'schemes'")
	}

	if !strings.Contains(template, "definitions") {
		t.Error("Expected template to contain 'definitions'")
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
	template := getDocTemplate()

	// Check for expected API paths
	expectedPaths := []string{"/devices", "/activities", "alerts", "metrics"}

	for _, path := range expectedPaths {
		if !strings.Contains(template, path) {
			t.Errorf("Expected template to contain path '%s'", path)
		}
	}
}

func TestTemplateContainsExpectedModels(t *testing.T) {
	template := getDocTemplate()

	// Check for expected model definitions
	expectedModels := []string{"models.Device", "models.DeviceAlerts", "models.DeviceMetrics"}

	for _, model := range expectedModels {
		if !strings.Contains(template, model) {
			t.Errorf("Expected template to contain model '%s'", model)
		}
	}
}
