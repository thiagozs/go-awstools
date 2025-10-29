# Troubleshooting Guide - AWS Tools SDK v2

## Erro 403: Access Denied no ListBuckets

### Problema

``` text
unable to list buckets: operation error S3: ListBuckets, 
https response error StatusCode: 403, RequestID: ..., 
api error AccessDenied: Access Denied.
```

### Causa

Este erro ocorre quando:

1. **MinIO**: As credenciais não têm permissão `s3:ListAllMyBuckets`
2. **AWS S3**: A política IAM não inclui `s3:ListAllMyBuckets`
3. **Credenciais limitadas**: Você tem acesso a buckets específicos, mas não pode listar todos

### ⚠️ Isso é Normal!

Em muitos casos de produção, é **comum e seguro** que as credenciais não tenham permissão para listar todos os buckets. Você ainda pode:

- ✅ Upload de arquivos
- ✅ Download de arquivos
- ✅ Listar objetos em buckets específicos
- ✅ Deletar arquivos
- ✅ Copiar/mover arquivos

### Solução 1: Ignorar o Erro (Recomendado)

Se você já sabe qual bucket vai usar, simplesmente ignore o erro:

```go
buckets, err := t.ListBucketsWithContext(ctx)
if err != nil {
    // Ignorar - é comum não ter permissão para listar todos os buckets
    log.Printf("Cannot list all buckets (this is OK): %v", err)
} else {
    // Processar buckets se tiver permissão
    for _, bucket := range buckets {
        log.Printf("Bucket: %s", *bucket.Name)
    }
}
```

### Solução 2: Adicionar Permissões (MinIO)

#### Política MinIO Completa

Crie um arquivo `policy.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "arn:aws:s3:::*"
      ]
    }
  ]
}
```

Aplique usando o MinIO Client:

```bash
mc admin policy create myminio fullaccess policy.json
mc admin user add myminio myuser mypassword
mc admin policy attach myminio fullaccess --user myuser
```

#### Política MinIO Restrita (Bucket Específico)

Para acesso apenas a um bucket específico:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::estellarx"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ],
      "Resource": [
        "arn:aws:s3:::estellarx/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets"
      ],
      "Resource": [
        "arn:aws:s3:::*"
      ]
    }
  ]
}
```

### Solução 3: Adicionar Permissões (AWS IAM)

#### Política AWS IAM Completa

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListAllMyBuckets"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:*"
            ],
            "Resource": [
                "arn:aws:s3:::your-bucket-name",
                "arn:aws:s3:::your-bucket-name/*"
            ]
        }
    ]
}
```

#### Política AWS IAM Mínima (Sem ListAllMyBuckets)

Para trabalhar sem listar todos os buckets:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket",
                "s3:GetBucketLocation"
            ],
            "Resource": "arn:aws:s3:::your-bucket-name"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject"
            ],
            "Resource": "arn:aws:s3:::your-bucket-name/*"
        }
    ]
}
```

## Outros Erros Comuns

### SSL Certificate Error

```text
x509: certificate signed by unknown authority
```

**Solução**: Use `WithDisableSSL(true)` para desenvolvimento local:

```go
opts := []awstools.Options{
    awstools.WithEndpoint("http://localhost:9000"),
    awstools.WithDisableSSL(true),
}
```

⚠️ **Nunca desabilite SSL em produção!**

### Connection Refused

```text
dial tcp 127.0.0.1:9000: connect: connection refused
```

**Causas**:

1. MinIO não está rodando
2. Endpoint incorreto
3. Porta incorreta

**Solução**:

```bash
# Verificar se MinIO está rodando
docker ps | grep minio

# Iniciar MinIO
docker run -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"
```

### Context Deadline Exceeded

```text
context deadline exceeded
```

**Causa**: Operação demorou mais que o timeout definido

**Solução**: Aumentar o timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()
```

### Invalid Access Key ID

```text
The AWS Access Key Id you provided does not exist in our records
```

**Solução**: Verificar variáveis de ambiente:

```bash
echo $AWS_ACCESS_KEY_ID
echo $AWS_SECRET_ACCESS_KEY
```

### No Such Bucket

```text
operation error S3: GetObject, https response error StatusCode: 404,
api error NoSuchBucket: The specified bucket does not exist
```

**Solução**:

1. Verificar nome do bucket
2. Verificar se o bucket existe

```go
objs, err := t.ListFilesInBucket("bucket-name")
if err != nil {
    if strings.Contains(err.Error(), "NoSuchBucket") {
        log.Fatal("Bucket does not exist!")
    }
}
```

1. Criar bucket via MinIO Console ou AWS Console

### Access Denied on File Operations

```text
operation error S3: GetObject, https response error StatusCode: 403,
api error AccessDenied: Access Denied.
```

**Causa**: Credenciais não têm permissão para a operação específica

**Solução**: Verificar política e adicionar permissões necessárias:

- `s3:GetObject` - Para download
- `s3:PutObject` - Para upload
- `s3:DeleteObject` - Para deletar
- `s3:ListBucket` - Para listar objetos

## Debugging

### Habilitar Logs Detalhados do SDK

```go
import (
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/smithy-go/logging"
)

// No NewAWSTools, adicionar logger
cfg.Logger = logging.NewStandardLogger(os.Stdout)
cfg.ClientLogMode = aws.LogRetries | aws.LogRequest | aws.LogResponse
```

### Verificar Conectividade

```bash
# Testar conexão com MinIO
curl http://localhost:9000/minio/health/live

# Testar credenciais
mc alias set myminio http://localhost:9000 minioadmin minioadmin
mc ls myminio
```

### Testar com AWS CLI

```bash
# Configurar endpoint
export AWS_ENDPOINT_URL=http://localhost:9000

# Listar buckets
aws s3 ls --endpoint-url http://localhost:9000

# Upload de teste
aws s3 cp test.txt s3://bucket-name/ --endpoint-url http://localhost:9000
```

## Boas Práticas

1. **Sempre use Context com timeout**

   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

2. **Trate erros específicos**

   ```go
   if err != nil {
       if strings.Contains(err.Error(), "NoSuchBucket") {
           // Bucket não existe
       } else if strings.Contains(err.Error(), "AccessDenied") {
           // Sem permissão
       }
   }
   ```

3. **Não exponha credenciais**
  
   ```go
   // ✅ Bom - usar variáveis de ambiente
   accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
   
   // ❌ Ruim - hardcoded
   accessKey := "AKIAIOSFODNN7EXAMPLE"
   ```

4. **Use retry para operações críticas**
  
   ```go
   maxRetries := 3
   for i := 0; i < maxRetries; i++ {
       err := t.UploadFileToS3WithContext(ctx, bucket, key, path)
       if err == nil {
           break
       }
       time.Sleep(time.Second * time.Duration(i+1))
   }
   ```

## Links Úteis

- [AWS SDK Go v2 Documentation](https://aws.github.io/aws-sdk-go-v2/)
- [MinIO Documentation](https://min.io/docs/minio/linux/index.html)
- [AWS S3 API Reference](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html)
- [MinIO IAM Policies](https://min.io/docs/minio/linux/administration/identity-access-management/policy-based-access-control.html)
