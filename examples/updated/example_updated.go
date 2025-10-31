package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/thiagozs/go-awstools"

	utils "github.com/thiagozs/go-xutils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	accessKeyID := envOrDefault("AWS_ACCESS_KEY_ID", "minioadmin")
	secretAccessKey := envOrDefault("AWS_SECRET_ACCESS_KEY", "minioadmin")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	endpoint := envOrDefault("AWS_S3_ENDPOINT", "http://localhost:9000")
	bucketName := envOrDefault("AWS_S3_BUCKET", "estellarx")

	u := utils.New()

	opts := []awstools.Options{
		awstools.WithRegion("us-east-1"),
		awstools.WithAccessKeyID(accessKeyID),
		awstools.WithSecretKey(secretAccessKey),
		awstools.WithSessionToken(sessionToken),
		awstools.WithEndpoint(endpoint),
		awstools.WithDisableSSL(true),
	}

	t, err := awstools.NewAWSTools(opts...)
	if err != nil {
		log.Fatalf("Failed to create AWSTools: %s", err)
	}

	// Usar context com timeout para todas as operações
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Println("Create file /tmp/helloworld.txt")
	if err := u.Files().SaveFile("/tmp/helloworld.txt", []byte("Hello World!\n")); err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}

	log.Println("Appending lorem ipsum text...")
	for i := 0; i < 10000; i++ {
		if err := u.Files().AppendFile("/tmp/helloworld.txt", []byte(GenRandomLorem(100)+"\n")); err != nil {
			log.Fatalf("Failed to append to file: %s", err)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Upload file /tmp/helloworld.txt to S3")
	if err := t.UploadFileToS3WithContext(ctx, bucketName, "helloworld.txt", "/tmp/helloworld.txt"); err != nil {
		log.Fatalf("Failed to upload file to S3: %s", err)
	}
	log.Println("✓ Upload successful")

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Download file helloworld.txt from S3")
	if err := t.DownloadFileFromS3WithContext(ctx, bucketName, "helloworld.txt", "/tmp/helloworld2.txt"); err != nil {
		log.Fatalf("Failed to download file from S3: %s", err)
	}
	log.Println("✓ Download successful")

	fmt.Println(strings.Repeat("-", 80))
	log.Println("List buckets")
	b, err := t.ListBucketsWithContext(ctx)
	if err != nil {
		// ListBuckets pode falhar com 403 se o usuário não tiver permissão
		// Isso é comum em MinIO com políticas restritas
		log.Printf("⚠ Failed to list buckets (this is OK if you don't have ListAllMyBuckets permission): %s", err)
	} else {
		log.Println("Buckets:")
		for _, bucket := range b {
			log.Printf("  * %s (created: %v)\n", *bucket.Name, bucket.CreationDate)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Printf("List objects in bucket %q\n", bucketName)
	objs, err := t.ListFilesInBucketWithContext(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to list objects in bucket %q: %s", bucketName, err)
	}

	log.Printf("Objects in %q:\n", bucketName)
	if len(objs) == 0 {
		log.Println("  (no objects found)")
	} else {
		for _, obj := range objs {
			log.Printf("  * %s (size: %d bytes, modified: %v)\n",
				*obj.Key, obj.Size, obj.LastModified)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Copy file in S3")
	if err := t.CopyFileInS3WithContext(ctx, bucketName, "helloworld.txt", "helloworld-copy.txt"); err != nil {
		log.Printf("⚠ Failed to copy file in S3: %s", err)
	} else {
		log.Println("✓ Copy successful")
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Move file in S3 (copy then delete)")
	if err := t.MoveFileInS3WithContext(ctx, bucketName, "helloworld-copy.txt", "helloworld-moved.txt"); err != nil {
		log.Printf("⚠ Failed to move file in S3: %s", err)
	} else {
		log.Println("✓ Move successful")
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Delete file helloworld.txt from S3")
	if err := t.DeleteFileInS3WithContext(ctx, bucketName, "helloworld.txt"); err != nil {
		log.Printf("⚠ Failed to delete file in S3: %s", err)
	} else {
		log.Println("✓ Delete successful")
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Delete file helloworld-moved.txt from S3")
	if err := t.DeleteFileInS3WithContext(ctx, bucketName, "helloworld-moved.txt"); err != nil {
		log.Printf("⚠ Failed to delete file in S3: %s", err)
	} else {
		log.Println("✓ Delete successful")
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Delete local file /tmp/helloworld.txt")
	if err := u.Files().RemoveFile("/tmp/helloworld.txt"); err != nil {
		log.Printf("⚠ Failed to delete local file: %s", err)
	} else {
		log.Println("✓ Local file deleted")
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Delete local file /tmp/helloworld2.txt")
	if err := u.Files().RemoveFile("/tmp/helloworld2.txt"); err != nil {
		log.Printf("⚠ Failed to delete local file: %s", err)
	} else {
		log.Println("✓ Local file deleted")
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("✅ All operations completed successfully!")
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func GenRandomLorem(wordCount int) string {
	loremWords := []string{
		"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
		"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
		"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
		"exercitation", "ullamco", "laboris", "nisi", "aliquip", "ex", "ea", "commodo",
		"consequat", "duis", "aute", "irure", "in", "reprehenderit", "voluptate",
		"velit", "esse", "cillum", "fugiat", "nulla", "pariatur", "excepteur",
		"sint", "occaecat", "cupidatat", "non", "proident", "sunt", "culpa", "qui",
		"officia", "deserunt", "mollit", "anim", "id", "est", "laborum",
	}

	var loremText []string
	caser := cases.Title(language.English)

	for i := 0; i < wordCount; i++ {
		word := loremWords[rand.Intn(len(loremWords))]
		loremText = append(loremText, word)

		// Add period every 12 words
		if i > 0 && i%12 == 0 {
			loremText[len(loremText)-1] += "."
		}
	}

	// Capitalize first word
	if len(loremText) > 0 {
		loremText[0] = caser.String(loremText[0])
	}

	// Ensure it ends with a period
	if len(loremText) > 0 && !strings.HasSuffix(loremText[len(loremText)-1], ".") {
		loremText[len(loremText)-1] += "."
	}

	return strings.Join(loremText, " ")
}
