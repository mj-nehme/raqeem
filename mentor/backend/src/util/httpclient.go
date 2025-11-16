package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HTTPClientWithRetry wraps an HTTP client with retry logic
type HTTPClientWithRetry struct {
	client      *http.Client
	maxRetries  int
	baseDelay   time.Duration
	maxDelay    time.Duration
	shouldRetry func(resp *http.Response, err error) bool
}

// NewHTTPClientWithRetry creates a new HTTP client with retry logic
func NewHTTPClientWithRetry(timeout time.Duration, maxRetries int) *HTTPClientWithRetry {
	return &HTTPClientWithRetry{
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   5 * time.Second,
		shouldRetry: func(resp *http.Response, err error) bool {
			// Retry on network errors
			if err != nil {
				return true
			}
			// Retry on 5xx server errors and 429 rate limit
			if resp != nil && (resp.StatusCode >= 500 || resp.StatusCode == 429) {
				return true
			}
			return false
		},
	}
}

// Do performs an HTTP request with retry logic
func (c *HTTPClientWithRetry) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	var bodyBytes []byte

	// Save body for retries
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
	}

	delay := c.baseDelay

	for attempt := 1; attempt <= c.maxRetries+1; attempt++ {
		// Restore body for retry
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = c.client.Do(req)

		// Success or non-retryable error
		if !c.shouldRetry(resp, err) {
			return resp, err
		}

		// Close response body if present
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		// Don't sleep after last attempt
		if attempt <= c.maxRetries {
			log.Printf("Request to %s failed (attempt %d/%d), retrying in %v: %v",
				req.URL, attempt, c.maxRetries+1, delay, err)
			time.Sleep(delay)

			// Exponential backoff with jitter
			delay = delay * 2
			if delay > c.maxDelay {
				delay = c.maxDelay
			}
		}
	}

	return resp, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, err)
}

// Post performs a POST request with retry logic
func (c *HTTPClientWithRetry) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}
