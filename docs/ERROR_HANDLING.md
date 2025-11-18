# Error Handling Guidelines

## Overview

This document outlines the error handling standards and best practices for the Raqeem platform. Consistent error handling improves reliability, debuggability, and user experience.

## Table of Contents

- [Principles](#principles)
- [Error Response Formats](#error-response-formats)
- [Python Backend (FastAPI)](#python-backend-fastapi)
- [Go Backend (Gin)](#go-backend-gin)
- [Frontend Error Handling](#frontend-error-handling)
- [Reliability Patterns](#reliability-patterns)
- [Testing Error Scenarios](#testing-error-scenarios)

## Principles

1. **Fail Fast**: Detect and report errors as early as possible
2. **Context Preservation**: Include relevant context in error messages
3. **Logging**: Log errors with appropriate severity levels
4. **User-Friendly Messages**: Return clear, actionable error messages to clients
5. **Consistent Format**: Use consistent error response formats across all endpoints
6. **Graceful Degradation**: Handle failures gracefully without cascading to other components

## Error Response Formats

### Standard Error Response

All APIs should return errors in a consistent JSON format:

```json
{
  "error": "Brief error description",
  "detail": "More detailed explanation (optional)",
  "code": "ERROR_CODE (optional)",
  "timestamp": "2025-11-18T05:00:00Z (optional)"
}
```

### HTTP Status Codes

Use appropriate HTTP status codes:

| Code | Meaning | Use Case |
|------|---------|----------|
| 400 | Bad Request | Invalid input, validation errors |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource already exists or conflicts |
| 422 | Unprocessable Entity | Semantic validation errors |
| 500 | Internal Server Error | Unexpected server errors |
| 502 | Bad Gateway | External service failure |
| 503 | Service Unavailable | Temporary service unavailability |

## Python Backend (FastAPI)

### Exception Handling

Use FastAPI's `HTTPException` for client errors:

```python
from fastapi import HTTPException

# Bad request - validation error
if not is_valid_uuid(device_id):
    raise HTTPException(
        status_code=400,
        detail=f"deviceid must be a valid UUID format: {device_id}"
    )

# Resource not found
if device is None:
    raise HTTPException(
        status_code=404,
        detail=f"Device not found: {device_id}"
    )
```

### Database Error Handling

Catch and handle database-specific errors:

```python
from sqlalchemy.exc import IntegrityError, OperationalError

try:
    await db.commit()
except IntegrityError as e:
    await db.rollback()
    logger.error(f"Integrity error: {e}")
    raise HTTPException(
        status_code=409,
        detail="Resource already exists or constraint violation"
    )
except OperationalError as e:
    await db.rollback()
    logger.error(f"Database operation failed: {e}")
    raise HTTPException(
        status_code=503,
        detail="Database temporarily unavailable"
    )
```

### Async Error Handling

Handle async operation errors properly:

```python
try:
    result = await async_operation()
except asyncio.TimeoutError:
    logger.error("Operation timed out")
    raise HTTPException(
        status_code=504,
        detail="Operation timed out"
    )
except Exception as e:
    logger.error(f"Unexpected error: {e}", exc_info=True)
    raise HTTPException(
        status_code=500,
        detail="Internal server error"
    )
```

### Input Validation

Validate inputs early and provide clear error messages:

```python
from pydantic import ValidationError

# Legacy field rejection
if "id" in request_data:
    raise HTTPException(
        status_code=400,
        detail="unsupported legacy field: id; use deviceid"
    )

# Required field validation
if not request_data.get("deviceid"):
    raise HTTPException(
        status_code=400,
        detail="missing required field: deviceid"
    )
```

### Logging Errors

Use appropriate log levels:

```python
import logging

logger = logging.getLogger(__name__)

# Error that requires attention
logger.error(f"Failed to process request: {error}", exc_info=True)

# Warning for recoverable issues
logger.warning(f"Retrying operation after failure: {error}")

# Info for expected errors
logger.info(f"Device not found: {device_id}")
```

## Go Backend (Gin)

### Error Responses

Use Gin's JSON response helpers:

```go
// Bad request
c.JSON(http.StatusBadRequest, gin.H{
    "error": "invalid request body: " + err.Error(),
})

// Not found
c.JSON(http.StatusNotFound, gin.H{
    "error": "Device not found",
})

// Internal server error
c.JSON(http.StatusInternalServerError, gin.H{
    "error": "database operation failed: " + err.Error(),
})
```

### Database Error Handling

Handle GORM errors appropriately:

```go
import (
    "gorm.io/gorm"
)

// Check for not found
if err := db.First(&device, deviceID).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
        return
    }
    log.Printf("Database error: %v", err)
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
    return
}
```

### Input Validation

Validate inputs and handle binding errors:

```go
var device models.Device
if err := c.BindJSON(&device); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "invalid request body: " + err.Error(),
    })
    return
}

// Validate UUID
if device.DeviceID == uuid.Nil {
    device.DeviceID = uuid.New()
}
```

### Error Wrapping

Use error wrapping for context:

```go
import "fmt"

if err := someOperation(); err != nil {
    return fmt.Errorf("failed to process device %s: %w", deviceID, err)
}
```

### Logging

Use structured logging:

```go
import "log"

// Error logging
log.Printf("ERROR: Failed to register device %s: %v", deviceID, err)

// Info logging
log.Printf("INFO: Device registered: %s", deviceID)

// Debug logging
log.Printf("DEBUG: Processing metrics for device %s", deviceID)
```

## Frontend Error Handling

### API Call Error Handling

Handle API errors in frontend code:

```javascript
try {
  const response = await fetch('/api/devices');
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.detail || error.error || 'Request failed');
  }
  
  return await response.json();
} catch (error) {
  console.error('API call failed:', error);
  
  // Show user-friendly error message
  showErrorNotification(error.message);
  
  // Report to error tracking service (if available)
  reportError(error);
}
```

### Error Boundaries (React)

Use error boundaries for component-level error handling:

```jsx
class ErrorBoundary extends React.Component {
  state = { hasError: false, error: null };

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} />;
    }
    return this.props.children;
  }
}
```

## Reliability Patterns

### Circuit Breaker

Use circuit breaker for external service calls:

```python
from app.core.reliability.circuit_breaker import CircuitBreaker

breaker = CircuitBreaker("mentor-api")

try:
    result = await breaker.call(call_mentor_api)
except CircuitBreakerError:
    logger.warning("Circuit breaker open, skipping call")
    # Fallback behavior
```

### Retry with Exponential Backoff

Use retry logic for transient failures:

```python
from app.core.reliability.retry import (
    retry_async_with_backoff,
    external_service_retry_config
)

result = await retry_async_with_backoff(
    config=external_service_retry_config(),
    operation=lambda: call_external_api(),
    operation_name="mentor-api-call"
)
```

### Timeout Handling

Set appropriate timeouts:

```python
import asyncio

try:
    result = await asyncio.wait_for(
        async_operation(),
        timeout=5.0  # 5 second timeout
    )
except asyncio.TimeoutError:
    logger.error("Operation timed out")
    raise HTTPException(
        status_code=504,
        detail="Operation timed out"
    )
```

## Testing Error Scenarios

### Unit Tests

Test error conditions explicitly:

```python
def test_device_not_found():
    """Test 404 error when device doesn't exist."""
    response = client.get(f"/api/v1/devices/{uuid.uuid4()}")
    assert response.status_code == 404
    assert "not found" in response.json()["detail"].lower()

def test_invalid_uuid_format():
    """Test 400 error with invalid UUID."""
    response = client.get("/api/v1/devices/not-a-uuid")
    assert response.status_code in [400, 422]
```

### Integration Tests

Test error propagation across services:

```python
def test_database_failure_handling():
    """Test graceful handling of database failures."""
    # Simulate database unavailability
    with patch_database_connection_failure():
        response = client.post("/api/v1/devices/register", json=device_data)
        assert response.status_code in [500, 503]
```

### Error Recovery Tests

Test recovery from error states:

```python
def test_circuit_breaker_recovery():
    """Test that circuit breaker recovers after service is back."""
    # Cause failures to open circuit
    for _ in range(5):
        try_operation_that_fails()
    
    # Wait for half-open state
    time.sleep(30)
    
    # Verify circuit recovers with successful calls
    assert try_successful_operation() == "success"
```

## Best Practices

### Do's

✅ **Always validate input before processing**
✅ **Log errors with context (request IDs, user IDs, timestamps)**
✅ **Use structured error responses**
✅ **Return appropriate HTTP status codes**
✅ **Test error scenarios explicitly**
✅ **Document expected errors in API specifications**
✅ **Use circuit breakers for external dependencies**
✅ **Implement retry logic for transient failures**

### Don'ts

❌ **Don't expose internal implementation details in error messages**
❌ **Don't return stack traces to clients in production**
❌ **Don't use generic "Something went wrong" messages**
❌ **Don't swallow errors silently**
❌ **Don't retry non-idempotent operations without careful consideration**
❌ **Don't forget to rollback database transactions on errors**

## Error Monitoring

### Metrics to Track

- Error rate by endpoint
- Error types and frequencies
- Response time for error cases
- Circuit breaker state transitions
- Retry attempt distributions

### Alerting

Set up alerts for:
- High error rates (>1% of requests)
- Repeated database connection failures
- Circuit breaker opening frequently
- Critical path failures

## See Also

- [Testing Guide](TESTING.md) - Comprehensive testing documentation
- [API Documentation](API.md) - API endpoint specifications
- [Development Guide](DEVELOPMENT.md) - Development setup and practices
- [Architecture](ARCHITECTURE.md) - System architecture overview

## Version History

- **2025-11-18**: Initial version - Error handling guidelines and best practices
