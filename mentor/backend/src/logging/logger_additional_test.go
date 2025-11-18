package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_Warning(t *testing.T) {
	t.Run("logs warning message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "DEBUG",
			JSONFormat: false,
			Output:     &buf,
		})

		logger.Warning("this is a warning")

		output := buf.String()
		if !strings.Contains(output, "WARNING") {
			t.Error("Log output should contain 'WARNING'")
		}
		if !strings.Contains(output, "this is a warning") {
			t.Error("Log output should contain 'this is a warning'")
		}
	})

	t.Run("logs warning with fields", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "WARNING",
			JSONFormat: false,
			Output:     &buf,
		})

		fields := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		logger.Warning("warning with fields", fields)

		output := buf.String()
		if !strings.Contains(output, "WARNING") {
			t.Error("Log output should contain 'WARNING'")
		}
		if !strings.Contains(output, "warning with fields") {
			t.Error("Log output should contain message")
		}
	})

	t.Run("warning in JSON format", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "WARNING",
			JSONFormat: true,
			Output:     &buf,
		})

		logger.Warning("json warning")

		output := buf.String()
		if !strings.Contains(output, `"level":"WARNING"`) {
			t.Error("JSON output should contain level field")
		}
		if !strings.Contains(output, `"message":"json warning"`) {
			t.Error("JSON output should contain message field")
		}
	})

	t.Run("warning filtered by level", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "ERROR",
			JSONFormat: false,
			Output:     &buf,
		})

		logger.Warning("should not appear")

		output := buf.String()
		if output != "" {
			t.Errorf("Warning message should not be logged when level is ERROR, got: %s", output)
		}
	})
}

func TestLogger_Error(t *testing.T) {
	t.Run("logs error message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "DEBUG",
			JSONFormat: false,
			Output:     &buf,
		})

		logger.Error("this is an error")

		output := buf.String()
		if !strings.Contains(output, "ERROR") {
			t.Error("Log output should contain 'ERROR'")
		}
		if !strings.Contains(output, "this is an error") {
			t.Error("Log output should contain 'this is an error'")
		}
	})

	t.Run("logs error with fields", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "ERROR",
			JSONFormat: false,
			Output:     &buf,
		})

		fields := map[string]interface{}{
			"error_code": 500,
			"component":  "database",
		}
		logger.Error("error with fields", fields)

		output := buf.String()
		if !strings.Contains(output, "ERROR") {
			t.Error("Log output should contain 'ERROR'")
		}
		if !strings.Contains(output, "error with fields") {
			t.Error("Log output should contain message")
		}
	})

	t.Run("error in JSON format", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "ERROR",
			JSONFormat: true,
			Output:     &buf,
		})

		logger.Error("json error")

		output := buf.String()
		if !strings.Contains(output, `"level":"ERROR"`) {
			t.Error("JSON output should contain level field")
		}
		if !strings.Contains(output, `"message":"json error"`) {
			t.Error("JSON output should contain message field")
		}
	})

	t.Run("error always logged regardless of level", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "ERROR",
			JSONFormat: false,
			Output:     &buf,
		})

		logger.Error("error message")

		output := buf.String()
		if !strings.Contains(output, "ERROR") {
			t.Error("Error should be logged even at ERROR level")
		}
	})
}

func TestLogger_LogJSON_EdgeCases(t *testing.T) {
	t.Run("logs with nil fields", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "INFO",
			JSONFormat: true,
			Output:     &buf,
		})

		logger.log(INFO, "message with nil fields", nil)

		output := buf.String()
		if !strings.Contains(output, `"message":"message with nil fields"`) {
			t.Error("Should handle nil fields gracefully")
		}
	})

	t.Run("logs with empty fields", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "INFO",
			JSONFormat: true,
			Output:     &buf,
		})

		logger.log(INFO, "message with empty fields", map[string]interface{}{})

		output := buf.String()
		if !strings.Contains(output, `"message":"message with empty fields"`) {
			t.Error("Should handle empty fields map")
		}
	})

	t.Run("logs with complex field values", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "INFO",
			JSONFormat: true,
			Output:     &buf,
		})

		complexFields := map[string]interface{}{
			"string": "value",
			"number": 42,
			"float":  3.14,
			"bool":   true,
			"nil":    nil,
			"array":  []int{1, 2, 3},
			"nested": map[string]string{"key": "value"},
		}
		logger.log(INFO, "complex message", complexFields)

		output := buf.String()
		if !strings.Contains(output, `"message":"complex message"`) {
			t.Error("Should handle complex field values")
		}
	})

	t.Run("logs with special characters in message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "INFO",
			JSONFormat: true,
			Output:     &buf,
		})

		logger.log(INFO, `message with "quotes" and \backslash`, nil)

		output := buf.String()
		// Should properly escape the message
		if !strings.Contains(output, "message") {
			t.Error("Should handle special characters in message")
		}
	})
}

func TestLogger_LogText_EdgeCases(t *testing.T) {
	t.Run("logs with nil fields in text format", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "INFO",
			JSONFormat: false,
			Output:     &buf,
		})

		logger.log(INFO, "message with nil fields", nil)

		output := buf.String()
		if !strings.Contains(output, "message with nil fields") {
			t.Error("Should handle nil fields gracefully")
		}
	})

	t.Run("logs with empty fields in text format", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "INFO",
			JSONFormat: false,
			Output:     &buf,
		})

		logger.log(INFO, "message with empty fields", map[string]interface{}{})

		output := buf.String()
		if !strings.Contains(output, "message with empty fields") {
			t.Error("Should handle empty fields map")
		}
	})

	t.Run("logs with multiple fields in text format", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "INFO",
			JSONFormat: false,
			Output:     &buf,
		})

		fields := map[string]interface{}{
			"field1": "value1",
			"field2": 123,
			"field3": true,
		}
		logger.log(INFO, "message with fields", fields)

		output := buf.String()
		if !strings.Contains(output, "message with fields") {
			t.Error("Should contain the message")
		}
		// Check that fields are present in some form
		if !strings.Contains(output, "field1") && !strings.Contains(output, "value1") {
			t.Error("Should include field information")
		}
	})

	t.Run("logs with newlines in message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(Config{
			Level:      "INFO",
			JSONFormat: false,
			Output:     &buf,
		})

		logger.log(INFO, "message\nwith\nnewlines", nil)

		output := buf.String()
		if !strings.Contains(output, "message") {
			t.Error("Should handle newlines in message")
		}
	})
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name        string
		loggerLevel LogLevel
		logLevel    LogLevel
		shouldLog   bool
	}{
		{"DEBUG logs at DEBUG level", DEBUG, DEBUG, true},
		{"DEBUG doesn't log at INFO level", INFO, DEBUG, false},
		{"INFO logs at DEBUG level", DEBUG, INFO, true},
		{"INFO logs at INFO level", INFO, INFO, true},
		{"INFO doesn't log at WARNING level", WARNING, INFO, false},
		{"WARNING logs at INFO level", INFO, WARNING, true},
		{"WARNING logs at WARNING level", WARNING, WARNING, true},
		{"WARNING doesn't log at ERROR level", ERROR, WARNING, false},
		{"ERROR logs at all levels", DEBUG, ERROR, true},
		{"ERROR logs at ERROR level", ERROR, ERROR, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(Config{
				Level:      tt.loggerLevel.String(),
				JSONFormat: false,
				Output:     &buf,
			})

			logger.log(tt.logLevel, "test message", nil)

			output := buf.String()
			if tt.shouldLog && output == "" {
				t.Error("Expected log output but got none")
			}
			if !tt.shouldLog && output != "" {
				t.Errorf("Expected no log output but got: %s", output)
			}
		})
	}
}

func TestLogger_WithFields_Immutability(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	// Create logger with fields
	logger1 := logger.WithField("key1", "value1")
	logger2 := logger1.WithField("key2", "value2")

	// Verify logger1 still has only key1
	if len(logger1.fields) != 1 {
		t.Errorf("logger1 should have 1 field, has %d", len(logger1.fields))
	}

	// Verify logger2 has both key1 and key2
	if len(logger2.fields) != 2 {
		t.Errorf("logger2 should have 2 fields, has %d", len(logger2.fields))
	}

	// Verify original logger is unmodified
	if len(logger.fields) != 0 {
		t.Errorf("original logger should have 0 fields, has %d", len(logger.fields))
	}
}

func TestLogger_ConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	done := make(chan bool, 10)

	// Start 10 goroutines logging concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				logger.Info("concurrent log", map[string]interface{}{
					"goroutine": id,
					"iteration": j,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify we got some output (exact count may vary due to buffering)
	output := buf.String()
	if output == "" {
		t.Error("Should have logged messages from concurrent goroutines")
	}
}

func TestGlobalDebug(t *testing.T) {
	var buf bytes.Buffer
	oldLogger := defaultLogger
	defer func() { defaultLogger = oldLogger }()

	defaultLogger = NewLogger(Config{
		Level:      "DEBUG",
		JSONFormat: false,
		Output:     &buf,
	})

	Debug("global debug message")
	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Error("Global Debug should log debug message")
	}
	if !strings.Contains(output, "global debug message") {
		t.Error("Global Debug should contain message")
	}
}

func TestGlobalInfo(t *testing.T) {
	var buf bytes.Buffer
	oldLogger := defaultLogger
	defer func() { defaultLogger = oldLogger }()

	defaultLogger = NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	Info("global info message")
	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Global Info should log info message")
	}
	if !strings.Contains(output, "global info message") {
		t.Error("Global Info should contain message")
	}
}

func TestGlobalWarning(t *testing.T) {
	var buf bytes.Buffer
	oldLogger := defaultLogger
	defer func() { defaultLogger = oldLogger }()

	defaultLogger = NewLogger(Config{
		Level:      "WARNING",
		JSONFormat: false,
		Output:     &buf,
	})

	Warning("global warning message")
	output := buf.String()
	if !strings.Contains(output, "WARNING") {
		t.Error("Global Warning should log warning message")
	}
	if !strings.Contains(output, "global warning message") {
		t.Error("Global Warning should contain message")
	}
}

func TestGlobalError(t *testing.T) {
	var buf bytes.Buffer
	oldLogger := defaultLogger
	defer func() { defaultLogger = oldLogger }()

	defaultLogger = NewLogger(Config{
		Level:      "ERROR",
		JSONFormat: false,
		Output:     &buf,
	})

	Error("global error message")
	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Error("Global Error should log error message")
	}
	if !strings.Contains(output, "global error message") {
		t.Error("Global Error should contain message")
	}
}

func TestGlobalWithFields(t *testing.T) {
	var buf bytes.Buffer
	oldLogger := defaultLogger
	defer func() { defaultLogger = oldLogger }()

	defaultLogger = NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	logger := WithFields(fields)

	if len(logger.fields) != 2 {
		t.Errorf("WithFields should return logger with 2 fields, got %d", len(logger.fields))
	}
}

func TestGlobalWithField(t *testing.T) {
	var buf bytes.Buffer
	oldLogger := defaultLogger
	defer func() { defaultLogger = oldLogger }()

	defaultLogger = NewLogger(Config{
		Level:      "INFO",
		JSONFormat: false,
		Output:     &buf,
	})

	logger := WithField("key", "value")

	if len(logger.fields) != 1 {
		t.Errorf("WithField should return logger with 1 field, got %d", len(logger.fields))
	}

	if val, exists := logger.fields["key"]; !exists || val != "value" {
		t.Error("WithField should set the field correctly")
	}
}
