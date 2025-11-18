// Package logging provides structured logging capabilities for the Mentor Backend.
//
// Features:
//   - Structured log format with JSON support
//   - Log levels (DEBUG, INFO, WARNING, ERROR)
//   - Contextual logging with fields
//   - Request ID tracking
//   - Performance monitoring
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DEBUG level for detailed diagnostic information
	DEBUG LogLevel = iota
	// INFO level for general informational messages
	INFO
	// WARNING level for warning messages
	WARNING
	// ERROR level for error messages
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality
type Logger struct {
	level      LogLevel
	output     io.Writer
	jsonFormat bool
	fields     map[string]interface{}
	mu         sync.Mutex // protects concurrent writes to output
}

// Config holds logger configuration
type Config struct {
	Level      string // "DEBUG", "INFO", "WARNING", "ERROR"
	JSONFormat bool   // true for JSON output, false for text
	Output     io.Writer
}

var defaultLogger *Logger

func init() {
	// Initialize default logger from environment
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "INFO"
	}
	jsonFormat := os.Getenv("LOG_FORMAT") == "json"

	defaultLogger = NewLogger(Config{
		Level:      levelStr,
		JSONFormat: jsonFormat,
		Output:     os.Stdout,
	})
}

// NewLogger creates a new Logger with the given configuration
func NewLogger(config Config) *Logger {
	level := parseLogLevel(config.Level)
	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	return &Logger{
		level:      level,
		output:     output,
		jsonFormat: config.JSONFormat,
		fields:     make(map[string]interface{}),
	}
}

// parseLogLevel converts a string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING", "WARN":
		return WARNING
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		level:      l.level,
		output:     l.output,
		jsonFormat: l.jsonFormat,
		fields:     newFields,
	}
}

// WithField returns a new logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

// log writes a log message at the given level
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	// Merge logger fields with message fields
	allFields := make(map[string]interface{})
	for k, v := range l.fields {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	if l.jsonFormat {
		l.logJSON(level, message, allFields)
	} else {
		l.logText(level, message, allFields)
	}
}

// logJSON writes a log message in JSON format
func (l *Logger) logJSON(level LogLevel, message string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"time":    time.Now().Format(time.RFC3339),
		"level":   level.String(),
		"message": message,
	}

	// Add fields
	for k, v := range fields {
		entry[k] = v
	}

	// Add caller information for errors
	if level == ERROR {
		_, file, line, ok := runtime.Caller(3)
		if ok {
			entry["caller"] = fmt.Sprintf("%s:%d", file, line)
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := fmt.Fprintln(l.output, string(data)); err != nil {
		log.Printf("Failed to write log entry: %v", err)
	}
}

// logText writes a log message in human-readable text format
func (l *Logger) logText(level LogLevel, message string, fields map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s [%s] %s", timestamp, level.String(), message))

	// Add fields
	if len(fields) > 0 {
		sb.WriteString(" |")
		for k, v := range fields {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
	}

	// Add caller information for errors
	if level == ERROR {
		_, file, line, ok := runtime.Caller(3)
		if ok {
			sb.WriteString(fmt.Sprintf(" | caller=%s:%d", file, line))
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := fmt.Fprintln(l.output, sb.String()); err != nil {
		log.Printf("Failed to write log entry: %v", err)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, message, f)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, message, f)
}

// Warning logs a warning message
func (l *Logger) Warning(message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARNING, message, f)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, message, f)
}

// Default logger functions for convenience

// Debug logs a debug message using the default logger
func Debug(message string, fields ...map[string]interface{}) {
	defaultLogger.Debug(message, fields...)
}

// Info logs an info message using the default logger
func Info(message string, fields ...map[string]interface{}) {
	defaultLogger.Info(message, fields...)
}

// Warning logs a warning message using the default logger
func Warning(message string, fields ...map[string]interface{}) {
	defaultLogger.Warning(message, fields...)
}

// Error logs an error message using the default logger
func Error(message string, fields ...map[string]interface{}) {
	defaultLogger.Error(message, fields...)
}

// WithFields returns a new logger with additional fields using the default logger
func WithFields(fields map[string]interface{}) *Logger {
	return defaultLogger.WithFields(fields)
}

// WithField returns a new logger with an additional field using the default logger
func WithField(key string, value interface{}) *Logger {
	return defaultLogger.WithField(key, value)
}
