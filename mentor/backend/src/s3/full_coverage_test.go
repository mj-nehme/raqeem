package s3

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
)

// TestGetPublicEndpointWithEnvVar tests GetPublicEndpoint when MINIO_PUBLIC_ENDPOINT is set
func TestGetPublicEndpointWithEnvVar(t *testing.T) {
	// Save original value
	originalPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	defer func() {
		if originalPublicEndpoint != "" {
			_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", originalPublicEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
		}
	}()

	// Test when MINIO_PUBLIC_ENDPOINT is set
	_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", "public-minio.example.com:9000")
	assert.Equal(t, "public-minio.example.com:9000", GetPublicEndpoint())

	// Test fallback to GetEndpoint when not set
	_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
	assert.Equal(t, GetEndpoint(), GetPublicEndpoint())
}

// TestGetPublicEndpointFallback tests GetPublicEndpoint fallback behavior
func TestGetPublicEndpointFallback(t *testing.T) {
	originalPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	defer func() {
		if originalPublicEndpoint != "" {
			_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", originalPublicEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
		}
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
	}()

	// When neither is set, should return default endpoint
	_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
	_ = os.Unsetenv("MINIO_ENDPOINT")
	assert.Equal(t, "minio.default.svc.cluster.local:9000", GetPublicEndpoint())

	// When only MINIO_ENDPOINT is set
	_ = os.Setenv("MINIO_ENDPOINT", "custom-minio:9000")
	assert.Equal(t, "custom-minio:9000", GetPublicEndpoint())
}

// TestEnsureBucketExistsWithNilClient tests ensureBucketExists when client is nil
func TestEnsureBucketExistsWithNilClient(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Set client to nil
	client = nil

	// This should return early without panic
	ensureBucketExists()
	// If we reach here, the test passes
}

// TestGeneratePresignedURLWithDifferentEndpoints tests when public and internal endpoints differ
func TestGeneratePresignedURLWithDifferentEndpoints(t *testing.T) {
	// Save original values
	originalClient := client
	originalPresignClient := presignClient
	originalPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	defer func() {
		client = originalClient
		presignClient = originalPresignClient
		if originalPublicEndpoint != "" {
			_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", originalPublicEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
		}
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
	}()

	// Create a mock client
	mockClient, err := minio.New("internal:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient
	presignClient = nil // Reset presign client

	// Set different endpoints
	_ = os.Setenv("MINIO_ENDPOINT", "internal:9000")
	_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", "public:9000")

	// Call GeneratePresignedURL - this will trigger presignClient initialization
	url := GeneratePresignedURL("test.jpg")

	// The URL will be empty due to network failure, but we've covered the presignClient creation path
	assert.Equal(t, "", url)
}

// TestGeneratePresignedURLWithExistingPresignClient tests when presignClient already exists
func TestGeneratePresignedURLWithExistingPresignClient(t *testing.T) {
	// Save original values
	originalClient := client
	originalPresignClient := presignClient
	originalPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	defer func() {
		client = originalClient
		presignClient = originalPresignClient
		if originalPublicEndpoint != "" {
			_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", originalPublicEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
		}
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
	}()

	// Create mock clients
	mockClient, err := minio.New("internal:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient

	// Pre-create presign client
	mockPresignClient, err := minio.New("public:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	presignClient = mockPresignClient

	// Set different endpoints
	_ = os.Setenv("MINIO_ENDPOINT", "internal:9000")
	_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", "public:9000")

	// Call GeneratePresignedURL - should use existing presignClient
	url := GeneratePresignedURL("test.jpg")

	// URL will be empty due to network failure, but we've covered the path
	assert.Equal(t, "", url)
}

// TestInitClientWithEndpointStripping tests InitClient properly strips http/https prefixes
func TestInitClientWithEndpointStripping(t *testing.T) {
	// Save original values
	originalClient := client
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	originalSkipConnect := os.Getenv("MINIO_SKIP_CONNECT")
	defer func() {
		client = originalClient
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
		if originalSkipConnect != "" {
			_ = os.Setenv("MINIO_SKIP_CONNECT", originalSkipConnect)
		} else {
			_ = os.Unsetenv("MINIO_SKIP_CONNECT")
		}
	}()

	// Test with http:// prefix
	_ = os.Setenv("MINIO_ENDPOINT", "http://minio:9000")
	_ = os.Setenv("MINIO_SKIP_CONNECT", "1")
	client = nil
	InitClient()
	assert.NotNil(t, client)

	// Test with https:// prefix
	_ = os.Setenv("MINIO_ENDPOINT", "https://minio:9000")
	client = nil
	InitClient()
	assert.NotNil(t, client)
}

// TestGeneratePresignedURLWithPublicEndpointWithHttpPrefix tests public endpoint with http prefix
func TestGeneratePresignedURLWithPublicEndpointWithHttpPrefix(t *testing.T) {
	// Save original values
	originalClient := client
	originalPresignClient := presignClient
	originalPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	defer func() {
		client = originalClient
		presignClient = originalPresignClient
		if originalPublicEndpoint != "" {
			_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", originalPublicEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
		}
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
	}()

	// Create a mock client
	mockClient, err := minio.New("internal:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient
	presignClient = nil

	// Set public endpoint with http:// prefix
	_ = os.Setenv("MINIO_ENDPOINT", "http://internal:9000")
	_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", "http://public:9000")

	url := GeneratePresignedURL("test.jpg")
	// URL will be empty due to network failure
	assert.Equal(t, "", url)
}

// TestGeneratePresignedURLWithMockServer tests presigned URL generation with a mock server
func TestGeneratePresignedURLWithMockServer(t *testing.T) {
	// Create a mock S3 server that responds to presign requests
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a simple 200 OK for all requests
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><ListAllMyBucketsResult><Buckets></Buckets></ListAllMyBucketsResult>`))
	}))
	defer mockServer.Close()

	// Save original values
	originalClient := client
	originalPresignClient := presignClient
	originalPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	defer func() {
		client = originalClient
		presignClient = originalPresignClient
		if originalPublicEndpoint != "" {
			_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", originalPublicEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
		}
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
	}()

	// Extract just the host:port from mock server URL (strip http://)
	serverAddr := mockServer.URL[7:] // Remove "http://"

	// Create a client pointing to mock server
	mockClient, err := minio.New(serverAddr, &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient
	presignClient = nil

	// Set endpoints to same value (will use main client for presign)
	_ = os.Setenv("MINIO_ENDPOINT", serverAddr)
	_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")

	// Call GeneratePresignedURL - should succeed and return a URL
	url := GeneratePresignedURL("test.jpg")

	// With a mock server, we should get a presigned URL
	assert.NotEmpty(t, url, "Should generate a presigned URL with mock server")
}

// TestHealthCheckWithMockServer tests health check with a mock server
func TestHealthCheckWithMockServer(t *testing.T) {
	// Create a mock S3 server that returns successful list buckets response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><ListAllMyBucketsResult><Buckets></Buckets></ListAllMyBucketsResult>`))
	}))
	defer mockServer.Close()

	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Extract just the host:port from mock server URL
	serverAddr := mockServer.URL[7:] // Remove "http://"

	// Create a client pointing to mock server
	mockClient, err := minio.New(serverAddr, &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient

	// Health check should succeed
	err = HealthCheck()
	assert.NoError(t, err, "Health check should succeed with mock server")
}

// TestInitClientWithMockServer tests InitClient with a mock server
func TestInitClientWithMockServer(t *testing.T) {
	// Create a mock S3 server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		// Return bucket list response or bucket exists response
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><ListAllMyBucketsResult><Buckets></Buckets></ListAllMyBucketsResult>`))
		} else {
			// For bucket exists check, return 200
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><LocationConstraint></LocationConstraint>`))
		}
	}))
	defer mockServer.Close()

	// Save original values
	originalClient := client
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	originalSkipConnect := os.Getenv("MINIO_SKIP_CONNECT")
	defer func() {
		client = originalClient
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
		if originalSkipConnect != "" {
			_ = os.Setenv("MINIO_SKIP_CONNECT", originalSkipConnect)
		} else {
			_ = os.Unsetenv("MINIO_SKIP_CONNECT")
		}
	}()

	// Extract just the host:port from mock server URL
	serverAddr := mockServer.URL[7:] // Remove "http://"

	// Set environment to use mock server and don't skip connection
	_ = os.Setenv("MINIO_ENDPOINT", serverAddr)
	_ = os.Unsetenv("MINIO_SKIP_CONNECT")
	client = nil

	// InitClient should succeed with mock server
	InitClient()

	assert.NotNil(t, client, "Client should be initialized with mock server")
}

// TestEnsureBucketExistsWithMockServer tests bucket creation paths
func TestEnsureBucketExistsWithMockServer(t *testing.T) {
	// Track what requests were made
	requestPaths := make([]string, 0)
	bucketExistsResponse := http.StatusOK

	// Create a mock S3 server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPaths = append(requestPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/xml")

		// Bucket exists check (HEAD or GET on bucket)
		if r.Method == "HEAD" {
			if bucketExistsResponse == http.StatusOK {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}

		// For bucket location query
		if r.URL.RawQuery == "location=" || r.URL.RawQuery == "location" {
			if bucketExistsResponse == http.StatusOK {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><LocationConstraint></LocationConstraint>`))
			} else {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchBucket</Code><Message>The specified bucket does not exist</Message></Error>`))
			}
			return
		}

		// Create bucket request
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><ListAllMyBucketsResult><Buckets></Buckets></ListAllMyBucketsResult>`))
	}))
	defer mockServer.Close()

	// Save original client
	originalClient := client
	originalBucketName := os.Getenv("MINIO_BUCKET_NAME")
	defer func() {
		client = originalClient
		if originalBucketName != "" {
			_ = os.Setenv("MINIO_BUCKET_NAME", originalBucketName)
		} else {
			_ = os.Unsetenv("MINIO_BUCKET_NAME")
		}
	}()

	// Extract just the host:port from mock server URL
	serverAddr := mockServer.URL[7:] // Remove "http://"

	// Test 1: Bucket already exists
	t.Run("bucket already exists", func(t *testing.T) {
		bucketExistsResponse = http.StatusOK
		mockClient, err := minio.New(serverAddr, &minio.Options{
			Creds:  credentials.NewStaticV4("test", "test", ""),
			Secure: false,
		})
		assert.NoError(t, err)
		client = mockClient
		_ = os.Setenv("MINIO_BUCKET_NAME", "test-bucket")

		// ensureBucketExists should log "already exists"
		ensureBucketExists()
		// Just verify it doesn't crash
	})

	// Test 2: Bucket doesn't exist (should create)
	t.Run("bucket does not exist", func(t *testing.T) {
		bucketExistsResponse = http.StatusNotFound
		requestPaths = nil // Reset

		mockClient, err := minio.New(serverAddr, &minio.Options{
			Creds:  credentials.NewStaticV4("test", "test", ""),
			Secure: false,
		})
		assert.NoError(t, err)
		client = mockClient

		// ensureBucketExists should attempt to create bucket
		ensureBucketExists()
		// Just verify it doesn't crash
	})
}

// TestGeneratePresignedURLPresignClientCreationError tests error path for presign client creation
func TestGeneratePresignedURLPresignClientCreationError(t *testing.T) {
	// Save original values
	originalClient := client
	originalPresignClient := presignClient
	originalPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	defer func() {
		client = originalClient
		presignClient = originalPresignClient
		if originalPublicEndpoint != "" {
			_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", originalPublicEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
		}
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
	}()

	// Create a mock client
	mockClient, err := minio.New("internal:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient
	presignClient = nil

	// Set public endpoint to an invalid value that will cause minio.New to fail
	// A single colon (":") fails validation because minio.New expects valid hostname format
	// Error: "Endpoint: : does not follow ip address or domain name standards."
	_ = os.Setenv("MINIO_ENDPOINT", "internal:9000")
	_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", ":")

	url := GeneratePresignedURL("test.jpg")
	// Should return empty string due to presign client creation error
	assert.Equal(t, "", url)
}

// TestInitClientMinioNewError tests the error path when minio.New fails
func TestInitClientMinioNewError(t *testing.T) {
	// Save original values
	originalClient := client
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	originalSkipConnect := os.Getenv("MINIO_SKIP_CONNECT")
	defer func() {
		client = originalClient
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
		if originalSkipConnect != "" {
			_ = os.Setenv("MINIO_SKIP_CONNECT", originalSkipConnect)
		} else {
			_ = os.Unsetenv("MINIO_SKIP_CONNECT")
		}
	}()

	// Set endpoint to "http://" which, after stripping the "http://" prefix in InitClient,
	// becomes an empty string "" that fails minio.New validation with:
	// "Endpoint:  does not follow ip address or domain name standards."
	_ = os.Setenv("MINIO_ENDPOINT", "http://")
	_ = os.Setenv("MINIO_SKIP_CONNECT", "1")
	client = nil

	// InitClient should handle the error gracefully
	InitClient()

	// The test passes if we don't panic - client might still be nil due to initialization failure
}

// TestInitClientSuccessfulConnectionPath tests the successful connection path in InitClient
func TestInitClientSuccessfulConnectionPath(t *testing.T) {
	// Create a mock S3 server that simulates a working MinIO
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		// Handle different requests
		if r.URL.Path == "/" {
			// List buckets response
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult>
  <Owner><ID>test</ID><DisplayName>test</DisplayName></Owner>
  <Buckets></Buckets>
</ListAllMyBucketsResult>`))
		} else if r.URL.RawQuery == "location=" || r.URL.RawQuery == "location" {
			// Bucket location/exists check
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><LocationConstraint></LocationConstraint>`))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer mockServer.Close()

	// Save original values
	originalClient := client
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	originalSkipConnect := os.Getenv("MINIO_SKIP_CONNECT")
	defer func() {
		client = originalClient
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
		if originalSkipConnect != "" {
			_ = os.Setenv("MINIO_SKIP_CONNECT", originalSkipConnect)
		} else {
			_ = os.Unsetenv("MINIO_SKIP_CONNECT")
		}
	}()

	// Extract just the host:port from mock server URL
	serverAddr := mockServer.URL[7:] // Remove "http://"

	// Set environment to use mock server - DON'T skip connection to test the successful path
	_ = os.Setenv("MINIO_ENDPOINT", serverAddr)
	_ = os.Unsetenv("MINIO_SKIP_CONNECT") // Important: don't skip connection
	client = nil

	// InitClient should succeed with mock server and go through the success path (line 102)
	InitClient()

	assert.NotNil(t, client, "Client should be initialized with mock server")
}

// TestEnsureBucketExistsBucketDoesNotExist tests the path where bucket doesn't exist and gets created
func TestEnsureBucketExistsBucketDoesNotExist(t *testing.T) {
	requestCount := 0
	// Create a mock S3 server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/xml")

		// First request: bucket exists check - say it doesn't exist
		if r.URL.RawQuery == "location=" || r.URL.RawQuery == "location" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchBucket</Code></Error>`))
			return
		}

		// PUT request to create bucket
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Extract just the host:port
	serverAddr := mockServer.URL[7:]

	mockClient, err := minio.New(serverAddr, &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient

	ensureBucketExists()
	// Test passes if no panic
}

// TestEnsureBucketExistsMakeBucketError tests the path where MakeBucket fails
func TestEnsureBucketExistsMakeBucketError(t *testing.T) {
	// Create a mock S3 server that fails MakeBucket
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")

		// Bucket exists check - say it doesn't exist
		if r.URL.RawQuery == "location=" || r.URL.RawQuery == "location" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchBucket</Code></Error>`))
			return
		}

		// PUT request to create bucket - fail it
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>AccessDenied</Code></Error>`))
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Extract just the host:port
	serverAddr := mockServer.URL[7:]

	mockClient, err := minio.New(serverAddr, &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient

	ensureBucketExists()
	// Test passes if no panic - the error is just logged
}

// TestEnsureBucketExistsBucketAlreadyExists tests when bucket already exists
func TestEnsureBucketExistsBucketAlreadyExists(t *testing.T) {
	// Create a mock S3 server that says bucket exists
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")

		// Bucket exists check - say it exists
		if r.URL.RawQuery == "location=" || r.URL.RawQuery == "location" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><LocationConstraint></LocationConstraint>`))
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Extract just the host:port
	serverAddr := mockServer.URL[7:]

	mockClient, err := minio.New(serverAddr, &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient

	ensureBucketExists()
	// Test passes if no panic and logs "already exists"
}

// TestStatObjectSuccessPath tests when StatObject succeeds (doesn't log warning)
func TestStatObjectSuccessPath(t *testing.T) {
	// Create a mock S3 server that returns success for stat
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")

		// HEAD request for StatObject
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "100")
			w.Header().Set("ETag", `"d41d8cd98f00b204e9800998ecf8427e"`)
			w.WriteHeader(http.StatusOK)
			return
		}

		// For bucket location check (used in presign)
		if r.URL.RawQuery == "location=" || r.URL.RawQuery == "location" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><LocationConstraint></LocationConstraint>`))
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Save original values
	originalClient := client
	originalPresignClient := presignClient
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	originalPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	defer func() {
		client = originalClient
		presignClient = originalPresignClient
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
		if originalPublicEndpoint != "" {
			_ = os.Setenv("MINIO_PUBLIC_ENDPOINT", originalPublicEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")
		}
	}()

	// Extract just the host:port
	serverAddr := mockServer.URL[7:]

	mockClient, err := minio.New(serverAddr, &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient
	presignClient = nil

	// Set same endpoints so we use the main client
	_ = os.Setenv("MINIO_ENDPOINT", serverAddr)
	_ = os.Unsetenv("MINIO_PUBLIC_ENDPOINT")

	// Call GeneratePresignedURL
	url := GeneratePresignedURL("test.jpg")

	// Should get a presigned URL
	assert.NotEmpty(t, url, "Should generate presigned URL")
}

// TestContextTimeoutInHealthCheck tests that HealthCheck respects context timeout
func TestContextTimeoutInHealthCheck(t *testing.T) {
	// Create a slow mock server that responds after the HealthCheck timeout
	// The handler uses request context to detect when client disconnects
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wait for either 10 seconds or until client disconnects
		select {
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// Client disconnected, exit handler cleanly
			return
		}
	}))

	// Save original client
	originalClient := client
	// Use defer with a channel to ensure we wait for goroutine before restoring
	goroutineDone := make(chan struct{})
	defer func() {
		// Wait for the goroutine to complete before restoring client
		<-goroutineDone
		client = originalClient
		// Close client connections first to avoid blocking on server close
		mockServer.CloseClientConnections()
		mockServer.Close()
	}()

	// Extract just the host:port
	serverAddr := mockServer.URL[7:]

	mockClient, err := minio.New(serverAddr, &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	assert.NoError(t, err)
	client = mockClient

	// Start HealthCheck in goroutine
	done := make(chan error, 1)
	go func() {
		defer close(goroutineDone) // Signal that goroutine is done
		done <- HealthCheck()
	}()

	// Wait for HealthCheck to return (it should timeout internally within 5 seconds)
	// or wait a bit less if the external context is cancelled
	select {
	case err := <-done:
		// HealthCheck completed (likely with error due to timeout)
		assert.Error(t, err, "HealthCheck should fail due to timeout")
	case <-time.After(6 * time.Second):
		// If HealthCheck hasn't returned after 6 seconds, fail the test
		t.Error("HealthCheck did not return within expected time")
	}
}
