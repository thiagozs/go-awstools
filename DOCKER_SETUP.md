# Docker Compose Setup

Setup completo de desenvolvimento com MinIO usando Docker Compose.

## Início Rápido

```bash
# Iniciar todos os serviços
docker-compose up -d

# Verificar status
docker-compose ps

# Ver logs
docker-compose logs -f
```

## Serviços Incluídos

### 1. MinIO (S3-Compatible Storage)

- **API**: <http://localhost:9000>
- **Console**: <http://localhost:9001>
- **Credenciais**:
  - User: `minioadmin`
  - Password: `minioadmin`

### 2. MinIO Setup

- Cria buckets automaticamente:
  - `estellarx`
  - `test-bucket`
  - `backup`
- Configura políticas de acesso

### 3. Go Development (Opcional)

- Container com Go 1.21
- Variáveis de ambiente pré-configuradas
- Volume montado no diretório atual

## Como Usar

### Opção 1: Desenvolvimento Local (Recomendado)

```bash
# 1. Iniciar apenas o MinIO
docker-compose up -d minio minio-setup

# 2. Aguardar inicialização
sleep 10

# 3. Configurar variáveis de ambiente
export AWS_ACCESS_KEY_ID=minioadmin
export AWS_SECRET_ACCESS_KEY=minioadmin
export AWS_S3_ENDPOINT=http://localhost:9000

# 4. Executar seu código
go run main.go
```

### Opção 2: Desenvolvimento no Container

```bash
# 1. Iniciar todos os serviços
docker-compose up -d

# 2. Entrar no container Go
docker-compose exec go-dev bash

# 3. Dentro do container
go mod download
go run main.go

# Ou executar testes
go test -v ./...
```

### Opção 3: Executar Comandos Diretos

```bash
# Executar exemplo
docker-compose exec go-dev go run example_updated.go

# Executar testes
docker-compose exec go-dev go test -v ./...

# Executar streaming
docker-compose exec go-dev go run example_streaming.go
```

## Comandos Úteis

### Gerenciamento de Containers

```bash
# Iniciar serviços
docker-compose up -d

# Parar serviços
docker-compose stop

# Parar e remover containers
docker-compose down

# Remover tudo (incluindo volumes)
docker-compose down -v

# Ver logs
docker-compose logs -f minio
docker-compose logs -f go-dev

# Reiniciar serviço específico
docker-compose restart minio
```

### MinIO CLI no Container

```bash
# Acessar MinIO CLI
docker-compose exec minio-setup mc --help

# Listar buckets
docker exec awstools-minio mc ls local/

# Criar bucket
docker exec awstools-minio mc mb local/novo-bucket

# Upload arquivo
docker exec awstools-minio mc cp /tmp/file.txt local/estellarx/

# Download arquivo
docker exec awstools-minio mc cp local/estellarx/file.txt /tmp/

# Ver política de bucket
docker exec awstools-minio mc policy get local/estellarx
```

### Container Go Dev

```bash
# Bash no container
docker-compose exec go-dev bash

# Executar comando Go
docker-compose exec go-dev go version

# Instalar dependências
docker-compose exec go-dev go mod download

# Formatar código
docker-compose exec go-dev go fmt ./...

# Executar testes
docker-compose exec go-dev go test -v ./...
```

## Estrutura de Volumes

``` text
minio-data/
├── .minio.sys/     # Metadados do MinIO
├── estellarx/      # Bucket principal
├── test-bucket/    # Bucket de testes
└── backup/         # Bucket de backup
```

### Backup e Restore

```bash
# Backup do volume
docker run --rm -v awstools_minio-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/minio-backup.tar.gz -C /data .

# Restore do volume
docker run --rm -v awstools_minio-data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/minio-backup.tar.gz -C /data
```

## Verificação de Saúde

### Verificar MinIO

```bash
# Health check
curl http://localhost:9000/minio/health/live

# Status
curl http://localhost:9000/minio/health/ready
```

### Verificar Buckets

```bash
# Via API
curl -X GET http://localhost:9000 \
  -H "Authorization: AWS4-HMAC-SHA256 ..."

# Via CLI
docker exec awstools-minio mc ls local/
```

## Troubleshooting

### MinIO não inicia

```bash
# Ver logs
docker-compose logs minio

# Verificar porta
lsof -i :9000
netstat -an | grep 9000

# Limpar e reiniciar
docker-compose down -v
docker-compose up -d
```

### Buckets não são criados

```bash
# Verificar logs do setup
docker-compose logs minio-setup

# Recriar setup
docker-compose restart minio-setup

# Ou manualmente
docker exec awstools-minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec awstools-minio mc mb local/estellarx
```

### Erro de permissão

```bash
# Verificar políticas
docker exec awstools-minio mc policy get local/estellarx

# Definir política pública para download
docker exec awstools-minio mc policy set download local/estellarx

# Ou política completa
docker exec awstools-minio mc policy set public local/estellarx
```

### Container Go não conecta ao MinIO

```bash
# Verificar rede
docker network inspect awstools_awstools-network

# Testar conectividade
docker-compose exec go-dev ping minio

# Verificar variáveis de ambiente
docker-compose exec go-dev env | grep AWS
```

## Segurança

### Mudar Credenciais Padrão

Edite o `docker-compose.yml`:

```yaml
environment:
  MINIO_ROOT_USER: seu-usuario
  MINIO_ROOT_PASSWORD: sua-senha-forte
```

### Criar Usuário Adicional

```bash
# Acessar container
docker exec -it awstools-minio bash

# Criar usuário
mc admin user add local newuser newpassword

# Criar política
cat > /tmp/readonly.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::estellarx/*"]
    }
  ]
}
EOF

# Aplicar política
mc admin policy create local readonly /tmp/readonly.json
mc admin policy attach local readonly --user newuser
```

## Monitoramento

### Prometheus Metrics

MinIO expõe métricas em:

- <http://localhost:9000/minio/v2/metrics/cluster>

### Health Checks

```bash
# Live (aceita requisições)
curl http://localhost:9000/minio/health/live

# Ready (pronto para servir)
curl http://localhost:9000/minio/health/ready
```

## Customização

### Adicionar mais buckets

Edite `docker-compose.yml`, seção `minio-setup`:

```yaml
mc mb local/novo-bucket --ignore-existing;
```

### Mudar portas

```yaml
ports:
  - "9002:9000"    # API na porta 9002
  - "9003:9001"    # Console na porta 9003
```

### Persistência em local específico

```yaml
volumes:
  - ./data:/data   # Dados em ./data ao invés de volume Docker
```

## Produção

### Não use para produção

Este setup é para **desenvolvimento apenas**. Para produção:

1. ✅ Use MinIO em cluster (distributed mode)
2. ✅ Configure SSL/TLS
3. ✅ Use credenciais fortes
4. ✅ Configure backup automático
5. ✅ Use volumes com melhor performance
6. ✅ Configure monitoring e alertas
7. ✅ Implemente políticas de acesso granulares

### Exemplo Produção (referência)

```yaml
# NÃO USE ESTE ARQUIVO, É APENAS REFERÊNCIA!
version: '3.8'
services:
  minio1:
    image: minio/minio:latest
    volumes:
      - /mnt/disk1/data:/data1
      - /mnt/disk2/data:/data2
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
    command: server http://minio{1...4}/data{1...2} --console-address ":9001"
    # ... configuração de cluster
```

## Recursos

- [MinIO Documentation](https://min.io/docs/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [MinIO Client (mc)](https://min.io/docs/minio/linux/reference/minio-mc.html)

---

Agora **Happy Developing! 🎉**
