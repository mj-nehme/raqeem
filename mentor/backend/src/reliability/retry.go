package reliability

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RetryConfig defines configuration for retry behavior
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// DatabaseRetryConfig returns retry configuration optimized for database operations
func DatabaseRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  200 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
	}
}

// ExternalServiceRetryConfig returns retry configuration for external service calls
func ExternalServiceRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Try the operation
		err := operation()
		if err == nil {
			if attempt > 1 {
				log.Printf("Operation succeeded after %d attempts", attempt)
			}
			return nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == config.MaxAttempts {
			break
		}

		log.Printf("Operation failed (attempt %d/%d): %v. Retrying in %v...",
			attempt, config.MaxAttempts, err, delay)

		// Wait with exponential backoff
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.BackoffFactor)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithTimeout wraps RetryWithBackoff with a timeout
func RetryWithTimeout(timeout time.Duration, config RetryConfig, operation func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return RetryWithBackoff(ctx, config, operation)
}
