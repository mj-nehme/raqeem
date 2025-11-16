package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HTTPClientWithRetry wraps an HTTP client with retry logic for resilient external service calls.
// This client automatically retries failed requests using exponential backoff, making the
// application more resilient to transient network issues and temporary service outages.
type HTTPClientWithRetry struct {
	client      *http.Client
	maxRetries  int
	baseDelay   time.Duration
	maxDelay    time.Duration
	shouldRetry func(resp *http.Response, err error) bool
}

// NewHTTPClientWithRetry creates a new HTTP client with retry logic.
// 
// The client implements the following retry strategy:
// - Retries network errors (connection failures, timeouts)
// - Retries server errors (5xx status codes)
// - Retries rate limiting (429 Too Many Requests)
// - Does not retry client errors (4xx except 429) as they indicate invalid requests
// - Uses exponential backoff with a cap to avoid overwhelming services during recovery
//
// Parameters:
//   - timeout: Maximum time for each request attempt
//   - maxRetries: Number of retry attempts after initial failure (e.g., 3 means 4 total attempts)
func NewHTTPClientWithRetry(timeout time.Duration, maxRetries int) *HTTPClientWithRetry {
	return &HTTPClientWithRetry{
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   5 * time.Second,
		shouldRetry: func(resp *http.Response, err error) bool {
			// Retry on network errors (connection failures, timeouts, DNS errors)
			if err != nil {
				return true
			}
			// Retry on 5xx server errors and 429 rate limit
			// Client errors (4xx except 429) are not retried as they indicate invalid requests
			if resp != nil && (resp.StatusCode >= 500 || resp.StatusCode == 429) {
				return true
			}
			return false
		},
	}
}

// Do performs an HTTP request with retry logic and exponential backoff.
// 
// The request body is buffered to allow retries. If the request fails with a retryable
// error, it will be retried with increasing delays between attempts. The delay doubles
// after each attempt until it reaches maxDelay, preventing rapid retry storms.
//
// This method properly handles request body restoration for retries and cleans up
// response bodies to prevent resource leaks.
func (c *HTTPClientWithRetry) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	var bodyBytes []byte

	// Save body for retries - the request body can only be read once, so we buffer it
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		if err := req.Body.Close(); err != nil {
			log.Printf("Failed to close request body: %v", err)
		}
	}

	delay := c.baseDelay

	for attempt := 1; attempt <= c.maxRetries+1; attempt++ {
		// Restore body for retry - create a new reader from the buffered bytes
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = c.client.Do(req)

		// Success or non-retryable error - return immediately
		if !c.shouldRetry(resp, err) {
			return resp, err
		}

		// Close response body if present to prevent resource leaks
		if resp != nil && resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				log.Printf("Failed to close response body: %v", err)
			}
		}

		// Don't sleep after last attempt - fail immediately
		if attempt <= c.maxRetries {
			log.Printf("Request to %s failed (attempt %d/%d), retrying in %v: %v",
				req.URL, attempt, c.maxRetries+1, delay, err)
			time.Sleep(delay)

			// Exponential backoff: double the delay each time, up to maxDelay
			// This prevents overwhelming recovering services while allowing fast recovery
			delay = delay * 2
			if delay > c.maxDelay {
				delay = c.maxDelay
			}
		}
	}

	return resp, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, err)
}

// Post performs a POST request with retry logic.
// This is a convenience wrapper around Do() for POST requests.
func (c *HTTPClientWithRetry) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}
