package s3

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	t.Run("returns error when client is nil", func(t *testing.T) {
		// Save and clear client
		originalClient := client
		client = nil
		defer func() { client = originalClient }()

		err := HealthCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MinIO client is not initialized")
	})

	t.Run("returns error when client cannot connect", func(t *testing.T) {
		// Save original client
		originalClient := client
		defer func() { client = originalClient }()

		// Create a client with invalid endpoint
		invalidClient, err := minio.New("invalid-endpoint:9000", &minio.Options{
			Creds:  credentials.NewStaticV4("test", "test", ""),
			Secure: false,
		})
		assert.NoError(t, err)
		client = invalidClient

		// Set a very short timeout for the test
		err = HealthCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MinIO health check failed")
	})

	t.Run("handles context timeout", func(t *testing.T) {
		// Save original client
		originalClient := client
		defer func() { client = originalClient }()

		// Create a client that will timeout
		timeoutClient, err := minio.New("192.0.2.1:9000", &minio.Options{ // Using TEST-NET-1 IP that won't respond
			Creds:  credentials.NewStaticV4("test", "test", ""),
			Secure: false,
		})
		assert.NoError(t, err)
		client = timeoutClient

		// HealthCheck has a 5-second timeout built in
		start := time.Now()
		err = HealthCheck()
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MinIO health check failed")
		// Should timeout within reasonable time
		assert.Less(t, elapsed, 10*time.Second)
	})
}

func TestInitClient(t *testing.T) {
	t.Run("initializes with environment variables", func(t *testing.T) {
		// Save original values
		originalEndpoint := os.Getenv("MINIO_ENDPOINT")
		originalAccessKey := os.Getenv("MINIO_ACCESS_KEY")
		originalSecretKey := os.Getenv("MINIO_SECRET_KEY")
		originalSkipConnect := os.Getenv("MINIO_SKIP_CONNECT")
		originalClient := client

		// Set test environment variables
		os.Setenv("MINIO_ENDPOINT", "test-endpoint:9000")
		os.Setenv("MINIO_ACCESS_KEY", "test-access")
		os.Setenv("MINIO_SECRET_KEY", "test-secret")
		os.Setenv("MINIO_SKIP_CONNECT", "1") // Skip connectivity check

		// Restore after test
		defer func() {
			if originalEndpoint != "" {
				os.Setenv("MINIO_ENDPOINT", originalEndpoint)
			} else {
				os.Unsetenv("MINIO_ENDPOINT")
			}
			if originalAccessKey != "" {
				os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
			} else {
				os.Unsetenv("MINIO_ACCESS_KEY")
			}
			if originalSecretKey != "" {
				os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
			} else {
				os.Unsetenv("MINIO_SECRET_KEY")
			}
			if originalSkipConnect != "" {
				os.Setenv("MINIO_SKIP_CONNECT", originalSkipConnect)
			} else {
				os.Unsetenv("MINIO_SKIP_CONNECT")
			}
			client = originalClient
		}()

		// Initialize client
		InitClient()

		// Verify client was created
		assert.NotNil(t, client)
	})

	t.Run("handles initialization failure gracefully", func(t *testing.T) {
		// Save original values
		originalEndpoint := os.Getenv("MINIO_ENDPOINT")
		originalAccessKey := os.Getenv("MINIO_ACCESS_KEY")
		originalSecretKey := os.Getenv("MINIO_SECRET_KEY")
		originalSkipConnect := os.Getenv("MINIO_SKIP_CONNECT")
		originalClient := client

		// Set invalid endpoint (will fail to connect but not crash)
		os.Setenv("MINIO_ENDPOINT", "invalid-endpoint-that-does-not-exist:9000")
		os.Setenv("MINIO_ACCESS_KEY", "test")
		os.Setenv("MINIO_SECRET_KEY", "test")
		os.Unsetenv("MINIO_SKIP_CONNECT") // Don't skip connect to test failure path

		// Restore after test
		defer func() {
			if originalEndpoint != "" {
				os.Setenv("MINIO_ENDPOINT", originalEndpoint)
			} else {
				os.Unsetenv("MINIO_ENDPOINT")
			}
			if originalAccessKey != "" {
				os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
			} else {
				os.Unsetenv("MINIO_ACCESS_KEY")
			}
			if originalSecretKey != "" {
				os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
			} else {
				os.Unsetenv("MINIO_SECRET_KEY")
			}
			if originalSkipConnect != "" {
				os.Setenv("MINIO_SKIP_CONNECT", originalSkipConnect)
			} else {
				os.Unsetenv("MINIO_SKIP_CONNECT")
			}
			client = originalClient
		}()

		// This should not panic even if connection fails
		InitClient()
		// Client may or may not be nil depending on initialization success
		// The important thing is that it doesn't crash
	})

	t.Run("skips connection check when MINIO_SKIP_CONNECT is set", func(t *testing.T) {
		// Save original values
		originalEndpoint := os.Getenv("MINIO_ENDPOINT")
		originalAccessKey := os.Getenv("MINIO_ACCESS_KEY")
		originalSecretKey := os.Getenv("MINIO_SECRET_KEY")
		originalSkipConnect := os.Getenv("MINIO_SKIP_CONNECT")
		originalClient := client

		// Set environment to skip connection
		os.Setenv("MINIO_ENDPOINT", "any-endpoint:9000")
		os.Setenv("MINIO_ACCESS_KEY", "any-key")
		os.Setenv("MINIO_SECRET_KEY", "any-secret")
		os.Setenv("MINIO_SKIP_CONNECT", "1")

		// Restore after test
		defer func() {
			if originalEndpoint != "" {
				os.Setenv("MINIO_ENDPOINT", originalEndpoint)
			} else {
				os.Unsetenv("MINIO_ENDPOINT")
			}
			if originalAccessKey != "" {
				os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
			} else {
				os.Unsetenv("MINIO_ACCESS_KEY")
			}
			if originalSecretKey != "" {
				os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
			} else {
				os.Unsetenv("MINIO_SECRET_KEY")
			}
			if originalSkipConnect != "" {
				os.Setenv("MINIO_SKIP_CONNECT", originalSkipConnect)
			} else {
				os.Unsetenv("MINIO_SKIP_CONNECT")
			}
			client = originalClient
		}()

		// This should succeed quickly without trying to connect
		start := time.Now()
		InitClient()
		elapsed := time.Since(start)

		// Should complete quickly since we're skipping the connection check
		assert.Less(t, elapsed, 2*time.Second)
		assert.NotNil(t, client)
	})
}

func TestGetFunctions(t *testing.T) {
	t.Run("GetEndpoint returns environment variable", func(t *testing.T) {
		originalEndpoint := os.Getenv("MINIO_ENDPOINT")
		defer func() {
			if originalEndpoint != "" {
				os.Setenv("MINIO_ENDPOINT", originalEndpoint)
			} else {
				os.Unsetenv("MINIO_ENDPOINT")
			}
		}()

		os.Setenv("MINIO_ENDPOINT", "custom-endpoint:9000")
		assert.Equal(t, "custom-endpoint:9000", GetEndpoint())
	})

	t.Run("GetEndpoint returns default when not set", func(t *testing.T) {
		originalEndpoint := os.Getenv("MINIO_ENDPOINT")
		os.Unsetenv("MINIO_ENDPOINT")
		defer func() {
			if originalEndpoint != "" {
				os.Setenv("MINIO_ENDPOINT", originalEndpoint)
			}
		}()

		assert.Equal(t, "minio.default.svc.cluster.local:9000", GetEndpoint())
	})

	t.Run("GetAccessKey returns environment variable", func(t *testing.T) {
		originalAccessKey := os.Getenv("MINIO_ACCESS_KEY")
		defer func() {
			if originalAccessKey != "" {
				os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
			} else {
				os.Unsetenv("MINIO_ACCESS_KEY")
			}
		}()

		os.Setenv("MINIO_ACCESS_KEY", "custom-access-key")
		assert.Equal(t, "custom-access-key", GetAccessKey())
	})

	t.Run("GetAccessKey returns default when not set", func(t *testing.T) {
		originalAccessKey := os.Getenv("MINIO_ACCESS_KEY")
		os.Unsetenv("MINIO_ACCESS_KEY")
		defer func() {
			if originalAccessKey != "" {
				os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
			}
		}()

		assert.Equal(t, "minioadmin", GetAccessKey())
	})

	t.Run("GetSecretKey returns environment variable", func(t *testing.T) {
		originalSecretKey := os.Getenv("MINIO_SECRET_KEY")
		defer func() {
			if originalSecretKey != "" {
				os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
			} else {
				os.Unsetenv("MINIO_SECRET_KEY")
			}
		}()

		os.Setenv("MINIO_SECRET_KEY", "custom-secret-key")
		assert.Equal(t, "custom-secret-key", GetSecretKey())
	})

	t.Run("GetSecretKey returns default when not set", func(t *testing.T) {
		originalSecretKey := os.Getenv("MINIO_SECRET_KEY")
		os.Unsetenv("MINIO_SECRET_KEY")
		defer func() {
			if originalSecretKey != "" {
				os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
			}
		}()

		assert.Equal(t, "minioadmin1234", GetSecretKey())
	})
}

func TestGeneratePresignedURLContext(t *testing.T) {
	t.Run("handles context properly", func(t *testing.T) {
		originalClient := client
		defer func() { client = originalClient }()

		// Test with nil client (should return empty string)
		client = nil
		url := GeneratePresignedURL("test.jpg")
		assert.Equal(t, "", url)

		// Test with empty filename
		url = GeneratePresignedURL("")
		assert.Equal(t, "", url)
	})
}

func TestHealthCheckWithContext(t *testing.T) {
	t.Run("respects context timeout", func(t *testing.T) {
		originalClient := client
		defer func() { client = originalClient }()

		// Create a client that will timeout
		timeoutClient, err := minio.New("192.0.2.1:9000", &minio.Options{
			Creds:  credentials.NewStaticV4("test", "test", ""),
			Secure: false,
		})
		assert.NoError(t, err)
		client = timeoutClient

		// Create a context with a very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Wait for context to timeout
		<-ctx.Done()

		// Now call health check which has its own timeout
		err = HealthCheck()
		assert.Error(t, err)
	})
}
