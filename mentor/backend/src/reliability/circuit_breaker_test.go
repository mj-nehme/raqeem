package reliability

import (
"errors"
"testing"
"time"
)

func TestCircuitBreaker_InitialStateClosed(t *testing.T) {
cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())

if cb.State() != StateClosed {
t.Errorf("Expected initial state to be Closed, got %v", cb.State())
}
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
config := CircuitBreakerConfig{
MaxFailures:         3,
Timeout:             1 * time.Second,
MaxHalfOpenRequests: 2,
}
cb := NewCircuitBreaker("test", config)

// Cause enough failures to open the circuit
for i := 0; i < 3; i++ {
_ = cb.Execute(func() error {
return errors.New("test error")
})
}

if cb.State() != StateOpen {
t.Errorf("Expected state to be Open after %d failures, got %v", config.MaxFailures, cb.State())
}

// Next request should fail immediately
err := cb.Execute(func() error {
return nil
})

if !errors.Is(err, ErrCircuitOpen) {
t.Errorf("Expected ErrCircuitOpen, got %v", err)
}
}

func TestCircuitBreaker_TransitionToHalfOpen(t *testing.T) {
config := CircuitBreakerConfig{
MaxFailures:         2,
Timeout:             100 * time.Millisecond,
MaxHalfOpenRequests: 2,
}
cb := NewCircuitBreaker("test", config)

// Open the circuit
for i := 0; i < 2; i++ {
_ = cb.Execute(func() error {
return errors.New("test error")
})
}

if cb.State() != StateOpen {
t.Error("Expected state to be Open")
}

// Wait for timeout
time.Sleep(150 * time.Millisecond)

// Next request should be allowed (Half-Open state)
err := cb.Execute(func() error {
return nil
})

if err != nil {
t.Errorf("Expected request to be allowed in Half-Open state, got error: %v", err)
}
}

func TestCircuitBreaker_CloseAfterSuccesses(t *testing.T) {
config := CircuitBreakerConfig{
MaxFailures:         2,
Timeout:             100 * time.Millisecond,
MaxHalfOpenRequests: 2,
}
cb := NewCircuitBreaker("test", config)

// Open the circuit
for i := 0; i < 2; i++ {
_ = cb.Execute(func() error {
return errors.New("test error")
})
}

// Wait for timeout
time.Sleep(150 * time.Millisecond)

// Succeed in Half-Open state
for i := 0; i < 2; i++ {
err := cb.Execute(func() error {
return nil
})
if err != nil {
t.Errorf("Request %d failed: %v", i, err)
}
}

if cb.State() != StateClosed {
t.Errorf("Expected state to be Closed after successes, got %v", cb.State())
}
}

func TestCircuitBreaker_Reset(t *testing.T) {
config := CircuitBreakerConfig{
MaxFailures:         2,
Timeout:             10 * time.Second,
MaxHalfOpenRequests: 2,
}
cb := NewCircuitBreaker("test", config)

// Open the circuit
for i := 0; i < 2; i++ {
_ = cb.Execute(func() error {
return errors.New("test error")
})
}

if cb.State() != StateOpen {
t.Error("Expected state to be Open")
}

// Reset the circuit breaker
cb.Reset()

if cb.State() != StateClosed {
t.Errorf("Expected state to be Closed after reset, got %v", cb.State())
}

// Should allow requests now
err := cb.Execute(func() error {
return nil
})

if err != nil {
t.Errorf("Expected request to succeed after reset, got error: %v", err)
}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
config := CircuitBreakerConfig{
MaxFailures:         3,
Timeout:             1 * time.Second,
MaxHalfOpenRequests: 2,
}
cb := NewCircuitBreaker("test", config)

// Some failures
_ = cb.Execute(func() error {
return errors.New("error 1")
})
_ = cb.Execute(func() error {
return errors.New("error 2")
})

// Success should reset count
_ = cb.Execute(func() error {
return nil
})

// Should still be closed
if cb.State() != StateClosed {
t.Error("Expected state to remain Closed")
}

// Two more failures shouldn't open it (count was reset)
_ = cb.Execute(func() error {
return errors.New("error 3")
})
_ = cb.Execute(func() error {
return errors.New("error 4")
})

if cb.State() != StateClosed {
t.Error("Expected state to remain Closed")
}
}
