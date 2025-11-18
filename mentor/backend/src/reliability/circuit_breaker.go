package reliability

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed means requests pass through normally
	StateClosed CircuitState = iota
	// StateOpen means requests are rejected immediately
	StateOpen
	// StateHalfOpen means limited requests are allowed to test if the service recovered
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig defines configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening the circuit
	MaxFailures int
	// Timeout is how long to wait before transitioning from Open to Half-Open
	Timeout time.Duration
	// MaxHalfOpenRequests is the number of requests to allow in Half-Open state
	MaxHalfOpenRequests int
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:         5,
		Timeout:             30 * time.Second,
		MaxHalfOpenRequests: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config            CircuitBreakerConfig
	state             CircuitState
	failures          int
	lastFailureTime   time.Time
	halfOpenRequests  int
	halfOpenSuccesses int
	mu                sync.RWMutex
	name              string
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		name:   name,
	}
}

// Execute runs the given operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	// Check if we should allow the request
	if !cb.allowRequest() {
		return fmt.Errorf("%w: circuit breaker '%s' is open", ErrCircuitOpen, cb.name)
	}

	// Execute the operation
	err := operation()

	// Record the result
	cb.recordResult(err)

	return err
}

// allowRequest checks if a request should be allowed based on the current state
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if we should transition to Half-Open
		if now.Sub(cb.lastFailureTime) >= cb.config.Timeout {
			log.Printf("Circuit breaker '%s': Transitioning from Open to Half-Open", cb.name)
			cb.state = StateHalfOpen
			cb.halfOpenRequests = 0
			cb.halfOpenSuccesses = 0
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests in Half-Open state
		if cb.halfOpenRequests < cb.config.MaxHalfOpenRequests {
			cb.halfOpenRequests++
			return true
		}
		return false

	default:
		return false
	}
}

// recordResult updates the circuit breaker state based on the operation result
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onSuccess handles a successful operation
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success in Closed state
		cb.failures = 0

	case StateHalfOpen:
		cb.halfOpenSuccesses++
		// If we've had enough successes, close the circuit
		if cb.halfOpenSuccesses >= cb.config.MaxHalfOpenRequests {
			log.Printf("Circuit breaker '%s': Transitioning from Half-Open to Closed", cb.name)
			cb.state = StateClosed
			cb.failures = 0
			cb.halfOpenRequests = 0
			cb.halfOpenSuccesses = 0
		}
	}
}

// onFailure handles a failed operation
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			log.Printf("Circuit breaker '%s': Transitioning from Closed to Open after %d failures",
				cb.name, cb.failures)
			cb.state = StateOpen
		}

	case StateHalfOpen:
		// Any failure in Half-Open state reopens the circuit
		log.Printf("Circuit breaker '%s': Transitioning from Half-Open to Open due to failure", cb.name)
		cb.state = StateOpen
		cb.halfOpenRequests = 0
		cb.halfOpenSuccesses = 0
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to Closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	log.Printf("Circuit breaker '%s': Manual reset to Closed state", cb.name)
	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenRequests = 0
	cb.halfOpenSuccesses = 0
}
