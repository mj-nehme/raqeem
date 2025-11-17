package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARNING, "WARNING"},
		{ERROR, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"debug", DEBUG},
		{"INFO", INFO},
		{"info", INFO},
		{"WARNING", WARNING},
		{"WARN", WARNING},
		{"ERROR", ERROR},
		{"error", ERROR},
		{"invalid", INFO}, // defaults to INFO
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseLogLevel(tt.input); got != tt.expected {
				t.Errorf("parseLogLevel(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      "DEBUG",
		JSONFormat: false,
		Output:     &buf,
	}

	logger := NewLogger(config)
	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}

	if logger.level != DEBUG {
		t.Errorf("logger.level = %v, want %v", logger.level, DEBUG)
	}

	if logger.jsonFormat {
		t.Error("logger.jsonFormat = true, want false")
	}
}

func TestLogger_WithField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	newLogger := logger.WithField("key", "value")
	if newLogger == logger {
		t.Error("WithField should return a new logger instance")
	}

	if _, exists := newLogger.fields["key"]; !exists {
		t.Error("Field 'key' not found in new logger")
	}

	if val := newLogger.fields["key"]; val != "value" {
		t.Errorf("Field 'key' = %v, want 'value'", val)
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	newLogger := logger.WithFields(fields)
	if newLogger == logger {
		t.Error("WithFields should return a new logger instance")
	}

	if len(newLogger.fields) != 2 {
		t.Errorf("newLogger.fields has %d fields, want 2", len(newLogger.fields))
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Log output should contain 'INFO'")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Log output should contain 'test message'")
	}
}

func TestLogger_Debug_FilteredByLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	logger.Debug("debug message")

	output := buf.String()
	if output != "" {
		t.Errorf("Debug message should not be logged when level is INFO, got: %s", output)
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:      "INFO",
		JSONFormat: true,
		Output:     &buf,
	})

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("JSON output should contain level field")
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Error("JSON output should contain message field")
	}
}
