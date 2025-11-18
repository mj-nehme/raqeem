package s3

import (
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
)

func TestGetEndpoint(t *testing.T) {
	// Save original environment
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
	}()

	// Test default endpoint
	_ = os.Unsetenv("MINIO_ENDPOINT")
	assert.Equal(t, "minio.default.svc.cluster.local:9000", GetEndpoint())

	// Test custom endpoint from environment
	_ = os.Setenv("MINIO_ENDPOINT", "custom-minio:9000")
	assert.Equal(t, "custom-minio:9000", GetEndpoint())
}

func TestGetAccessKey(t *testing.T) {
	// Save original environment
	originalAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	defer func() {
		if originalAccessKey != "" {
			_ = os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
		} else {
			_ = os.Unsetenv("MINIO_ACCESS_KEY")
		}
	}()

	// Test default access key
	_ = os.Unsetenv("MINIO_ACCESS_KEY")
	assert.Equal(t, "minioadmin", GetAccessKey())

	// Test custom access key from environment
	_ = os.Setenv("MINIO_ACCESS_KEY", "customuser")
	assert.Equal(t, "customuser", GetAccessKey())
}

func TestGetSecretKey(t *testing.T) {
	// Save original environment
	originalSecretKey := os.Getenv("MINIO_SECRET_KEY")
	defer func() {
		if originalSecretKey != "" {
			_ = os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
		} else {
			_ = os.Unsetenv("MINIO_SECRET_KEY")
		}
	}()

	// Test default secret key
	_ = os.Unsetenv("MINIO_SECRET_KEY")
	assert.Equal(t, "minioadmin1234", GetSecretKey())

	// Test custom secret key from environment
	_ = os.Setenv("MINIO_SECRET_KEY", "custompass")
	assert.Equal(t, "custompass", GetSecretKey())
}

func TestGetClient(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Test when client is nil
	client = nil
	assert.Nil(t, GetClient())

	// Test when client is set
	mockClient, _ := minio.New("test:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	client = mockClient
	assert.Equal(t, mockClient, GetClient())
}

func TestSetClient(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Create a mock client
	mockClient, _ := minio.New("test:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})

	// Test setting client
	SetClient(mockClient)
	assert.Equal(t, mockClient, client)

	// Test setting nil client
	SetClient(nil)
	assert.Nil(t, client)
}

func TestGeneratePresignedURLWithEmptyFilename(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Create a mock client
	mockClient, _ := minio.New("test:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	SetClient(mockClient)

	// Test empty filename
	url := GeneratePresignedURL("")
	assert.Equal(t, "", url, "Should return empty string for empty filename")
}

func TestGeneratePresignedURLValidationEdgeCases(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	testCases := []struct {
		name        string
		filename    string
		expectEmpty bool
	}{
		{"empty string", "", true},
		{"space only", " ", false},
		{"tab only", "\t", false},
		{"newline only", "\n", false},
		{"valid filename", "test.jpg", false},
	}

	// Create a mock client that will fail (invalid endpoint)
	mockClient, _ := minio.New("invalid:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("test", "test", ""),
		Secure: false,
	})
	SetClient(mockClient)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := GeneratePresignedURL(tc.filename)
			if tc.expectEmpty {
				assert.Equal(t, "", url)
			} else {
				// For non-empty filenames with invalid client, it should return empty due to error
				assert.Equal(t, "", url)
			}
		})
	}
}

func TestInitClientWithEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEndpoint := os.Getenv("MINIO_ENDPOINT")
	originalAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	originalSecretKey := os.Getenv("MINIO_SECRET_KEY")
	originalClient := client

	defer func() {
		if originalEndpoint != "" {
			_ = os.Setenv("MINIO_ENDPOINT", originalEndpoint)
		} else {
			_ = os.Unsetenv("MINIO_ENDPOINT")
		}
		if originalAccessKey != "" {
			_ = os.Setenv("MINIO_ACCESS_KEY", originalAccessKey)
		} else {
			_ = os.Unsetenv("MINIO_ACCESS_KEY")
		}
		if originalSecretKey != "" {
			_ = os.Setenv("MINIO_SECRET_KEY", originalSecretKey)
		} else {
			_ = os.Unsetenv("MINIO_SECRET_KEY")
		}
		client = originalClient
	}()

	// Set custom environment variables
	_ = os.Setenv("MINIO_ENDPOINT", "test-minio:9000")
	_ = os.Setenv("MINIO_ACCESS_KEY", "testuser")
	_ = os.Setenv("MINIO_SECRET_KEY", "testpass")

	// Test that environment variables are used
	assert.Equal(t, "test-minio:9000", GetEndpoint())
	assert.Equal(t, "testuser", GetAccessKey())
	assert.Equal(t, "testpass", GetSecretKey())
}

func TestClientFunctionality(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// Test nil client scenario
	client = nil
	assert.Nil(t, GetClient())
	assert.Equal(t, "", GeneratePresignedURL("test.jpg"))

	// Test with mock client
	mockClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("testuser", "testpass", ""),
		Secure: false,
	})
	assert.NoError(t, err)

	SetClient(mockClient)
	assert.NotNil(t, GetClient())
	assert.Equal(t, mockClient, GetClient())
}
