# =============================================================================
# Makefile — OrkaPay Payment Core
# Comandos padronizados para desenvolvimento, teste e build
# =============================================================================

APP_NAME    := payment-core
CMD_DIR     := ./cmd/server
BIN_DIR     := ./bin
BINARY      := $(BIN_DIR)/$(APP_NAME)
COVER_FILE  := coverage.out
COVER_HTML  := coverage.html

# Detectar SO
ifeq ($(OS),Windows_NT)
    RM       := del /Q
    MKDIR    := mkdir
    NULL     := NUL
    SEP      := \\
else
    RM       := rm -f
    MKDIR    := mkdir -p
    NULL     := /dev/null
    SEP      := /
endif

.PHONY: all build run test test-short test-race test-cover lint vet tidy deps deps-audit clean docker-build docker-run git-hooks help

all: clean lint vet test build

# ---- Build ----
build: ## Compilar o binário
	@echo "🔨 Compilando $(APP_NAME)..."
	if not exist $(BIN_DIR) mkdir $(BIN_DIR)
	CGO_ENABLED=0 go build -v -ldflags="-s -w" -o $(BINARY) $(CMD_DIR)
	@echo "✅ Binário: $(BINARY)"

build-linux: ## Compilar para Linux (cross-compile)
	@echo "🔨 Compilando $(APP_NAME) para Linux..."
	$(MKDIR) $(BIN_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -ldflags="-s -w" -o $(BINARY)-linux $(CMD_DIR)
	@echo "✅ Binário: $(BINARY)-linux"

# ---- Run ----
run: ## Executar o servidor localmente
	@echo "🚀 Iniciando $(APP_NAME)..."
	go run $(CMD_DIR)

# ---- Testes ----
test: ## Executar todos os testes
	@echo "🧪 Executando testes..."
	go test ./... -v -count=1

test-short: ## Executar testes rápidos (sem integração)
	@echo "🧪 Executando testes rápidos..."
	go test ./... -short -count=1

test-race: ## Executar testes com Race Detector
	@echo "🧪 Executando testes com Race Detector..."
	go test ./... -race -count=1

test-cover: ## Executar testes com cobertura
	@echo "📊 Executando testes com cobertura..."
	go test ./... -coverprofile=$(COVER_FILE) -covermode=atomic
	go tool cover -html=$(COVER_FILE) -o $(COVER_HTML)
	@echo "✅ Cobertura: $(COVER_HTML)"
	@go tool cover -func=$(COVER_FILE) | findstr total

# ---- Lint ----
lint: ## Executar linters (golangci-lint)
	@echo "🔍 Executando linters..."
	golangci-lint run ./... --timeout=5m || true

vet: ## Executar go vet
	@echo "🔍 Executando go vet..."
	go vet ./...

# ---- Dependências ----
tidy: ## Executar go mod tidy
	go mod tidy

deps: ## Atualizar dependências
	@echo "📦 Atualizando dependências..."
	go get -u ./...
	go mod tidy

deps-audit: ## Auditar dependências (segurança)
	@echo "🔒 Auditando dependências..."
	go list -m -u all 2>&1 | grep -E "\[" || echo "✅ Todas atualizadas"

# ---- Clean ----
clean: ## Limpar artefatos de build
	@echo "🧹 Limpando..."
	$(RM) $(BINARY) 2>$(NULL) || true
	$(RM) $(BINARY)-linux 2>$(NULL) || true
	$(RM) $(COVER_FILE) 2>$(NULL) || true
	$(RM) $(COVER_HTML) 2>$(NULL) || true
	@echo "✅ Limpo"

# ---- Docker ----
docker-build: ## Build da imagem Docker
	@echo "🐳 Buildando imagem Docker..."
	docker build -t orkapay/$(APP_NAME):latest .

docker-run: ## Executar container Docker
	@echo "🐳 Executando container..."
	docker run --rm -p 8080:8080 --name $(APP_NAME) orkapay/$(APP_NAME):latest

# ---- Git ----
git-hooks: ## Instalar hooks do Git
	@echo "🔗 Instalando hooks do Git..."
	cp scripts/pre-commit .git/hooks/pre-commit 2>/dev/null || true
	chmod +x .git/hooks/pre-commit 2>/dev/null || true

# ---- Help ----
help: ## Mostrar esta ajuda
	@echo "╔══════════════════════════════════════════╗"
	@echo "║   OrkaPay Payment Core — Comandos       ║"
	@echo "╚══════════════════════════════════════════╝"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Exemplos:"
	@echo "  make          → clean + lint + vet + test + build"
	@echo "  make run      → go run ./cmd/server"
	@echo "  make test     → go test ./..."
	@echo "  make build    → compilar binário"
