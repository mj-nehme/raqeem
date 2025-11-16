package reliability

import (
"context"
"errors"
"testing"
"time"
)

func TestRetryWithBackoff_Success(t *testing.T) {
config := DefaultRetryConfig()
attempts := 0

operation := func() error {
attempts++
if attempts < 2 {
return errors.New("temporary error")
}
return nil
}

ctx := context.Background()
err := RetryWithBackoff(ctx, config, operation)

if err != nil {
t.Errorf("Expected success, got error: %v", err)
}

if attempts != 2 {
t.Errorf("Expected 2 attempts, got %d", attempts)
}
}

func TestRetryWithBackoff_MaxAttemptsExceeded(t *testing.T) {
config := RetryConfig{
MaxAttempts:   3,
InitialDelay:  10 * time.Millisecond,
MaxDelay:      100 * time.Millisecond,
BackoffFactor: 2.0,
}
attempts := 0

operation := func() error {
attempts++
return errors.New("persistent error")
}

ctx := context.Background()
err := RetryWithBackoff(ctx, config, operation)

if err == nil {
t.Error("Expected error, got nil")
}

if attempts != 3 {
t.Errorf("Expected 3 attempts, got %d", attempts)
}
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
config := DefaultRetryConfig()
ctx, cancel := context.WithCancel(context.Background())

operation := func() error {
time.Sleep(50 * time.Millisecond)
return errors.New("error")
}

// Cancel context after a short delay
go func() {
time.Sleep(100 * time.Millisecond)
cancel()
}()

err := RetryWithBackoff(ctx, config, operation)

if err == nil {
t.Error("Expected error due to context cancellation")
}
}

func TestRetryWithTimeout(t *testing.T) {
config := DefaultRetryConfig()
attempts := 0

operation := func() error {
attempts++
time.Sleep(100 * time.Millisecond)
return errors.New("error")
}

err := RetryWithTimeout(200*time.Millisecond, config, operation)

if err == nil {
t.Error("Expected error")
}

// Should have attempted at least once
if attempts == 0 {
t.Error("Expected at least one attempt")
}
}

func TestRetryConfig_Presets(t *testing.T) {
defaultConfig := DefaultRetryConfig()
if defaultConfig.MaxAttempts != 3 {
t.Errorf("Expected default max attempts 3, got %d", defaultConfig.MaxAttempts)
}

dbConfig := DatabaseRetryConfig()
if dbConfig.MaxAttempts != 5 {
t.Errorf("Expected database max attempts 5, got %d", dbConfig.MaxAttempts)
}

serviceConfig := ExternalServiceRetryConfig()
if serviceConfig.MaxAttempts != 3 {
t.Errorf("Expected service max attempts 3, got %d", serviceConfig.MaxAttempts)
}
}
