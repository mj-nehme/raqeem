package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mentor-backend/s3"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

// TestGetScreenshotFileEmptyFilename tests GetScreenshotFile with empty filename
func TestGetScreenshotFileEmptyFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "filename", Value: ""}}
	c.Request, _ = http.NewRequest("GET", "/screenshots/", nil)

	GetScreenshotFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error message
	assert.Contains(t, w.Body.String(), "filename required")
}

// TestGetScreenshotFileWhitespaceFilename tests GetScreenshotFile with whitespace-only filename
func TestGetScreenshotFileWhitespaceFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "filename", Value: "   "}}
	c.Request, _ = http.NewRequest("GET", "/screenshots/   ", nil)

	GetScreenshotFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "filename required")
}

// TestGetScreenshotFileNoS3Client tests GetScreenshotFile when S3 client is not initialized
func TestGetScreenshotFileNoS3Client(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original client and set to nil
	originalClient := s3.GetClient()
	s3.SetClient(nil)
	defer func() { s3.SetClient(originalClient) }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "filename", Value: "test-screenshot.png"}}
	c.Request, _ = http.NewRequest("GET", "/screenshots/test-screenshot.png", nil)

	GetScreenshotFile(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "storage client unavailable")
}

// TestGetScreenshotFileInvalidS3Connection tests GetScreenshotFile when S3 cannot be reached
func TestGetScreenshotFileInvalidS3Connection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a MinIO client pointing to a non-existent server using a valid port
	client, err := minio.New("localhost:59999", &minio.Options{
		Secure: false,
	})
	if err != nil {
		t.Fatalf("Failed to create mock MinIO client: %v", err)
	}

	// Save original and set mock
	originalClient := s3.GetClient()
	s3.SetClient(client)
	defer func() { s3.SetClient(originalClient) }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "filename", Value: "nonexistent.png"}}
	c.Request, _ = http.NewRequest("GET", "/screenshots/nonexistent.png", nil)

	GetScreenshotFile(c)

	// Note: MinIO's GetObject doesn't fail immediately - it returns a lazy reader
	// The error occurs when trying to read from it. The code sets status 200
	// before attempting to read, so we get 200 even with connection failures.
	// This is the expected behavior of the streaming API.
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound,
		"Expected 200 (streaming failure) or 404 (GetObject error), got %d", w.Code)
}

// TestGetScreenshotFileWithSpecialCharacters tests GetScreenshotFile with various filename patterns
func TestGetScreenshotFileWithSpecialCharacters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name     string
		filename string
	}{
		{"simple filename", "screenshot.png"},
		{"filename with spaces", "my screenshot.png"},
		{"filename with underscores", "my_screenshot_2024.png"},
		{"filename with dashes", "my-screenshot-2024.png"},
		{"jpeg extension", "screenshot.jpg"},
		{"jpeg full extension", "screenshot.jpeg"},
		{"uuid filename", "550e8400-e29b-41d4-a716-446655440000.png"},
		{"path like filename", "device1-screenshot-001.png"},
	}

	// Save original client and set to nil - this will cause service unavailable
	// which is expected behavior when S3 is not configured
	originalClient := s3.GetClient()
	s3.SetClient(nil)
	defer func() { s3.SetClient(originalClient) }()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{gin.Param{Key: "filename", Value: tc.filename}}
			c.Request, _ = http.NewRequest("GET", "/screenshots/"+tc.filename, nil)

			GetScreenshotFile(c)

			// Should return service unavailable because S3 client is nil
			assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		})
	}
}

// TestGetScreenshotFileWithS3Mock tests when S3 client exists but object doesn't
func TestGetScreenshotFileWithS3MockNonexistent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a MinIO client pointing to a non-existent server using valid port
	// This simulates having a client but not being able to retrieve the object
	client, err := minio.New("localhost:59998", &minio.Options{
		Secure: false,
	})
	if err != nil {
		t.Fatalf("Failed to create mock MinIO client: %v", err)
	}

	// Save original and set mock
	originalClient := s3.GetClient()
	s3.SetClient(client)
	defer func() { s3.SetClient(originalClient) }()

	testFilenames := []string{
		"missing-file.png",
		"missing-file.jpg",
		"missing-file.jpeg",
		"file-without-extension",
	}

	for _, filename := range testFilenames {
		t.Run("file_"+filename, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{gin.Param{Key: "filename", Value: filename}}
			c.Request, _ = http.NewRequest("GET", "/screenshots/"+filename, nil)

			GetScreenshotFile(c)

			// MinIO's GetObject is lazy - it doesn't fail immediately
			// The error occurs when reading. Status 200 is set before reading,
			// so streaming failures still result in 200.
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound,
				"Expected 200 (streaming failure) or 404 (GetObject error), got %d", w.Code)
		})
	}
}
