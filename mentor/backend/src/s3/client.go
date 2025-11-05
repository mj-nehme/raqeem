package s3

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var client *minio.Client

func InitClient() {
	endpoint := "minio.default.svc.cluster.local:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin1234"

	var err error
	client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func GeneratePresignedURL(filename string) string {
	// Return empty string if client is not initialized (e.g., in tests)
	if client == nil {
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
