package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Exemplo de uso com SDK v2

func main() {
	// Criar instância do AWSTools
	tools, err := NewAWSTools(
		WithAccessKeyID("your-access-key"),
		WithSecretKey("your-secret-key"),
		WithRegion("us-east-1"),
		// Para MinIO ou S3 customizado:
		// WithEndpoint("http://localhost:9000"),
		// WithDisableSSL(true),
	)
	if err != nil {
		log.Fatalf("Failed to create AWS tools: %v", err)
	}

	// ============================================
	// EXEMPLO 1: Upload de arquivo (compatibilidade)
	// ============================================
	err = tools.UploadFileToS3("my-bucket", "remote-file.txt", "/path/to/local/file.txt")
	if err != nil {
		log.Printf("Upload failed: %v", err)
	}

	// ============================================
	// EXEMPLO 2: Upload com context e timeout
	// ============================================
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = tools.UploadFileToS3WithContext(ctx, "my-bucket", "remote-file.txt", "/path/to/local/file.txt")
	if err != nil {
		log.Printf("Upload failed: %v", err)
	}

	// ============================================
	// EXEMPLO 3: Download de arquivo
	// ============================================
	err = tools.DownloadFileFromS3("my-bucket", "remote-file.txt", "/path/to/local/file.txt")
	if err != nil {
		log.Printf("Download failed: %v", err)
	}

	// ============================================
	// EXEMPLO 4: Listar arquivos em bucket
	// ============================================
	objects, err := tools.ListFilesInBucket("my-bucket")
	if err != nil {
		log.Printf("List failed: %v", err)
	}

	for _, obj := range objects {
		// Note: obj não é mais um ponteiro no v2
		fmt.Printf("File: %s, Size: %d, Modified: %v\n",
			*obj.Key, obj.Size, obj.LastModified)
	}

	// ============================================
	// EXEMPLO 5: Listar todos os buckets
	// ============================================
	buckets, err := tools.ListBuckets()
	if err != nil {
		log.Printf("List buckets failed: %v", err)
	}

	for _, bucket := range buckets {
		// Note: bucket não é mais um ponteiro no v2
		fmt.Printf("Bucket: %s, Created: %v\n",
			*bucket.Name, bucket.CreationDate)
	}

	// ============================================
	// EXEMPLO 6: Ler arquivo linha por linha (streaming)
	// ============================================
	lineCount := 0
	callback := func(line string) error {
		lineCount++
		fmt.Printf("Processing line %d: %s", lineCount, line)
		// Processar a linha aqui
		return nil
	}

	errorChan := tools.ReadFileStreamFromS3("my-bucket", "large-file.txt", callback)

	// Aguardar processamento e verificar erros
	for err := range errorChan {
		if err != nil {
			log.Printf("Error during streaming: %v", err)
		}
	}

	totalLines := tools.GetLines("large-file.txt")
	fmt.Printf("Total lines processed: %d\n", totalLines)

	// ============================================
	// EXEMPLO 7: Streaming com context e cancelamento
	// ============================================
	ctx2, cancel2 := context.WithCancel(context.Background())

	callback2 := func(line string) error {
		// Simular condição de cancelamento
		if lineCount > 1000 {
			cancel2() // Cancelar o context
			return fmt.Errorf("stopping after 1000 lines")
		}
		return nil
	}

	errorChan2 := tools.ReadFileStreamFromS3WithContext(ctx2, "my-bucket", "huge-file.txt", callback2)

	for err := range errorChan2 {
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}

	// ============================================
	// EXEMPLO 8: Copiar arquivo dentro do S3
	// ============================================
	err = tools.CopyFileInS3("my-bucket", "source.txt", "destination.txt")
	if err != nil {
		log.Printf("Copy failed: %v", err)
	}

	// ============================================
	// EXEMPLO 9: Mover arquivo (copiar + deletar)
	// ============================================
	err = tools.MoveFileInS3("my-bucket", "old-name.txt", "new-name.txt")
	if err != nil {
		log.Printf("Move failed: %v", err)
	}

	// ============================================
	// EXEMPLO 10: Deletar arquivo
	// ============================================
	err = tools.DeleteFileInS3("my-bucket", "file-to-delete.txt")
	if err != nil {
		log.Printf("Delete failed: %v", err)
	}

	// ============================================
	// EXEMPLO 11: Reset contador de linhas
	// ============================================
	tools.ResetLines("large-file.txt")

	// ============================================
	// EXEMPLO 12: Uso com MinIO (S3 compatível)
	// ============================================
	minioTools, err := NewAWSTools(
		WithAccessKeyID("minioadmin"),
		WithSecretKey("minioadmin"),
		WithRegion("us-east-1"),
		WithEndpoint("http://localhost:9000"),
		WithDisableSSL(true),
	)
	if err != nil {
		log.Fatalf("Failed to create MinIO tools: %v", err)
	}

	buckets, err = minioTools.ListBuckets()
	if err != nil {
		log.Printf("List MinIO buckets failed: %v", err)
	}

	// ============================================
	// EXEMPLO 13: Operações com retry e timeout
	// ============================================
	ctxRetry, cancelRetry := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelRetry()

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err = tools.UploadFileToS3WithContext(ctxRetry, "my-bucket", "important.txt", "/path/to/file.txt")
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			log.Printf("Upload attempt %d failed, retrying: %v", i+1, err)
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	if err != nil {
		log.Printf("Upload failed after %d attempts: %v", maxRetries, err)
	}
}

// ============================================
// EXEMPLO 14: Processamento paralelo com streaming
// ============================================
func parallelStreamProcessing() {
	tools, err := NewAWSTools(
		WithAccessKeyID("your-key"),
		WithSecretKey("your-secret"),
		WithRegion("us-east-1"),
		WithAmountWorkersRLS(8), // 8 workers paralelos
		WithBufferLimit(1000),    // Buffer de 1000 linhas
	)
	if err != nil {
		log.Fatal(err)
	}

	type Result struct {
		LineNum int
		Data    string
		Error   error
	}

	results := make(chan Result, 100)

	callback := func(line string) error {
		// Processar linha (simulação)
		processed := fmt.Sprintf("Processed: %s", line)
		results <- Result{Data: processed}
		return nil
	}

	ctx := context.Background()
	errorChan := tools.ReadFileStreamFromS3WithContext(ctx, "my-bucket", "data.csv", callback)

	// Goroutine para processar resultados
	go func() {
		for result := range results {
			if result.Error != nil {
				log.Printf("Processing error: %v", result.Error)
				continue
			}
			// Fazer algo com o resultado
			fmt.Println(result.Data)
		}
	}()

	// Aguardar erros
	for err := range errorChan {
		if err != nil {
			log.Printf("Stream error: %v", err)
		}
	}

	close(results)
}
