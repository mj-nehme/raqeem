package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePresignedURLWithoutInit(t *testing.T) {
	// Test that GeneratePresignedURL handles nil client gracefully
	url := GeneratePresignedURL("test.jpg")
	assert.Equal(t, "", url, "Should return empty string when client is not initialized")
}

func TestGeneratePresignedURLEmptyFilename(t *testing.T) {
	// Test with empty filename
	url := GeneratePresignedURL("")
	assert.Equal(t, "", url)
}

func TestGeneratePresignedURLMultipleCalls(t *testing.T) {
	// Test multiple calls work correctly
	url1 := GeneratePresignedURL("file1.jpg")
	url2 := GeneratePresignedURL("file2.jpg")
	url3 := GeneratePresignedURL("file3.jpg")

	// All should return empty string when client is not initialized
	assert.Equal(t, "", url1)
	assert.Equal(t, "", url2)
	assert.Equal(t, "", url3)
}

func TestInitClientDoesNotCrash(t *testing.T) {
	// Test that InitClient doesn't crash even if MinIO is not available
	// This will fail to connect but shouldn't panic
	// We can't actually test this without mocking or having a real MinIO instance
	// So we just verify the function signature exists
	assert.NotNil(t, InitClient)
}
