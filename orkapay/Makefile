.PHONY: help up down build test lint migrate

help:
	@echo "Comandos disponíveis:"
	@echo "  make up       - Sobe todos os containers"
	@echo "  make down     - Derruba todos os containers"
	@echo "  make build    - Build de todos os serviços"
	@echo "  make test     - Executa testes de todos os serviços"
	@echo "  make lint     - Executa linting em todos os serviços"
	@echo "  make migrate  - Executa migrações do banco de dados"

up:
	docker-compose up -d

down:
	docker-compose down

build:
	@echo "Building services..."
	cd services/payment-core && go build -o bin/payment-core ./cmd/main.go
	cd services/cashin-engine && go build -o bin/cashin-engine ./cmd/main.go
	cd services/cashout-engine && go build -o bin/cashout-engine ./cmd/main.go
	cd services/boleto-engine && go build -o bin/boleto-engine ./cmd/main.go
	cd services/rewards-engine && go build -o bin/rewards-engine ./cmd/main.go
	cd services/smart-retry && go build -o bin/smart-retry ./cmd/main.go
	cd services/reconciliation && go build -o bin/reconciliation ./cmd/main.go
	cd services/notification && go build -o bin/notification ./cmd/main.go

test:
	@echo "Running tests..."
	cd services/payment-core && go test ./...
	cd services/cashin-engine && go test ./...
	cd services/cashout-engine && go test ./...
	cd services/boleto-engine && go test ./...
	cd services/rewards-engine && go test ./...
	cd services/smart-retry && go test ./...
	cd services/reconciliation && go test ./...
	cd services/notification && go test ./...

lint:
	@echo "Running linters..."
	cd services/payment-core && golangci-lint run
	cd services/cashin-engine && golangci-lint run
	cd services/cashout-engine && golangci-lint run
	cd services/boleto-engine && golangci-lint run
	cd services/rewards-engine && golangci-lint run
	cd services/smart-retry && golangci-lint run
	cd services/reconciliation && golangci-lint run
	cd services/notification && golangci-lint run

migrate:
	@echo "Running database migrations..."
	cd api-gateway && npx typeorm migration:run
