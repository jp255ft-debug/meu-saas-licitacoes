# 📁 Estrutura do Projeto Orkapay

> **Gerado em:** 2026-07-01  
> **Total de arquivos:** 252 (excluindo .git, node_modules, binários e temporários)

---

## 🏗️ Visão Geral

```
orkapay/
├── .clinerules/          # Regras e padrões do Cline
├── .memory/              # Contexto de memória do projeto
├── .vscode/              # Configurações do VS Code
├── api-gateway/          # API Gateway (NestJS/TypeScript)
├── deployments/          # CI/CD e Helm charts
├── docs/                 # Documentação
├── infrastructure/     # Infraestrutura como código
├── proto/                # Protocol Buffers
├── scripts/              # Scripts auxiliares
├── services/             # Microsserviços Go
├── shared/               # Código compartilhado
├── AGENTS.md             # Instruções para agentes
├── LICENSE               # Licença
├── Makefile              # Comandos make
├── README.md             # Documentação principal
├── docker-compose*.yml   # Docker Compose
├── estrutura_projeto.txt # Árvore legada
└── update-structure.ps1  # Script de atualização
```

---

## 📂 Detalhamento por Diretório

### `.clinerules/` — Regras do Cline (8 arquivos)
Regras de codificação, arquitetura e workflow para o assistente Cline.

| Arquivo | Descrição |
|---------|-----------|
| `00-core.md` | Regras fundamentais |
| `01-architecture.md` | Padrões arquiteturais |
| `02-golang-standards.md` | Padrões Go |
| `03-nestjs-standards.md` | Padrões NestJS |
| `04-testing.md` | Estratégia de testes |
| `05-git-workflow.md` | Fluxo Git |
| `06-error-handling.md` | Tratamento de erros |
| `07-structure-map.md` | Mapa de estrutura |

---

### `.memory/` — Contexto do Projeto (5 arquivos)
Memória ativa do projeto para o Cline manter contexto entre sessões.

| Arquivo | Descrição |
|---------|-----------|
| `active-context.md` | Contexto atual |
| `progress.md` | Progresso do projeto |
| `project-brief.md` | Briefing do projeto |
| `system-patterns.md` | Padrões do sistema |
| `tech-context.md` | Contexto técnico |

---

### `api-gateway/` — API Gateway (NestJS)
Gateway unificado para todos os serviços de pagamento.

```
api-gateway/
├── src/
│   ├── common/              # Config, decorators, filters, guards, interceptors, logger, middlewares
│   ├── infrastructure/    # Cache, database, gRPC, HTTP clients (Asaas, Boleto, Celcoin, Fitbank, Starkbank), messaging (Kafka, RabbitMQ)
│   ├── migrations/         # Migrações de banco
│   ├── modules/            # Módulos de negócio
│   │   ├── boleto/         # Emissão e gestão de boletos
│   │   ├── cashin/         # Recebimentos (Pix, Invoice)
│   │   ├── cashout/        # Pagamentos (Transferências, Lotes)
│   │   ├── health/         # Health checks
│   │   ├── merchant/       # Gestão de comerciantes
│   │   ├── payment/        # Pagamentos gerais
│   │   ├── reconciliation/ # Reconciliação
│   │   ├── rewards/        # Recompensas e cashback
│   │   └── subscription/   # Assinaturas
│   └── test/               # Testes e2e e unitários
├── Dockerfile
├── nest-cli.json
├── package.json
└── tsconfig*.json
```

---

### `deployments/` — CI/CD e Helm (11 arquivos)

```
deployments/
├── github/workflows/        # GitHub Actions
│   ├── ci.yaml              # CI pipeline
│   ├── cd-staging.yaml      # Deploy staging
│   └── cd-production.yaml   # Deploy produção
└── helm/                    # Helm charts (8 serviços)
    ├── api-gateway/
    ├── boleto-engine/
    ├── cashin-engine/
    ├── cashout-engine/
    ├── notification/
    ├── payment-core/
    ├── reconciliation/
    ├── rewards-engine/
    └── smart-retry/
```

---

### `docs/` — Documentação (21 arquivos)

```
docs/
├── adr/                     # Architecture Decision Records (7)
│   ├── 001-event-sourcing.md
│   ├── 002-message-broker.md
│   ├── 003-smart-retry.md
│   ├── 004-baas-providers.md
│   ├── 005-cash-in-cash-out.md
│   ├── 006-boleto-hybrid-orchestration.md
│   └── 007-revenue-share-model.md
├── api/
│   └── openapi.yaml         # Especificação OpenAPI
├── business/                # Documentação de negócio (5)
│   ├── README.md
│   ├── business-model.md
│   ├── client-onboarding.md
│   ├── financial-projections.md
│   └── partnership-program.md
├── pricing/                 # Precificação (6)
│   ├── README.md
│   ├── competitor-analysis.md
│   ├── plans-comparison.md
│   ├── pricing-table.md
│   ├── revenue-share-policy.md
│   └── roi-calculator.md
└── runbooks/                # Runbooks operacionais (2)
    ├── incident-response.md
    └── on-call.md
```

---

### `infrastructure/` — Infraestrutura como Código

```
infrastructure/
├── kubernetes/
│   ├── base/                # Manifestos base (configmap, namespace, secrets)
│   └── overlays/            # Overlays por ambiente
│       ├── dev/
│       ├── staging/
│       └── production/
├── monitoring/
│   ├── alerts/              # Regras de alerta
│   ├── dashboards/          # Dashboards Grafana
│   └── logging/             # Configuração de logs
└── terraform/
    ├── environments/        # Configurações por ambiente
    │   ├── dev/
    │   ├── staging/
    │   └── production/
    └── modules/             # Módulos reutilizáveis
        ├── database/
        ├── kubernetes/
        └── networking/
```

---

### `proto/` — Protocol Buffers (5 domínios)

```
proto/
├── boleto/
│   └── boleto.proto
├── cashin/
│   └── cashin.proto
├── cashout/
│   └── cashout.proto
├── payment/
│   └── payment.proto
└── subscription/
    └── subscription.proto
```

---

### `scripts/` — Scripts Auxiliares

```
scripts/
├── deployment/              # Scripts de deploy
├── migrations/              # Scripts de migração
└── seed/                    # Scripts de seed de dados
```

---

### `services/` — Microsserviços Go (8 serviços)

#### `payment-core/` — Core de Pagamentos (mais desenvolvido)

```
payment-core/
├── cmd/
│   ├── main.go                    # Entry point vazio
│   ├── server/
│   │   └── main.go                # Servidor HTTP
│   ├── api/                       # API gRPC (futuro)
│   ├── consumer/                  # Consumers (Kafka, RabbitMQ)
│   └── scheduler/                 # Jobs agendados
├── internal/
│   ├── adapter/
│   │   ├── inbound/
│   │   │   ├── consumer/          # Kafka/RabbitMQ consumers
│   │   │   ├── grpc/              # gRPC handlers
│   │   │   └── rest/
│   │   │       ├── dto/           # Data Transfer Objects
│   │   │       ├── handler/       # HTTP handlers
│   │   │       │   ├── health_handler.go
│   │   │       │   └── payment_handler.go
│   │   │       ├── middleware/    # Middlewares HTTP
│   │   │       └── router/
│   │   │           └── router.go
│   │   └── outbound/
│   │       ├── eventstore/        # Event Store
│   │       ├── gateway/           # Adaptadores de gateway
│   │       │   ├── asaas/
│   │       │   ├── boleto/
│   │       │   ├── celcoin/
│   │       │   ├── fitbank/
│   │       │   └── starkbank/
│   │       ├── messaging/         # Kafka/RabbitMQ publishers
│   │       │   └── kafka/
│   │       │       └── mock_publisher.go
│   │       └── postgres/
│   │           ├── db.go
│   │           ├── mapper.go
│   │           ├── payment_repository.go
│   │           └── migrations/
│   │               ├── 001_create_payments_table.up.sql
│   │               └── 001_create_payments_table.down.sql
│   ├── core/
│   │   ├── application/           # Casos de uso
│   │   │   ├── cancel_payment.go
│   │   │   ├── cancel_payment_test.go
│   │   │   ├── create_payment.go
│   │   │   ├── create_payment_test.go
│   │   │   ├── mocks_test.go
│   │   │   ├── process_payment.go
│   │   │   ├── process_payment_test.go
│   │   │   ├── webhook_payment.go
│   │   │   └── webhook_payment_test.go
│   │   ├── domain/                # Entidades e regras de negócio
│   │   │   ├── errors.go
│   │   │   ├── money.go
│   │   │   ├── money_test.go
│   │   │   ├── payment.go
│   │   │   ├── payment_test.go
│   │   │   ├── pix_key.go
│   │   │   ├── pix_key_test.go
│   │   │   └── event/
│   │   │       └── domain_event.go
│   │   └── port/                  # Interfaces (ports)
│   │       ├── event_publisher.go
│   │       ├── payment_gateway.go
│   │       ├── payment_repository.go
│   │       └── payment_service.go
│   ├── infrastructure/
│   │   ├── cache/
│   │   ├── config/
│   │   ├── database/
│   │   ├── logger/
│   │   └── metrics/
│   ├── service/                   # Serviços de domínio
│   │   ├── boleto/
│   │   ├── cashback/
│   │   ├── payment/
│   │   └── subscription/
│   └── pkg/                       # Pacotes utilitários
│       ├── config/
│       ├── crypto/
│       ├── errors/
│       └── utils/
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
├── coverage.html                  # Relatório de cobertura
├── coverage.out                   # Dados de cobertura
└── payment-core.exe              # Binário compilado
```

#### Outros Serviços (estrutura similar)

| Serviço | Linguagem | Descrição |
|---------|-----------|-----------|
| `boleto-engine/` | Go | Emissão e gestão de boletos |
| `cashin-engine/` | Go | Recebimentos via Pix e Invoice |
| `cashout-engine/` | Go | Transferências e pagamentos |
| `notification/` | Go | Envio de notificações (email, webhook) |
| `reconciliation/` | Go | Reconciliação de transações |
| `rewards-engine/` | Go | Cashback e revenue share |
| `smart-retry/` | Go | Retry inteligente com ML |

---

### `shared/` — Código Compartilhado

```
shared/
├── baas/                        # Clientes BaaS (BaaS as a Service)
│   ├── asaas/
│   │   ├── invoice.go
│   │   └── pix_auto.go
│   ├── boleto/
│   │   ├── boleto.go
│   │   └── hybrid.go
│   ├── celcoin/
│   ├── fitbank/
│   └── starkbank/
│       ├── batch.go
│       └── transfer.go
├── contracts/                   # Schemas Avro
│   ├── boleto/
│   │   ├── boleto_issued.avsc
│   │   └── boleto_paid.avsc
│   ├── cashin/
│   │   ├── invoice_created.avsc
│   │   └── invoice_paid.avsc
│   ├── cashout/
│   │   ├── transfer_created.avsc
│   │   └── transfer_settled.avsc
│   ├── payment/
│   │   ├── payment_created.avsc
│   │   └── payment_settled.avsc
│   └── rewards/
│       ├── cashback_calculated.avsc
│       ├── cashback_credited.avsc
│       └── revenue_share_distributed.avsc
└── types/                       # Tipos compartilhados
```

---

## 📊 Estatísticas

| Categoria | Quantidade |
|-----------|------------|
| **Serviços Go** | 8 |
| **Módulos API Gateway** | 9 |
| **Helm Charts** | 9 |
| **ADRs** | 7 |
| **Schemas Avro** | 9 |
| **Protobufs** | 5 |
| **GitHub Workflows** | 3 |
| **Arquivos Go (payment-core)** | 22 |
| **Arquivos de Teste Go** | 8 |
| **Total de Arquivos** | ~252 |

---

## 🧪 Testes (payment-core)

| Pacote | Testes | Cobertura |
|--------|--------|-----------|
| `internal/core/domain` | 40 | 83.5% |
| `internal/core/application` | 30 | 88.4% |

**Total: 70/70 testes passando** ✅

---

## 🚀 Próximos Passos

1. **Implementar endpoints restantes** no `payment-core` (GET /payments/{id})
2. **Criar testes para adaptadores** (handlers HTTP, repositório PostgreSQL)
3. **Desenvolver outros serviços** (boleto-engine, cashin-engine, etc.)
4. **Configurar CI/CD** com GitHub Actions
5. **Deploy em Kubernetes** com Helm charts

---

*Documento gerado automaticamente. Para atualizar, execute o script `update-structure.ps1`.*
