package s3

import (
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
)

func TestGeneratePresignedURLWithoutInit(t *testing.T) {
	// Reset client to nil to ensure clean state
	client = nil

	// Test that GeneratePresignedURL handles nil client gracefully
	url := GeneratePresignedURL("test.jpg")
	assert.Equal(t, "", url, "Should return empty string when client is not initialized")
}

func TestInitClientCreatesClient(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Reset client to nil
	client = nil

	// Call InitClient - it should create a client object
	// The minio.New() call succeeds even if the server is not reachable
	InitClient()

	// Verify that client was initialized
	assert.NotNil(t, client, "Client should be initialized after InitClient")
}

func TestGeneratePresignedURLWithMockClient(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Create a mock client (will fail to connect but that's ok for this test)
	mockClient, err := minio.New("invalid-endpoint:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})

	// Client creation should succeed even if endpoint is unreachable
	assert.NoError(t, err)
	client = mockClient

	// Try to generate URL - it will fail because endpoint is invalid
	// but we're testing the non-nil path
	url := GeneratePresignedURL("test-file.jpg")

	// Since the endpoint is invalid, it should return empty string due to error
	assert.Equal(t, "", url)
}

func TestInitClientEnvironmentVariables(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Reset client to nil
	client = nil

	// Call InitClient - it uses hardcoded values, not env vars
	InitClient()

	// Verify that client was initialized with hardcoded credentials
	assert.NotNil(t, client, "Client should be initialized")
}

func TestGeneratePresignedURLEmptyFilename(t *testing.T) {
	// Ensure client is nil
	client = nil

	// Test with empty filename
	url := GeneratePresignedURL("")
	assert.Equal(t, "", url)
}

func TestGeneratePresignedURLMultipleCalls(t *testing.T) {
	// Ensure client is nil
	client = nil

	// Test multiple calls work correctly
	url1 := GeneratePresignedURL("file1.jpg")
	url2 := GeneratePresignedURL("file2.jpg")
	url3 := GeneratePresignedURL("file3.jpg")

	// All should return empty string when client is not initialized
	assert.Equal(t, "", url1)
	assert.Equal(t, "", url2)
	assert.Equal(t, "", url3)
}

func TestInitClientMultipleCalls(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Reset client to nil
	client = nil

	// Call InitClient multiple times - should not crash
	InitClient()
	assert.NotNil(t, client, "Client should be initialized")

	// Store first client reference
	_ = client

	// Call again - should reinitialize
	InitClient()
	assert.NotNil(t, client, "Client should still be initialized")

	// Client may be the same or different, just verify it's still valid
	assert.NotNil(t, client)
}

func TestGeneratePresignedURLVariousFilenames(t *testing.T) {
	// Ensure client is nil
	client = nil

	testCases := []struct {
		name     string
		filename string
		expected string
	}{
		{"normal file", "screenshot.jpg", ""},
		{"file with path", "folder/screenshot.jpg", ""},
		{"file with spaces", "my file.jpg", ""},
		{"file with special chars", "file@2024.jpg", ""},
		{"long filename", "very_long_filename_that_exceeds_normal_length_limits_12345678901234567890.jpg", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := GeneratePresignedURL(tc.filename)
			assert.Equal(t, tc.expected, url, "Should return empty string when client is not initialized")
		})
	}
}

func TestGeneratePresignedURLNilSafety(t *testing.T) {
	// Explicitly set client to nil and test
	client = nil

	// Call multiple times to ensure no state issues
	for i := 0; i < 5; i++ {
		url := GeneratePresignedURL("test.jpg")
		assert.Equal(t, "", url, "Should consistently return empty string")
	}
}

func TestGeneratePresignedURLConcurrency(t *testing.T) {
	// Ensure client is nil
	client = nil

	// Test concurrent calls don't cause issues
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			url := GeneratePresignedURL("concurrent-test.jpg")
			assert.Equal(t, "", url)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestGeneratePresignedURLEdgeCases(t *testing.T) {
	// Ensure client is nil
	client = nil

	edgeCases := []string{
		"",          // empty
		" ",         // space
		".",         // dot
		"..",        // double dot
		"/",         // slash
		"\\",        // backslash
		"file.jpeg", // different extension
		"file.png",  // another extension
		"FILE.JPG",  // uppercase
		"file",      // no extension
		"file.",     // trailing dot
		".file",     // leading dot
	}

	for _, filename := range edgeCases {
		url := GeneratePresignedURL(filename)
		assert.Equal(t, "", url, "Should return empty string for: %s", filename)
	}
}

func TestInitClientSetsGlobalClient(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Reset client to nil
	client = nil
	assert.Nil(t, client, "Client should be nil before initialization")

	// Call InitClient
	InitClient()

	// Verify that the global client variable was set
	assert.NotNil(t, client, "Client should be set after InitClient")
}

func TestInitClientInitializesValidClient(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Reset client to nil
	client = nil

	// Call InitClient
	InitClient()

	// Verify client is initialized and has expected properties
	assert.NotNil(t, client, "Client should be initialized")

	// The client should be a valid minio.Client instance
	// We can't test actual operations without a running MinIO server,
	// but we can verify the client object exists
	assert.IsType(t, &minio.Client{}, client, "Client should be a minio.Client")
}

func TestInitClientAfterPreviousInitialization(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Initialize client first time
	client = nil
	InitClient()
	assert.NotNil(t, client, "First initialization should create client")

	// Initialize again
	InitClient()
	assert.NotNil(t, client, "Second initialization should create client")
}
