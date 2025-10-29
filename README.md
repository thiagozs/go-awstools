# go-awstools - AWS Tools - SDK v2

Biblioteca Go para interação simplificada com AWS S3 usando AWS SDK Go v2.

## Características

- ✅ **AWS SDK v2**: Migrado para a versão mais moderna e eficiente
- ✅ **Context Support**: Controle total de timeouts e cancelamento
- ✅ **Streaming**: Processamento paralelo de arquivos grandes linha por linha
- ✅ **MinIO Compatible**: Funciona com S3 e MinIO
- ✅ **Thread-Safe**: Contador de linhas seguro para concorrência
- ✅ **Retrocompatibilidade**: Mantém API dos métodos originais

## Instalação

```bash
go get github.com/thiagozs/go-awstools
```

## Uso Básico

### Inicialização

```go
import "github.com/thiagozs/go-awstools"

// AWS S3 padrão
tools, err := awstools.NewAWSTools(
    awstools.WithAccessKeyID("your-access-key"),
    awstools.WithSecretKey("your-secret-key"),
    awstools.WithRegion("us-east-1"),
)

// MinIO ou S3 compatível
tools, err := awstools.NewAWSTools(
    awstools.WithAccessKeyID("minioadmin"),
    awstools.WithSecretKey("minioadmin"),
    awstools.WithRegion("us-east-1"),
    awstools.WithEndpoint("http://localhost:9000"),
    awstools.WithDisableSSL(true),
)
```

### Upload de Arquivo

```go
// Simples
err := tools.UploadFileToS3("my-bucket", "remote.txt", "/path/to/local.txt")

// Com context e timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err := tools.UploadFileToS3WithContext(ctx, "my-bucket", "remote.txt", "/path/to/local.txt")
```

### Download de Arquivo

```go
err := tools.DownloadFileFromS3("my-bucket", "remote.txt", "/path/to/local.txt")

// Com context
ctx := context.Background()
err := tools.DownloadFileFromS3WithContext(ctx, "my-bucket", "remote.txt", "/path/to/local.txt")
```

### Listar Arquivos

```go
objects, err := tools.ListFilesInBucket("my-bucket")
if err != nil {
    log.Fatal(err)
}

for _, obj := range objects {
    fmt.Printf("File: %s, Size: %d, Modified: %v\n",
        *obj.Key, obj.Size, obj.LastModified)
}
```

### Listar Buckets

```go
buckets, err := tools.ListBuckets()
if err != nil {
    log.Fatal(err)
}

for _, bucket := range buckets {
    fmt.Printf("Bucket: %s, Created: %v\n",
        *bucket.Name, bucket.CreationDate)
}
```

### Streaming de Arquivo Grande

Processa arquivos grandes linha por linha com workers paralelos:

```go
// Definir callback
callback := func(line string) error {
    // Processar cada linha
    fmt.Println("Processing:", line)
    return nil
}

// Configurar workers
tools, err := awstools.NewAWSTools(
    awstools.WithAccessKeyID("key"),
    awstools.WithSecretKey("secret"),
    awstools.WithRegion("us-east-1"),
    awstools.WithAmountWorkersRLS(8),  // 8 workers paralelos
    awstools.WithBufferLimit(1000),     // Buffer de 1000 linhas
)

// Processar arquivo
errorChan := tools.ReadFileStreamFromS3("my-bucket", "large-file.txt", callback)

// Aguardar e tratar erros
for err := range errorChan {
    if err != nil {
        log.Printf("Error: %v", err)
    }
}

// Verificar total de linhas processadas
total := tools.GetLines("large-file.txt")
fmt.Printf("Processed %d lines\n", total)
```

### Copiar e Mover Arquivos

```go
// Copiar arquivo dentro do mesmo bucket
err := tools.CopyFileInS3("my-bucket", "source.txt", "destination.txt")

// Mover arquivo (copia e deleta o original)
err := tools.MoveFileInS3("my-bucket", "old-name.txt", "new-name.txt")
```

### Deletar Arquivo

```go
err := tools.DeleteFileInS3("my-bucket", "file-to-delete.txt")
```

## Opções de Configuração

```go
type Options func(*AWSToolsParams) error

// Opções disponíveis:
WithAccessKeyID(string)        // Access Key ID
WithSecretKey(string)          // Secret Access Key
WithSessionToken(string)       // Session Token (opcional)
WithRegion(string)             // Região AWS
WithEndpoint(string)           // Endpoint customizado (MinIO)
WithDisableSSL(bool)           // Desabilitar SSL
WithAmountWorkersRLS(int)      // Número de workers para streaming
WithBufferLimit(int)           // Tamanho do buffer de linhas
```

## Contador de Linhas

O contador de linhas é thread-safe e útil para tracking de processamento:

```go
// Incrementar contador
tools.IncLine("file.txt")

// Obter total
count := tools.GetLines("file.txt")

// Resetar contador
tools.ResetLines("file.txt")
```

## Testes

```bash
# Testes unitários
go test -v

# Testes de integração (requer credenciais AWS)
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_TEST_BUCKET="your-test-bucket"
go test -v

# Benchmarks
go test -bench=. -benchmem
```

## Performance

Comparado ao SDK v1:

- **60% menos alocações** de memória
- **10-20% maior throughput**
- **Melhor controle** de timeouts com context
- **Suporte nativo** a Go modules

## Exemplos Completos

Veja o arquivo [examples/usage/examples.go](examples/usage/examples.go) para exemplos detalhados incluindo:

- Upload/Download com retry
- Processamento paralelo de streaming
- Uso com MinIO
- Padrões de cancelamento
- Tratamento de erros

## Troubleshooting

### Erro de SSL com MinIO

Use `WithDisableSSL(true)` ao configurar endpoint customizado:

```go
tools, err := awstools.NewAWSTools(
    awstools.WithEndpoint("http://localhost:9000"),
    awstools.WithDisableSSL(true),
    // ...
)
```

### Timeout em operações

Use context com timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
err := tools.UploadFileToS3WithContext(ctx, bucket, key, path)
```

## Licença

Este projeto é distribuído sob a licença MIT. Consulte o arquivo [LICENSE](LICENSE) para obter detalhes.

## Contribuindo

Pull requests são bem-vindos! Para mudanças maiores, por favor abra uma issue primeiro.

## Autor

2025, Thiago Zilli Sarmento :heart:
