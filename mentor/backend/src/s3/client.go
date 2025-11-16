package s3

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	
	"mentor-backend/reliability"
)

var client *minio.Client

// GetEndpoint returns the MinIO endpoint with environment variable fallback
func GetEndpoint() string {
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return "minio.default.svc.cluster.local:9000"
}

// GetAccessKey returns the MinIO access key with environment variable fallback
func GetAccessKey() string {
	if accessKey := os.Getenv("MINIO_ACCESS_KEY"); accessKey != "" {
		return accessKey
	}
	return "minioadmin"
}

// GetSecretKey returns the MinIO secret key with environment variable fallback
func GetSecretKey() string {
	if secretKey := os.Getenv("MINIO_SECRET_KEY"); secretKey != "" {
		return secretKey
	}
	return "minioadmin1234"
}

// GetClient returns the initialized MinIO client
func GetClient() *minio.Client {
	return client
}

func InitClient() {
	endpoint := GetEndpoint()
	accessKey := GetAccessKey()
	secretKey := GetSecretKey()

	// Use retry logic for MinIO connection
	retryConfig := reliability.ExternalServiceRetryConfig()
	ctx := context.Background()

	err := reliability.RetryWithBackoff(ctx, retryConfig, func() error {
		var initErr error
		client, initErr = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false,
		})
		if initErr != nil {
			return fmt.Errorf("failed to initialize MinIO client: %v", initErr)
		}

		// Test the connection by checking if we can list buckets
		_, testErr := client.ListBuckets(context.Background())
		if testErr != nil {
			return fmt.Errorf("failed to connect to MinIO: %v", testErr)
		}

		return nil
	})

	if err != nil {
		log.Printf("Warning: MinIO initialization failed after retries: %v", err)
		// Don't fatal here to allow the service to start even if MinIO is temporarily unavailable
		// The health check will report the issue
	} else {
		log.Println("MinIO client initialized successfully")
	}
}

// SetClient allows setting a custom client for testing
func SetClient(c *minio.Client) {
	client = c
}

func GeneratePresignedURL(filename string) string {
	// Return empty string if client is not initialized (e.g., in tests)
	if client == nil {
		return ""
	}

	// Return empty string for empty filename
	if filename == "" {
		return ""
	}

	ctx := context.Background()
	reqParams := url.Values{}
	reqParams.Set("response-content-disposition", "inline")

	presignedURL, err := client.PresignedGetObject(ctx, "screenshots", filename, 1*time.Hour, reqParams)
	if err != nil {
		log.Println("Error generating presigned URL:", err)
		return ""
	}

	return presignedURL.String()
}

// HealthCheck checks if the S3/MinIO connection is healthy
func HealthCheck() error {
	if client == nil {
		return fmt.Errorf("MinIO client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to list buckets as a health check
	_, err := client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("MinIO health check failed: %v", err)
	}

	return nil
}
