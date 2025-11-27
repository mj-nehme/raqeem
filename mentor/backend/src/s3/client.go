package s3

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"mentor-backend/reliability"
)

var client *minio.Client
var presignClient *minio.Client // separate client for public endpoint presigning (host must match signature)

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

// GetBucketName returns the MinIO bucket name with environment variable fallback
func GetBucketName() string {
	if bucketName := os.Getenv("MINIO_BUCKET_NAME"); bucketName != "" {
		return bucketName
	}
	return "raqeem-screenshots"
}

// GetPublicEndpoint returns externally reachable MinIO endpoint used for browser presigned URLs.
// Falls back to internal endpoint if MINIO_PUBLIC_ENDPOINT is not set.
func GetPublicEndpoint() string {
	if publicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT"); publicEndpoint != "" {
		return publicEndpoint
	}
	return GetEndpoint()
}

// GetClient returns the initialized MinIO client
func GetClient() *minio.Client {
	return client
}

func InitClient() {
	endpoint := GetEndpoint()
	accessKey := GetAccessKey()
	secretKey := GetSecretKey()

	// Strip http:// or https:// prefix if present
	// MinIO client doesn't want the protocol in the endpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

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

		// Allow tests to skip the connectivity check to avoid long network timeouts
		if os.Getenv("MINIO_SKIP_CONNECT") == "1" {
			return nil
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

		// Ensure the bucket exists
		ensureBucketExists()
	}
}

// ensureBucketExists creates the screenshots bucket if it doesn't exist
func ensureBucketExists() {
	if client == nil {
		return
	}

	bucketName := GetBucketName()
	ctx := context.Background()

	// Check if bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Printf("Warning: Failed to check if bucket '%s' exists: %v", bucketName, err)
		return
	}

	if !exists {
		// Create the bucket
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("Warning: Failed to create bucket '%s': %v", bucketName, err)
			return
		}
		log.Printf("Created MinIO bucket: %s", bucketName)
	} else {
		log.Printf("MinIO bucket '%s' already exists", bucketName)
	}
}

// SetClient allows setting a custom client for testing
func SetClient(c *minio.Client) {
	client = c
}

func GeneratePresignedURL(filename string) string {
	// Return empty string if client is not initialized (e.g., in tests)
	if client == nil {
		log.Printf("Warning: MinIO client is not initialized, cannot generate presigned URL for: %s", filename)
		return ""
	}

	// Return empty string for empty filename
	if filename == "" {
		log.Println("Warning: Empty filename provided to GeneratePresignedURL")
		return ""
	}

	ctx := context.Background()
	reqParams := url.Values{}
	reqParams.Set("response-content-disposition", "inline")
	bucketName := GetBucketName()

	publicEndpoint := GetPublicEndpoint()
	internalEndpoint := GetEndpoint()
	publicEndpointClean := strings.TrimPrefix(strings.TrimPrefix(publicEndpoint, "http://"), "https://")
	internalEndpointClean := strings.TrimPrefix(strings.TrimPrefix(internalEndpoint, "http://"), "https://")

	// Attempt stat (non-blocking)
	if _, statErr := client.StatObject(ctx, bucketName, filename, minio.StatObjectOptions{}); statErr != nil {
		log.Printf("Notice: StatObject failed for '%s' in bucket '%s': %v (continuing to presign)", filename, bucketName, statErr)
	}

	// Determine which client to use for presign
	var presignTarget *minio.Client
	if publicEndpointClean != internalEndpointClean {
		if presignClient == nil {
			pc, err := minio.New(publicEndpointClean, &minio.Options{Creds: credentials.NewStaticV4(GetAccessKey(), GetSecretKey(), ""), Secure: false})
			if err != nil {
				log.Printf("Error creating presign client for public endpoint '%s': %v", publicEndpointClean, err)
				return ""
			}
			presignClient = pc
			log.Printf("Presign client initialized for public endpoint: %s", publicEndpointClean)
		}
		presignTarget = presignClient
	} else {
		presignTarget = client
	}

	presignedURL, err := presignTarget.PresignedGetObject(ctx, bucketName, filename, 1*time.Hour, reqParams)
	if err != nil {
		log.Printf("Error generating presigned URL for %s in bucket %s using endpoint %s: %v", filename, bucketName, publicEndpointClean, err)
		return ""
	}
	finalURL := presignedURL.String()
	log.Printf("Presigned URL generated for '%s' via endpoint '%s': %s", filename, publicEndpointClean, finalURL)
	return finalURL
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
