package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"blog/pkg/awstools"
)

func main() {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
	bucketName := "estellarx"
	fileName := "helloworld.txt"

	opts := []awstools.Options{
		awstools.WithRegion("us-east-1"),
		awstools.WithAccessKeyID(accessKeyID),
		awstools.WithSecretKey(secretAccessKey),
		awstools.WithSessionToken(sessionToken),
		awstools.WithEndpoint(endpoint),
		awstools.WithDisableSSL(true),
		awstools.WithAmountWorkersRLS(8),  // 8 workers paralelos
		awstools.WithBufferLimit(1000),     // Buffer de 1000 linhas
	}

	t, err := awstools.NewAWSTools(opts...)
	if err != nil {
		log.Fatalf("Failed to create AWSTools: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Verificar se o arquivo existe
	log.Printf("Checking if file %q exists in bucket %q...\n", fileName, bucketName)
	objs, err := t.ListFilesInBucketWithContext(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to list files: %s", err)
	}

	fileExists := false
	for _, obj := range objs {
		if *obj.Key == fileName {
			fileExists = true
			log.Printf("✓ File found: %s (size: %d bytes)\n", *obj.Key, obj.Size)
			break
		}
	}

	if !fileExists {
		log.Fatalf("File %q not found in bucket %q. Please run the upload example first.", fileName, bucketName)
	}

	fmt.Println(strings.Repeat("-", 80))
	log.Println("Starting streaming read with parallel processing...")
	log.Println("This will process the file line by line using 8 workers")

	// Reset line counter
	t.ResetLines(fileName)

	// Statistics
	var (
		processedLines   atomic.Int64
		totalWords       atomic.Int64
		totalChars       atomic.Int64
		longestLine      atomic.Int64
		shortestLine     atomic.Int64
		linesWithLorem   atomic.Int64
		errorsEncountered atomic.Int64
	)

	shortestLine.Store(999999) // Initialize with large number

	startTime := time.Now()

	// Define callback to process each line
	callback := func(line string) error {
		// Remove newline
		line = strings.TrimSpace(line)
		if line == "" {
			return nil
		}

		// Update statistics
		processedLines.Add(1)
		
		lineLen := int64(len(line))
		totalChars.Add(lineLen)

		// Update longest line
		for {
			current := longestLine.Load()
			if lineLen <= current {
				break
			}
			if longestLine.CompareAndSwap(current, lineLen) {
				break
			}
		}

		// Update shortest line
		for {
			current := shortestLine.Load()
			if lineLen >= current {
				break
			}
			if shortestLine.CompareAndSwap(current, lineLen) {
				break
			}
		}

		// Count words
		words := strings.Fields(line)
		totalWords.Add(int64(len(words)))

		// Check if line contains "lorem"
		if strings.Contains(strings.ToLower(line), "lorem") {
			linesWithLorem.Add(1)
		}

		// Print progress every 1000 lines
		current := processedLines.Load()
		if current%1000 == 0 {
			elapsed := time.Since(startTime)
			linesPerSec := float64(current) / elapsed.Seconds()
			log.Printf("Progress: %d lines processed (%.2f lines/sec)", current, linesPerSec)
		}

		return nil
	}

	// Start streaming
	errorChan := t.ReadFileStreamFromS3WithContext(ctx, bucketName, fileName, callback)

	// Wait for completion and handle errors
	for err := range errorChan {
		if err != nil {
			log.Printf("⚠ Error during streaming: %v", err)
			errorsEncountered.Add(1)
		}
	}

	elapsed := time.Since(startTime)

	// Print final statistics
	fmt.Println(strings.Repeat("=", 80))
	log.Println("📊 STREAMING STATISTICS")
	fmt.Println(strings.Repeat("=", 80))
	
	lines := processedLines.Load()
	words := totalWords.Load()
	chars := totalChars.Load()
	longest := longestLine.Load()
	shortest := shortestLine.Load()
	lorem := linesWithLorem.Load()
	errors := errorsEncountered.Load()

	log.Printf("Total lines processed:       %d", lines)
	log.Printf("Total words:                 %d", words)
	log.Printf("Total characters:            %d", chars)
	log.Printf("Lines containing 'lorem':    %d", lorem)
	log.Printf("Longest line:                %d chars", longest)
	log.Printf("Shortest line:               %d chars", shortest)
	log.Printf("Errors encountered:          %d", errors)
	log.Printf("Processing time:             %v", elapsed)
	
	if lines > 0 {
		log.Printf("Average line length:         %.2f chars", float64(chars)/float64(lines))
		log.Printf("Average words per line:      %.2f", float64(words)/float64(lines))
		log.Printf("Lines per second:            %.2f", float64(lines)/elapsed.Seconds())
		log.Printf("Words per second:            %.2f", float64(words)/elapsed.Seconds())
		log.Printf("MB per second:               %.2f", float64(chars)/1024/1024/elapsed.Seconds())
	}

	// Verify with internal counter
	internalCount := t.GetLines(fileName)
	log.Printf("\nInternal counter:            %d lines", internalCount)

	if internalCount != lines {
		log.Printf("⚠ Warning: Internal counter mismatch! Expected %d, got %d", lines, internalCount)
	}

	fmt.Println(strings.Repeat("=", 80))
	log.Println("✅ Streaming completed successfully!")
}
