# Quick Start Guide - AWS Tools SDK v2

## Setup em 5 Minutos

### 1. Iniciar MinIO Local (Opcional - para testes)

```bash
# Usando o Makefile
make minio-start

# Ou manualmente com Docker
docker run -d \
  --name minio-dev \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

# Criar bucket
docker exec minio-dev mc mb local/estellarx
```

Acesse o console: <http://localhost:9001>

- User: `minioadmin`
- Pass: `minioadmin`

### 2. Configurar Variáveis de Ambiente

```bash
# Para MinIO local
export AWS_ACCESS_KEY_ID=minioadmin
export AWS_SECRET_ACCESS_KEY=minioadmin
export AWS_S3_ENDPOINT=http://localhost:9000

# Para AWS S3 real
export AWS_ACCESS_KEY_ID=sua-access-key
export AWS_SECRET_ACCESS_KEY=sua-secret-key
# AWS_S3_ENDPOINT não é necessário para S3 real
```

Ou use o Makefile:

```bash
make setup-env
source .env
```

### 3. Instalar Dependências

```bash
go mod init seu-projeto  # se ainda não existe

# Adicionar dependências
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager
go get github.com/aws/aws-sdk-go-v2/credentials

# Ou use o Makefile
make deps
```

### 4. Código Mínimo

Crie um arquivo `main.go`:

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "seu-projeto/awstools"
)

func main() {
    // Configurar AWS Tools
    tools, err := awstools.NewAWSTools(
        awstools.WithRegion("us-east-1"),
        awstools.WithAccessKeyID(os.Getenv("AWS_ACCESS_KEY_ID")),
        awstools.WithSecretKey(os.Getenv("AWS_SECRET_ACCESS_KEY")),
        awstools.WithEndpoint(os.Getenv("AWS_S3_ENDPOINT")),
        awstools.WithDisableSSL(true), // Apenas para MinIO local
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Upload
    err = tools.UploadFileToS3WithContext(ctx, "estellarx", "test.txt", "local.txt")
    if err != nil {
        log.Fatal(err)
    }
    log.Println("✓ Upload bem-sucedido!")

    // Download
    err = tools.DownloadFileFromS3WithContext(ctx, "estellarx", "test.txt", "downloaded.txt")
    if err != nil {
        log.Fatal(err)
    }
    log.Println("✓ Download bem-sucedido!")
}
```

### 5. Executar

```bash
go run main.go

# Ou com Makefile
make run-example
```

## Exemplos Prontos

### Upload Simples

```go
err := tools.UploadFileToS3("bucket", "remote.txt", "/path/local.txt")
```

### Download Simples

```go
err := tools.DownloadFileFromS3("bucket", "remote.txt", "/path/local.txt")
```

### Listar Arquivos

```go
objects, err := tools.ListFilesInBucket("bucket")
for _, obj := range objects {
    fmt.Printf("%s (%d bytes)\n", *obj.Key, obj.Size)
}
```

### Streaming de Arquivo Grande

```go
callback := func(line string) error {
    fmt.Println(line)
    return nil
}

errorChan := tools.ReadFileStreamFromS3("bucket", "large.txt", callback)
for err := range errorChan {
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
```

### Copiar Arquivo

```go
err := tools.CopyFileInS3("bucket", "source.txt", "dest.txt")
```

### Mover Arquivo

```go
err := tools.MoveFileInS3("bucket", "old.txt", "new.txt")
```

### Deletar Arquivo

```go
err := tools.DeleteFileInS3("bucket", "file.txt")
```

## Comandos Úteis do Makefile

```bash
make help              # Ver todos os comandos
make minio-start       # Iniciar MinIO local
make setup-env         # Criar arquivo .env
make run-example       # Executar exemplo completo
make run-streaming     # Executar exemplo de streaming
make test              # Executar testes
make quick-test        # Setup + executar (tudo de uma vez!)
make clean             # Limpar arquivos temporários
make minio-stop        # Parar MinIO
```

## Teste Completo em 30 Segundos

```bash
# Tudo de uma vez!
make quick-test
```

Isso vai

1. ✅ Iniciar MinIO
2. ✅ Criar bucket
3. ✅ Configurar variáveis
4. ✅ Executar exemplo completo

## ⚠️ Problemas Comuns

### "Access Denied" no ListBuckets

**Normal!** Veja [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - seção "Erro 403"

### "Connection Refused"

MinIO não está rodando:

```bash
make minio-start
```

### "No Such Bucket"

Criar bucket:

```bash
docker exec minio-dev mc mb local/seu-bucket
```

### SSL Certificate Error

Adicione `WithDisableSSL(true)` para MinIO local:

```go
awstools.WithDisableSSL(true)
```

## Próximos Passos

1. ✅ Ler [README.md](README.md) para documentação completa
2. ✅ Ver [examples/usage/examples.go](examples/usage/examples.go) para exemplos avançados
3. ✅ Verificar [TROUBLESHOOTING.md](TROUBLESHOOTING.md) para resolver problemas

## Conceitos Importantes

### Context

Sempre use context com timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### Tratamento de Erros

```go
if err != nil {
    if strings.Contains(err.Error(), "NoSuchBucket") {
        // Bucket não existe
    } else if strings.Contains(err.Error(), "AccessDenied") {
        // Sem permissão
    }
}
```

### Streaming

Use para arquivos grandes (>100MB):

```go
tools, _ := awstools.NewAWSTools(
    // ...
    awstools.WithAmountWorkersRLS(8),  // 8 workers paralelos
    awstools.WithBufferLimit(1000),     // Buffer de 1000 linhas
)
```

## Recursos

- **Documentação AWS SDK v2**: <https://aws.github.io/aws-sdk-go-v2/>
- **MinIO Docs**: <https://min.io/docs/>
- **Go por Exemplo**: <https://gobyexample.com/>

## Dicas de Performance

1. **Use context com timeout apropriado**
   - Upload pequeno: 30s
   - Upload grande: 5m
   - Streaming: 10m+

2. **Configure workers para streaming**

   ```go
   awstools.WithAmountWorkersRLS(8)  // 4-16 workers
   ```

3. **Use buffer adequado**

   ```go
   awstools.WithBufferLimit(1000)  // 100-5000 linhas
   ```

4. **Implemente retry para operações críticas**

   ```go
   for i := 0; i < 3; i++ {
       if err = upload(); err == nil {
           break
       }
       time.Sleep(time.Second * time.Duration(i+1))
   }
   ```

## Checklist

- [ ] Credenciais via variáveis de ambiente (nunca hardcoded)
- [ ] SSL habilitado (sem `WithDisableSSL`)
- [ ] Context com timeout em todas as operações
- [ ] Logging de erros implementado
- [ ] Retry para operações críticas
- [ ] Testes de integração rodando
- [ ] Políticas IAM/MinIO configuradas corretamente
- [ ] Monitoramento de uso de recursos

## Contribuindo

Encontrou um bug? Tem uma sugestão?

1. Abra uma issue
2. Faça um PR
3. Melhore a documentação

---

Agora **Happy coding! 🎉**
