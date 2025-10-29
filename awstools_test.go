package awstools

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// TestNewAWSTools testa a criação de uma nova instância
func TestNewAWSTools(t *testing.T) {
	tools, err := NewAWSTools(
		WithAccessKeyID("test-key"),
		WithSecretKey("test-secret"),
		WithRegion("us-east-1"),
	)
	if err != nil {
		t.Fatalf("Failed to create AWSTools: %v", err)
	}

	if tools == nil {
		t.Fatal("AWSTools instance is nil")
	}

	if tools.queueWorkers != 4 {
		t.Errorf("Expected 4 workers, got %d", tools.queueWorkers)
	}
}

// TestNewAWSToolsWithCustomWorkers testa criação com workers customizados
func TestNewAWSToolsWithCustomWorkers(t *testing.T) {
	tools, err := NewAWSTools(
		WithAccessKeyID("test-key"),
		WithSecretKey("test-secret"),
		WithRegion("us-east-1"),
		WithAmountWorkersRLS(8),
	)
	if err != nil {
		t.Fatalf("Failed to create AWSTools: %v", err)
	}

	if tools.queueWorkers != 8 {
		t.Errorf("Expected 8 workers, got %d", tools.queueWorkers)
	}
}

// TestNewAWSToolsWithEndpoint testa criação com endpoint customizado (MinIO)
func TestNewAWSToolsWithEndpoint(t *testing.T) {
	tools, err := NewAWSTools(
		WithAccessKeyID("minioadmin"),
		WithSecretKey("minioadmin"),
		WithRegion("us-east-1"),
		WithEndpoint("http://localhost:9000"),
		WithDisableSSL(true),
	)
	if err != nil {
		t.Fatalf("Failed to create AWSTools with endpoint: %v", err)
	}

	if tools == nil {
		t.Fatal("AWSTools instance is nil")
	}
}

// TestLineCounter testa as funções de contador de linhas
func TestLineCounter(t *testing.T) {
	tools, err := NewAWSTools(
		WithAccessKeyID("test-key"),
		WithSecretKey("test-secret"),
		WithRegion("us-east-1"),
	)
	if err != nil {
		t.Fatalf("Failed to create AWSTools: %v", err)
	}

	ref := "test-file.txt"

	// Testar valor inicial
	if count := tools.GetLines(ref); count != 0 {
		t.Errorf("Expected initial count 0, got %d", count)
	}

	// Incrementar
	tools.IncLine(ref)
	tools.IncLine(ref)
	tools.IncLine(ref)

	if count := tools.GetLines(ref); count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Reset
	tools.ResetLines(ref)
	if count := tools.GetLines(ref); count != 0 {
		t.Errorf("Expected count 0 after reset, got %d", count)
	}
}

// TestLineCounterConcurrency testa thread-safety do contador
func TestLineCounterConcurrency(t *testing.T) {
	tools, err := NewAWSTools(
		WithAccessKeyID("test-key"),
		WithSecretKey("test-secret"),
		WithRegion("us-east-1"),
	)
	if err != nil {
		t.Fatalf("Failed to create AWSTools: %v", err)
	}

	ref := "concurrent-test.txt"
	iterations := 1000

	// Múltiplas goroutines incrementando
	done := make(chan bool)
	for range 10 {
		go func() {
			for j := 0; j < iterations; j++ {
				tools.IncLine(ref)
			}
			done <- true
		}()
	}

	// Aguardar todas as goroutines
	for range 10 {
		<-done
	}

	expected := int64(10 * iterations)
	if count := tools.GetLines(ref); count != expected {
		t.Errorf("Expected count %d, got %d", expected, count)
	}
}

// TestUploadDownloadIntegration é um teste de integração (skip por padrão)
func TestUploadDownloadIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Usar variáveis de ambiente para credenciais
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucket := os.Getenv("AWS_TEST_BUCKET")

	if accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("AWS credentials not set")
	}

	tools, err := NewAWSTools(
		WithAccessKeyID(accessKey),
		WithSecretKey(secretKey),
		WithRegion("us-east-1"),
	)
	if err != nil {
		t.Fatalf("Failed to create AWSTools: %v", err)
	}

	// Criar arquivo temporário
	tmpFile := "/tmp/test-upload.txt"
	content := []byte("Hello, AWS SDK v2!")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Upload
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	remoteName := "test-upload-" + time.Now().Format("20060102150405") + ".txt"
	if err := tools.UploadFileToS3WithContext(ctx, bucket, remoteName, tmpFile); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Download
	downloadPath := "/tmp/test-download.txt"
	defer os.Remove(downloadPath)

	if err := tools.DownloadFileFromS3WithContext(ctx, bucket, remoteName, downloadPath); err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verificar conteúdo
	downloaded, err := os.ReadFile(downloadPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(downloaded) != string(content) {
		t.Errorf("Content mismatch. Expected %s, got %s", content, downloaded)
	}

	// Cleanup
	if err := tools.DeleteFileInS3WithContext(ctx, bucket, remoteName); err != nil {
		t.Logf("Cleanup failed: %v", err)
	}
}

// TestListBucketsIntegration é um teste de integração
func TestListBucketsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKey == "" || secretKey == "" {
		t.Skip("AWS credentials not set")
	}

	tools, err := NewAWSTools(
		WithAccessKeyID(accessKey),
		WithSecretKey(secretKey),
		WithRegion("us-east-1"),
	)
	if err != nil {
		t.Fatalf("Failed to create AWSTools: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	buckets, err := tools.ListBucketsWithContext(ctx)
	if err != nil {
		t.Fatalf("ListBuckets failed: %v", err)
	}

	t.Logf("Found %d buckets", len(buckets))
	for _, bucket := range buckets {
		t.Logf("Bucket: %s", *bucket.Name)
	}
}

// TestCopyFileIntegration testa copy e move
func TestCopyMoveIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucket := os.Getenv("AWS_TEST_BUCKET")

	if accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("AWS credentials not set")
	}

	tools, err := NewAWSTools(
		WithAccessKeyID(accessKey),
		WithSecretKey(secretKey),
		WithRegion("us-east-1"),
	)
	if err != nil {
		t.Fatalf("Failed to create AWSTools: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Criar arquivo de teste
	tmpFile := "/tmp/test-copy.txt"
	if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	source := "test-source-" + time.Now().Format("20060102150405") + ".txt"
	dest := "test-dest-" + time.Now().Format("20060102150405") + ".txt"
	moved := "test-moved-" + time.Now().Format("20060102150405") + ".txt"

	// Upload arquivo original
	if err := tools.UploadFileToS3WithContext(ctx, bucket, source, tmpFile); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Testar Copy
	if err := tools.CopyFileInS3WithContext(ctx, bucket, source, dest); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Testar Move
	if err := tools.MoveFileInS3WithContext(ctx, bucket, dest, moved); err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	// Verificar que o destino do move existe
	objects, err := tools.ListFilesInBucketWithContext(ctx, bucket)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	foundMoved := false
	foundDest := false
	for _, obj := range objects {
		if *obj.Key == moved {
			foundMoved = true
		}
		if *obj.Key == dest {
			foundDest = true
		}
	}

	if !foundMoved {
		t.Error("Moved file not found")
	}
	if foundDest {
		t.Error("Original file still exists after move")
	}

	// Cleanup
	_ = tools.DeleteFileInS3WithContext(ctx, bucket, source)
	_ = tools.DeleteFileInS3WithContext(ctx, bucket, dest)
	_ = tools.DeleteFileInS3WithContext(ctx, bucket, moved)
}

// BenchmarkIncLine benchmarks the line counter
func BenchmarkIncLine(b *testing.B) {
	tools, _ := NewAWSTools(
		WithAccessKeyID("test"),
		WithSecretKey("test"),
		WithRegion("us-east-1"),
	)

	ref := "bench-test.txt"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tools.IncLine(ref)
	}
}

// BenchmarkIncLineConcurrent benchmarks concurrent access
func BenchmarkIncLineConcurrent(b *testing.B) {
	tools, _ := NewAWSTools(
		WithAccessKeyID("test"),
		WithSecretKey("test"),
		WithRegion("us-east-1"),
	)

	ref := "bench-concurrent.txt"
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tools.IncLine(ref)
		}
	})
}

// Exemplo de tipo verificação - compilará apenas se os tipos estiverem corretos
func _typeCheck() {
	var _ []types.Object
	var _ []types.Bucket
}
