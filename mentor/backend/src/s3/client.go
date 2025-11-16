package s3

import (
	"context"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	var err error
	client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

// SetClient allows setting a custom client for testing
func SetClient(c *minio.Client) {
	client = c
}

// HealthCheck verifies MinIO connectivity
func HealthCheck() error {
	if client == nil {
		return nil // Client not initialized, skip health check (optional dependency)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// List buckets to verify connectivity
	_, err := client.ListBuckets(ctx)
	if err != nil {
		return err
	}

	return nil
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
