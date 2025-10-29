package awstools

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type CallBack func(lineStr string) error

type AWSTools struct {
	params       *AWSToolsParams
	cfg          aws.Config
	s3Client     *s3.Client
	mu           *sync.Mutex
	queueWorkers int
	exitWorkers  map[int]chan struct{}
	lines        map[string]int64
}

func NewAWSTools(opts ...Options) (*AWSTools, error) {
	params, err := newAWSToolsParams(opts...)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Load default config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(params.Region()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			params.AccessKeyID(),
			params.SecretKey(),
			params.SessionToken(),
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %s", err)
	}

	// Create S3 client with custom options if needed
	var s3Options []func(*s3.Options)

	// If custom endpoint is used (MinIO or custom S3)
	if len(params.Endpoint()) != 0 {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(params.Endpoint())
			o.UsePathStyle = true
		})
	}

	// Handle DisableSSL if needed
	if params.DisableSSL() {
		// Note: In v2, SSL configuration is typically handled at the HTTP client level
		// You might need to customize the HTTP client if you need to disable SSL
		s3Options = append(s3Options, func(o *s3.Options) {
			// Custom HTTP client configuration would go here if needed
		})
	}

	s3Client := s3.NewFromConfig(cfg, s3Options...)

	qWorkers := params.AmountWorkersRLS()
	if qWorkers <= 0 {
		qWorkers = 4
	}

	return &AWSTools{
		params:       params,
		cfg:          cfg,
		s3Client:     s3Client,
		queueWorkers: qWorkers,
		exitWorkers:  make(map[int]chan struct{}),
		lines:        make(map[string]int64),
		mu:           &sync.Mutex{},
	}, nil
}

func (a *AWSTools) IncLine(ref string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lines[ref]++
}

func (a *AWSTools) GetLines(ref string) int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lines[ref]
}

func (a *AWSTools) ResetLines(ref string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lines[ref] = 0
}

func (a *AWSTools) UploadFileToS3(bucket, fileName, filePath string) error {
	return a.UploadFileToS3WithContext(context.Background(), bucket, fileName, filePath)
}

func (a *AWSTools) UploadFileToS3WithContext(ctx context.Context, bucket, fileName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filePath, err)
	}
	defer file.Close()

	uploader := manager.NewUploader(a.s3Client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})

	return err
}

func (a *AWSTools) DownloadFileFromS3(bucket, fileName, filePath string) error {
	return a.DownloadFileFromS3WithContext(context.Background(), bucket, fileName, filePath)
}

func (a *AWSTools) DownloadFileFromS3WithContext(ctx context.Context, bucket, fileName, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %q, %v", filePath, err)
	}
	defer file.Close()

	downloader := manager.NewDownloader(a.s3Client)
	_, err = downloader.Download(ctx, file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fileName),
		})
	if err != nil {
		return fmt.Errorf("failed to download file, %v", err)
	}

	return nil
}

func (a *AWSTools) ListFilesInBucket(bucket string) ([]types.Object, error) {
	return a.ListFilesInBucketWithContext(context.Background(), bucket)
}

func (a *AWSTools) ListFilesInBucketWithContext(ctx context.Context, bucket string) ([]types.Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	result, err := a.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to list items in bucket %q, %v", bucket, err)
	}

	return result.Contents, nil
}

func (a *AWSTools) ListBuckets() ([]types.Bucket, error) {
	return a.ListBucketsWithContext(context.Background())
}

func (a *AWSTools) ListBucketsWithContext(ctx context.Context) ([]types.Bucket, error) {
	result, err := a.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list buckets: %s", err)
	}

	return result.Buckets, nil
}

func (a *AWSTools) DeleteFileInS3(bucket, fileName string) error {
	return a.DeleteFileInS3WithContext(context.Background(), bucket, fileName)
}

func (a *AWSTools) DeleteFileInS3WithContext(ctx context.Context, bucket, fileName string) error {
	_, err := a.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return fmt.Errorf("unable to delete object %q from bucket %q, %v", fileName, bucket, err)
	}

	return nil
}

func (a *AWSTools) ReadFileStreamFromS3(bucket, fileName string, cb CallBack) chan error {
	return a.ReadFileStreamFromS3WithContext(context.Background(), bucket, fileName, cb)
}

func (a *AWSTools) ReadFileStreamFromS3WithContext(ctx context.Context, bucket, fileName string, cb CallBack) chan error {
	errorChan := make(chan error, a.queueWorkers)
	queueFS := make(chan string, a.params.BufferLimit())

	wg := &sync.WaitGroup{}

	// Start workers
	for i := 0; i < a.queueWorkers; i++ {
		wg.Add(1)
		a.exitWorkers[i] = make(chan struct{}, 1)
		go a.workerReadStreamLine(i, wg, queueFS, a.exitWorkers[i], cb, fileName)
	}

	// Read file and send lines to workers
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer func() {
			close(queueFS)
			wg.Done()
		}()

		resp, err := a.s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fileName),
		})
		if err != nil {
			errorChan <- fmt.Errorf("Failed to get file: %v", err)
			return
		}
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		fmt.Println("Starting to read lines from S3")

		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				// send remaining part if any
				if len(line) > 0 {
					queueFS <- line
				}
				break
			}

			if err != nil {
				errorChan <- fmt.Errorf("Read line error: %v", err)
				return
			}

			queueFS <- line
		}
	}(wg)

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	return errorChan
}

func (a *AWSTools) workerReadStreamLine(id int, wg *sync.WaitGroup,
	lines <-chan string, exit <-chan struct{}, cb CallBack, fileName string) {
	defer wg.Done()
	fmt.Printf("Worker %d started\n", id)
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				fmt.Printf("Worker %d done\n", id)
				return
			}

			if err := cb(line); err != nil {
				fmt.Printf("Worker %d error: %s\n", id, err)
				return
			}

			// Increment line counter
			a.IncLine(fileName)

		case <-exit:
			fmt.Printf("Worker %d received exit\n", id)
			return

		case <-time.After(120 * time.Second):
			fmt.Printf("Worker %d received timeout\n", id)
			return
		}
	}
}

func (a *AWSTools) stopReadFileStreamFromS3() {
	for i := 0; i < a.queueWorkers; i++ {
		a.exitWorkers[i] <- struct{}{}
	}
}

func (a *AWSTools) MoveFileInS3(bucket, source, dest string) error {
	return a.MoveFileInS3WithContext(context.Background(), bucket, source, dest)
}

func (a *AWSTools) MoveFileInS3WithContext(ctx context.Context, bucket, source, dest string) error {
	if err := a.CopyFileInS3WithContext(ctx, bucket, source, dest); err != nil {
		return err
	}

	if err := a.DeleteFileInS3WithContext(ctx, bucket, source); err != nil {
		return err
	}

	return nil
}

func (a *AWSTools) CopyFileInS3(bucket, source, dest string) error {
	return a.CopyFileInS3WithContext(context.Background(), bucket, source, dest)
}

func (a *AWSTools) CopyFileInS3WithContext(ctx context.Context, bucket, source, dest string) error {
	_, err := a.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		CopySource: aws.String(bucket + "/" + source),
		Key:        aws.String(dest),
	})

	if err != nil {
		return fmt.Errorf("unable to copy object %q from bucket %q to %q, %v", source, bucket, dest, err)
	}

	return nil
}
