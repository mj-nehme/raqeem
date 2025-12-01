package util

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPClientWithRetry(t *testing.T) {
	client := NewHTTPClientWithRetry(5*time.Second, 3)
	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}
	if client.maxRetries != 3 {
		t.Errorf("Expected maxRetries to be 3, got %d", client.maxRetries)
	}
	if client.baseDelay != 100*time.Millisecond {
		t.Errorf("Expected baseDelay to be 100ms, got %v", client.baseDelay)
	}
	if client.maxDelay != 5*time.Second {
		t.Errorf("Expected maxDelay to be 5s, got %v", client.maxDelay)
	}
}

func TestHTTPClientWithRetrySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewHTTPClientWithRetry(5*time.Second, 3)
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClientWithRetryServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPClientWithRetry(5*time.Second, 2)
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	// After all retries fail, we expect an error
	if err == nil {
		t.Fatalf("Expected error after all retries, got nil")
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	// Should retry twice (maxRetries=2) plus initial attempt = 3 total
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestHTTPClientWithRetryPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "test data" {
			t.Errorf("Expected 'test data', got '%s'", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClientWithRetry(5*time.Second, 3)
	resp, err := client.Post(server.URL, "text/plain", strings.NewReader("test data"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClientWithRetryRateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewHTTPClientWithRetry(5*time.Second, 3)
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 after retry, got %d", resp.StatusCode)
	}
	if attempts < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", attempts)
	}
}

func TestHTTPClientWithRetryClientError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewHTTPClientWithRetry(5*time.Second, 3)
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Should not retry on 4xx errors (except 429)
	if attempts != 1 {
		t.Errorf("Expected 1 attempt for 4xx error, got %d", attempts)
	}
}

func TestHTTPClientWithRetryNetworkError(t *testing.T) {
	// Create a server and immediately close it to simulate network error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := server.URL
	server.Close()

	client := NewHTTPClientWithRetry(100*time.Millisecond, 1)
	req, err := http.NewRequest("GET", serverURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = client.Do(req)
	// Expect error after retries exhausted
	if err == nil {
		t.Fatal("Expected error for network failure, got nil")
	}
}

func TestHTTPClientWithRetryMaxDelayCapWithManyRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Fail for many attempts to ensure delay reaches maxDelay
		if attempts < 6 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create client with short delays and many retries to test delay cap
	client := &HTTPClientWithRetry{
		client:     &http.Client{Timeout: 5 * time.Second},
		maxRetries: 6,
		baseDelay:  1 * time.Millisecond,
		maxDelay:   10 * time.Millisecond, // Very short max delay
		shouldRetry: func(resp *http.Response, err error) bool {
			if err != nil {
				return true
			}
			if resp != nil && (resp.StatusCode >= 500 || resp.StatusCode == 429) {
				return true
			}
			return false
		},
	}

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no error after successful retry, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if attempts < 6 {
		t.Errorf("Expected at least 6 attempts, got %d", attempts)
	}
}

func TestHTTPClientWithRetryPostInvalidURL(t *testing.T) {
	client := NewHTTPClientWithRetry(5*time.Second, 3)
	// Use an invalid URL that will cause http.NewRequest to fail
	_, err := client.Post("://invalid-url", "text/plain", strings.NewReader("test"))
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}
}

// errorReader is a reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestHTTPClientWithRetryBodyReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClientWithRetry(5*time.Second, 3)
	req, err := http.NewRequest("POST", server.URL, &errorReader{})
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = client.Do(req)
	if err == nil {
		t.Fatal("Expected error when reading body fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read request body") {
		t.Errorf("Expected 'failed to read request body' error, got %v", err)
	}
}

// errorCloser is a ReadCloser that returns an error when closed
type errorCloser struct {
	io.Reader
}

func (e *errorCloser) Close() error {
	return io.ErrClosedPipe
}

func TestHTTPClientWithRetryRequestBodyCloseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClientWithRetry(5*time.Second, 3)
	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	// Assign body that will fail on close
	req.Body = &errorCloser{Reader: strings.NewReader("test data")}

	resp, err := client.Do(req)
	// The request should still succeed, even if close fails (only logs error)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// mockTransport is a custom RoundTripper that returns a mocked response
type mockTransport struct {
	statusCode int
	body       io.ReadCloser
	err        error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       m.body,
		Header:     make(http.Header),
	}, nil
}

func TestHTTPClientWithRetryResponseBodyCloseError(t *testing.T) {
	// Create a client with a mock transport that returns a 500 error with a body that fails to close
	client := &HTTPClientWithRetry{
		client: &http.Client{
			Transport: &mockTransport{
				statusCode: http.StatusInternalServerError,
				body:       &errorCloser{Reader: strings.NewReader("error response")},
			},
			Timeout: 5 * time.Second,
		},
		maxRetries: 1,
		baseDelay:  1 * time.Millisecond,
		maxDelay:   10 * time.Millisecond,
		shouldRetry: func(resp *http.Response, err error) bool {
			if err != nil {
				return true
			}
			if resp != nil && (resp.StatusCode >= 500 || resp.StatusCode == 429) {
				return true
			}
			return false
		},
	}

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// This should trigger the response body close error logging path
	_, err = client.Do(req)
	// We expect an error because all retries failed
	if err == nil {
		t.Fatal("Expected error after retries exhausted, got nil")
	}
}
