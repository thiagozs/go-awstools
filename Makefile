.PHONY: help build test test-integration test-unit clean run-example run-streaming deps minio-start minio-stop

# Variáveis
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=awstools-example
MINIO_PORT=9000
MINIO_CONSOLE_PORT=9001

# Cores para output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Mostrar esta ajuda
	@echo "$(GREEN)AWS Tools - Comandos disponíveis:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

deps: ## Instalar dependências
	@echo "$(GREEN)Instalando dependências...$(NC)"
	$(GOGET) github.com/aws/aws-sdk-go-v2
	$(GOGET) github.com/aws/aws-sdk-go-v2/config
	$(GOGET) github.com/aws/aws-sdk-go-v2/credentials
	$(GOGET) github.com/aws/aws-sdk-go-v2/service/s3
	$(GOGET) github.com/aws/aws-sdk-go-v2/feature/s3/manager
	@echo "$(GREEN)✓ Dependências instaladas$(NC)"

build: ## Compilar o código
	@echo "$(GREEN)Compilando...$(NC)"
	$(GOBUILD) -v -o $(BINARY_NAME) .
	@echo "$(GREEN)✓ Build completo: $(BINARY_NAME)$(NC)"

test: ## Executar todos os testes
	@echo "$(GREEN)Executando testes...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Testes concluídos$(NC)"

test-unit: ## Executar apenas testes unitários
	@echo "$(GREEN)Executando testes unitários...$(NC)"
	$(GOTEST) -v -short -race ./...
	@echo "$(GREEN)✓ Testes unitários concluídos$(NC)"

test-integration: ## Executar testes de integração (requer AWS/MinIO)
	@echo "$(YELLOW)Executando testes de integração...$(NC)"
	@echo "$(YELLOW)Certifique-se de configurar as variáveis de ambiente:$(NC)"
	@echo "  AWS_ACCESS_KEY_ID"
	@echo "  AWS_SECRET_ACCESS_KEY"
	@echo "  AWS_TEST_BUCKET"
	$(GOTEST) -v -race ./...
	@echo "$(GREEN)✓ Testes de integração concluídos$(NC)"

bench: ## Executar benchmarks
	@echo "$(GREEN)Executando benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

coverage: test ## Gerar relatório de cobertura
	@echo "$(GREEN)Gerando relatório de cobertura...$(NC)"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Relatório gerado: coverage.html$(NC)"

clean: ## Limpar arquivos gerados
	@echo "$(GREEN)Limpando...$(NC)"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -f /tmp/helloworld*.txt
	@echo "$(GREEN)✓ Limpeza concluída$(NC)"

minio-start: ## Iniciar MinIO local com Docker
	@echo "$(GREEN)Iniciando MinIO...$(NC)"
	@docker run -d \
		--name minio-dev \
		-p $(MINIO_PORT):9000 \
		-p $(MINIO_CONSOLE_PORT):9001 \
		-e MINIO_ROOT_USER=minioadmin \
		-e MINIO_ROOT_PASSWORD=minioadmin \
		-v minio-data:/data \
		minio/minio server /data --console-address ":9001" || \
		(echo "$(YELLOW)Container já existe, iniciando...$(NC)" && docker start minio-dev)
	@echo "$(GREEN)✓ MinIO rodando em:$(NC)"
	@echo "  API:     http://localhost:$(MINIO_PORT)"
	@echo "  Console: http://localhost:$(MINIO_CONSOLE_PORT)"
	@echo "  User:    minioadmin"
	@echo "  Pass:    minioadmin"
	@sleep 3
	@echo "$(GREEN)Criando bucket 'estellarx'...$(NC)"
	@docker exec minio-dev mc alias set local http://localhost:9000 minioadmin minioadmin 2>/dev/null || true
	@docker exec minio-dev mc mb local/estellarx 2>/dev/null || echo "$(YELLOW)Bucket já existe$(NC)"
	@echo "$(GREEN)✓ Setup completo$(NC)"

minio-stop: ## Parar MinIO local
	@echo "$(GREEN)Parando MinIO...$(NC)"
	@docker stop minio-dev 2>/dev/null || true
	@echo "$(GREEN)✓ MinIO parado$(NC)"

minio-remove: ## Remover container e dados do MinIO
	@echo "$(RED)Removendo MinIO e dados...$(NC)"
	@docker stop minio-dev 2>/dev/null || true
	@docker rm minio-dev 2>/dev/null || true
	@docker volume rm minio-data 2>/dev/null || true
	@echo "$(GREEN)✓ MinIO removido$(NC)"

minio-logs: ## Mostrar logs do MinIO
	@docker logs -f minio-dev

setup-env: ## Criar arquivo .env com variáveis de exemplo
	@echo "$(GREEN)Criando arquivo .env...$(NC)"
	@echo "# AWS Credentials" > .env
	@echo "export AWS_ACCESS_KEY_ID=minioadmin" >> .env
	@echo "export AWS_SECRET_ACCESS_KEY=minioadmin" >> .env
	@echo "export AWS_SESSION_TOKEN=" >> .env
	@echo "export AWS_S3_ENDPOINT=http://localhost:9000" >> .env
	@echo "export AWS_TEST_BUCKET=estellarx" >> .env
	@echo "" >> .env
	@echo "# Para usar: source .env" >> .env
	@echo "$(GREEN)✓ Arquivo .env criado$(NC)"
	@echo "$(YELLOW)Execute: source .env$(NC)"

run-example: ## Executar exemplo completo
	@echo "$(GREEN)Executando exemplo...$(NC)"
	@if [ -z "$$AWS_ACCESS_KEY_ID" ]; then \
		echo "$(RED)Erro: Variáveis de ambiente não configuradas!$(NC)"; \
		echo "$(YELLOW)Execute: make setup-env && source .env$(NC)"; \
		exit 1; \
	fi
	$(GOCMD) run example_updated.go

run-streaming: ## Executar exemplo de streaming
	@echo "$(GREEN)Executando exemplo de streaming...$(NC)"
	@if [ -z "$$AWS_ACCESS_KEY_ID" ]; then \
		echo "$(RED)Erro: Variáveis de ambiente não configuradas!$(NC)"; \
		echo "$(YELLOW)Execute: make setup-env && source .env$(NC)"; \
		exit 1; \
	fi
	$(GOCMD) run example_streaming.go

quick-test: minio-start setup-env ## Setup completo + executar exemplo
	@echo "$(GREEN)Aguardando MinIO inicializar...$(NC)"
	@sleep 5
	@bash -c "source .env && $(GOCMD) run example_updated.go"

fmt: ## Formatar código
	@echo "$(GREEN)Formatando código...$(NC)"
	$(GOCMD) fmt ./...
	@echo "$(GREEN)✓ Código formatado$(NC)"

lint: ## Executar linter
	@echo "$(GREEN)Executando linter...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "$(YELLOW)Instalando golangci-lint...$(NC)" && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin)
	golangci-lint run ./...
	@echo "$(GREEN)✓ Linting completo$(NC)"

vet: ## Executar go vet
	@echo "$(GREEN)Executando go vet...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)✓ Vet completo$(NC)"

check: fmt vet lint test-unit ## Executar todas as verificações
	@echo "$(GREEN)✅ Todas as verificações passaram!$(NC)"

docker-test: ## Executar testes em container Docker
	@echo "$(GREEN)Executando testes em Docker...$(NC)"
	docker run --rm \
		-v $(PWD):/app \
		-w /app \
		golang:1.21 \
		sh -c "go mod download && go test -v ./..."

install-tools: ## Instalar ferramentas de desenvolvimento
	@echo "$(GREEN)Instalando ferramentas...$(NC)"
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u golang.org/x/tools/cmd/goimports
	@echo "$(GREEN)✓ Ferramentas instaladas$(NC)"

mod-tidy: ## Limpar go.mod
	@echo "$(GREEN)Executando go mod tidy...$(NC)"
	$(GOCMD) mod tidy
	@echo "$(GREEN)✓ go.mod limpo$(NC)"

mod-verify: ## Verificar dependências
	@echo "$(GREEN)Verificando dependências...$(NC)"
	$(GOCMD) mod verify
	@echo "$(GREEN)✓ Dependências verificadas$(NC)"

update-deps: ## Atualizar dependências
	@echo "$(GREEN)Atualizando dependências...$(NC)"
	$(GOCMD) get -u ./...
	$(GOCMD) mod tidy
	@echo "$(GREEN)✓ Dependências atualizadas$(NC)"

.DEFAULT_GOAL := help
