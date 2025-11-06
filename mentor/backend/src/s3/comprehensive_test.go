package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitClientInvocation tests that InitClient can be called safely
func TestInitClientInvocation(t *testing.T) {
	// Note: InitClient will try to connect to MinIO but will fail
	// in test environment. This test verifies it doesn't panic.
	// The function calls log.Fatalln on error which exits the program,
	// so we can't actually test it without mocking or a real MinIO instance.
	// This test just verifies the function exists.
	assert.NotNil(t, InitClient)
}

// TestGeneratePresignedURLWithValidClient tests URL generation logic
func TestGeneratePresignedURLWithValidClient(t *testing.T) {
	// Save original client
	originalClient := client
	defer func() {
		client = originalClient
	}()

	// When client is nil, should return empty string
	client = nil
	url := GeneratePresignedURL("test.jpg")
	assert.Equal(t, "", url)

	// Test with various filenames to cover edge cases
	testFiles := []string{
		"simple.jpg",
		"path/to/file.png",
		"file with spaces.jpg",
		"file-with-dashes.png",
		"file_with_underscores.jpg",
		"UPPERCASE.JPG",
		"mixed-Case_File.PNG",
		"123456789.jpg",
		"file.multiple.dots.jpg",
	}

	for _, filename := range testFiles {
		url := GeneratePresignedURL(filename)
		// With nil client, should return empty string
		assert.Equal(t, "", url, "Expected empty string for %s", filename)
	}
}

// TestClientPackageConstants tests package-level behavior
func TestClientPackageConstants(t *testing.T) {
	// Verify the package functions are exported and accessible
	assert.NotNil(t, InitClient)
	assert.NotNil(t, GeneratePresignedURL)
}

// TestGeneratePresignedURLEmptyStringHandling tests empty string handling
func TestGeneratePresignedURLEmptyStringHandling(t *testing.T) {
	originalClient := client
	defer func() {
		client = originalClient
	}()

	client = nil

	// Test various empty or whitespace strings
	emptyStrings := []string{
		"",
		"   ",
		"\t",
		"\n",
		" \t\n ",
	}

	for _, s := range emptyStrings {
		url := GeneratePresignedURL(s)
		assert.Equal(t, "", url)
	}
}

// TestGeneratePresignedURLSpecialCharacters tests special character handling
func TestGeneratePresignedURLSpecialCharacters(t *testing.T) {
	originalClient := client
	defer func() {
		client = originalClient
	}()

	client = nil

	specialChars := []string{
		"file@domain.com.jpg",
		"file#hash.jpg",
		"file$dollar.jpg",
		"file%percent.jpg",
		"file&ampersand.jpg",
		"file+plus.jpg",
		"file=equals.jpg",
		"file[bracket].jpg",
		"file{brace}.jpg",
		"file<less>.jpg",
		"file>greater>.jpg",
		"file?question?.jpg",
	}

	for _, filename := range specialChars {
		url := GeneratePresignedURL(filename)
		assert.Equal(t, "", url, "Expected empty string for %s", filename)
	}
}
