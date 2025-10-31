package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/thiagozs/go-awstools"
)

func main() {
	accessKeyID := envOrDefault("AWS_ACCESS_KEY_ID", "minioadmin")
	secretAccessKey := envOrDefault("AWS_SECRET_ACCESS_KEY", "minioadmin")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	endpoint := envOrDefault("AWS_S3_ENDPOINT", "http://localhost:9000")
	bucketName := envOrDefault("AWS_S3_BUCKET", "estellarx")

	tools, err := awstools.NewAWSTools(
		awstools.WithAccessKeyID(accessKeyID),
		awstools.WithSecretKey(secretAccessKey),
		awstools.WithSessionToken(sessionToken),
		awstools.WithRegion("us-east-1"),
		awstools.WithEndpoint(endpoint),
		awstools.WithDisableSSL(true),
	)
	if err != nil {
		log.Fatalf("failed to create awstools: %v", err)
	}

	localFile := projectPath("assets/gopher_this_fine.png")
	remoteKey := filepath.Base(localFile)

	err = tools.UploadFileToS3WithOptions(
		bucketName,
		remoteKey,
		localFile,
		awstools.WithUploadContentType("image/png"),
		awstools.WithUploadMetadata(map[string]string{
			"processed-by": "go-awstools",
		}),
		awstools.WithUploadCacheControl("public, max-age=604800"),
		awstools.WithUploadACL(types.ObjectCannedACLPublicRead),
	)
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}

	log.Println("upload completed successfully")
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func projectPath(rel string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return rel
	}
	base := filepath.Dir(file)
	return filepath.Join(base, rel)
}
