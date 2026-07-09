# Setup e Funcionalidades — OrkaPay

> **OrkaPay** é uma plataforma de orquestração de pagamentos B2B que conecta o mundo analógico do boleto ao digital do Pix. Oferecemos uma suíte financeira completa com cashback e participação nos lucros.

---

## Sumário

1. [Funcionalidades do Sistema](#1-funcionalidades-do-sistema)
2. [Para Clientes — Como Usar a API](#2-para-clientes--como-usar-a-api)
3. [Para Colaboradores — Instalação e Desenvolvimento](#3-para-colaboradores--instalação-e-desenvolvimento)
4. [Arquitetura dos Serviços](#4-arquitetura-dos-serviços)
5. [Variáveis de Ambiente](#5-variáveis-de-ambiente)
6. [Deploy no Render](#6-deploy-no-render)
7. [Testes](#7-testes)
8. [Troubleshooting](#8-troubleshooting)
9. [Roadmap e Próximos Passos](#9-roadmap-e-próximos-passos)

---

## 1. Funcionalidades do Sistema

### 1.1. Core Financeiro (Fase 3 — Operação) — 92% completo

| Funcionalidade | Endpoint | Descrição |
|----------------|----------|-----------|
| **Criar fatura** | `POST /api/v1/invoices` | Cria cobrança com valor, vencimento e dados do pagador |
| **Emitir fatura** | `POST /api/v1/invoices/{id}/issue` | Gera PIX QR Code, chave Pix e/ou boleto |
| **Pagar fatura** | `POST /api/v1/invoices/{id}/pay` | Registra pagamento manualmente |
| **Cancelar fatura** | `POST /api/v1/invoices/{id}/cancel` | Cancela antes do vencimento |
| **Expirar fatura** | `POST /api/v1/invoices/{id}/expire` | Marca como vencida |
| **Consultar fatura** | `GET /api/v1/invoices/{id}` | Status, QR Code, boleto |
| **Listar faturas** | `GET /api/v1/invoices?status=` | Filtro por status (PENDING, ISSUED, PAID, CANCELLED, EXPIRED) |
| **Criar transferência** | `POST /api/v1/transfers` | Solicita saque Pix |
| **Aprovar transferência** | `POST /api/v1/transfers/{id}/approve` | Aprova para liquidação |
| **Liquidar transferência** | `POST /api/v1/transfers/{id}/settle` | Executa o Pix via provedor |
| **Falhar transferência** | `POST /api/v1/transfers/{id}/fail` | Marca como falha |
| **Cancelar transferência** | `POST /api/v1/transfers/{id}/cancel` | Cancela antes de liquidar |
| **Reconciliação** | `GET /api/v1/reconciliation` | Conciliação entre extrato e invoices |
| **Webhook de provedor** | `POST /webhooks/asaas` | Recebe notificações do Asaas (pagamento confirmado) |
| **Cadastro de webhook** | `POST /api/v1/webhooks` | Cliente registra URL para receber eventos |
| **Teste de webhook** | `POST /api/v1/webhooks/{id}/test` | Envia evento de teste |
| **Health check** | `GET /health` | Status do serviço |

### 1.2. Canais de Entrada

| Canal | Descrição | Status |
|-------|-----------|--------|
| **API REST** (API Gateway NestJS) | 22 endpoints documentados via Swagger | ✅ |
| **WhatsApp** (whatsapp-gateway) | Comando `/cobrar 10.00 Joao Silva` gera invoice | ✅ |
| **Webhook Asaas** | Notificações automáticas de pagamento confirmado | ✅ |

### 1.3. Provedores de Pagamento (BaaS)

| Provedor | Produto | Status |
|----------|---------|--------|
| **Asaas** | Pagamento Pix/Boleto (cobrança) | ✅ Sandbox funcional |
| **Stark Bank** | Transferência Pix (saque) | ⬜ Em implementação |

---

## 2. Para Clientes — Como Usar a API

### 2.1. Obter sua API Key

Sua API Key é fornecida no momento do cadastro. Envie-a no header de todas as requisições:

```bash
X-API-Key: sua-chave-aqui
```

> **Limite:** 100 requisições por minuto por API Key.

### 2.2. Exemplos de Uso

#### Criar e emitir uma fatura (R$ 100,00)

```bash
# 1. Criar a fatura
curl -X POST https://api.orkapay.com/api/v1/invoices \
  -H "X-API-Key: sua-chave" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_cents": 10000,
    "due_date": "2026-08-15",
    "payer_name": "Empresa XPTO",
    "payer_tax_id": "11222333000181"
  }'

# Resposta (201 Created):
# {
#   "ID": "cin-20260709-abc123",
#   "amount_cents": 10000,
#   "status": "PENDING",
#   ...
# }

# 2. Emitir (gerar PIX/boleto)
curl -X POST https://api.orkapay.com/api/v1/invoices/cin-20260709-abc123/issue \
  -H "X-API-Key: sua-chave"

# Resposta: status "ISSUED" com pix_qr_code, pix_key e boleto_barcode
```

#### Consultar uma fatura

```bash
curl -X GET https://api.orkapay.com/api/v1/invoices/cin-20260709-abc123 \
  -H "X-API-Key: sua-chave"
```

#### Listar faturas por status

```bash
curl -X GET "https://api.orkapay.com/api/v1/invoices?status=PAID&limit=10&offset=0" \
  -H "X-API-Key: sua-chave"
```

#### Criar uma transferência (saque Pix)

```bash
curl -X POST https://api.orkapay.com/api/v1/transfers \
  -H "X-API-Key: sua-chave" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_cents": 50000,
    "payee_name": "Fornecedor XPTO",
    "payee_tax_id": "11222333000181",
    "payee_pix_key": "chave@email.com"
  }'
```

#### Configurar webhook para receber notificações

```bash
curl -X POST https://api.orkapay.com/api/v1/webhooks \
  -H "X-API-Key: sua-chave" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://meusistema.com/webhook",
    "events": ["payment.received", "invoice.expired"],
    "secret": "meu-token-secreto"
  }'
```

### 2.3. Fluxo Completo (Primeira Transação)

```
1. POST /api/v1/invoices           → Criar fatura (status: PENDING)
2. POST /api/v1/invoices/{id}/issue → Emitir (status: ISSUED, recebe PIX)
3. Cliente paga via Pix/Boleto     → Asaas envia webhook
4. Webhook recebido                → Status atualiza para PAID
5. GET /api/v1/invoices/{id}       → Confirmar pagamento
```

### 2.4. Ambientes

| Ambiente | URL | Descrição |
|----------|-----|-----------|
| **Produção** | `https://api.orkapay.com` | Ambiente real com provedores reais |
| **Sandbox** | `https://sandbox.orkapay.com` | Testes com Asaas Sandbox |

> No Sandbox, use o CNPJ `11222333000181` para testes — ele é aceito pelo Asaas Sandbox.

---

## 3. Para Colaboradores — Instalação e Desenvolvimento

### 3.1. Pré-requisitos

- Go 1.26+
- Node.js 22+ (para API Gateway)
- Docker + Docker Compose
- Git
- PostgreSQL 16 (via Docker)
- Kafka 3.8+ (via Docker)

### 3.2. Clone e Configuração Inicial

```bash
# Clone o repositório
git clone https://github.com/sua-org/orkapay.git
cd orkapay

# Copie os arquivos .env
cp services/cashin-engine/.env.example services/cashin-engine/.env
cp services/cashout-engine/.env.example services/cashout-engine/.env

# Suba a infraestrutura (PostgreSQL, Kafka, Redis)
docker-compose -f docker-compose.dev.yml up -d

# Execute as migrações dos serviços
cd services/cashin-engine && go run cmd/migrate/main.go && cd ../..
cd services/cashout-engine && go run cmd/migrate/main.go && cd ../..
```

### 3.3. Build e Execução dos Serviços

#### Microsserviços Go (executar em terminais separados)

```bash
# Terminal 1 — Cashin Engine
cd services/cashin-engine
go build -o bin/cashin ./cmd/api
./bin/cashin
# → http://localhost:8080

# Terminal 2 — Cashout Engine
cd services/cashout-engine
go build -o bin/cashout ./cmd/api
./bin/cashout
# → http://localhost:8081

# Terminal 3 — WhatsApp Gateway
cd services/whatsapp-gateway
go build -o bin/whatsapp-gateway ./cmd/api
./bin/whatsapp-gateway
# → http://localhost:8084
```

#### API Gateway (NestJS)

```bash
cd api-gateway
npm install
npm run build
npm run start:dev
# → http://localhost:3000
# → Swagger: http://localhost:3000/api
```

### 3.4. Estrutura de Diretórios

```
orkapay/
├── api-gateway/              # API Gateway (NestJS)
│   └── src/
│       ├── modules/
│       │   ├── cashin/        # 7 endpoints de fatura
│       │   ├── cashout/       # 7 endpoints de transferência
│       │   ├── reconciliation/# 2 endpoints de reconciliação
│       │   ├── notification/  # 4 endpoints de webhook
│       │   └── health/        # 2 endpoints de health check
│       ├── common/            # Guards, interceptors, filters
│       └── config/            # Configuração centralizada
├── services/
│   ├── cashin-engine/         # Microsserviço de cobrança (Go)
│   ├── cashout-engine/        # Microsserviço de saque (Go)
│   ├── whatsapp-gateway/      # Gateway WhatsApp (Go)
│   ├── payment-core/          # Core de pagamentos (Go)
│   ├── ledger-engine/         # Contabilidade (Go)
│   ├── reconciliation/        # Reconciliação (Go)
│   ├── notification/          # Notificações (Go)
│   └── boleto-engine/        # Boleto bancário (Go)
├── shared/                    # Código compartilhado
├── tests/                     # Testes de integração
│   ├── wiremock/              # Mocks de provedores
│   ├── toxiproxy/             # Testes de caos/rede
│   └── load/                  # Testes de carga (k6)
├── monitoring/                # Prometheus + Grafana
└── docs/                      # Documentação
```

---

## 4. Arquitetura dos Serviços

```
┌─────────────────────────────────────────────────────┐
│                    CLIENTES                          │
│  (WhatsApp / API REST / Integração B2B)             │
└──────────┬──────────────────────────────┬───────────┘
           │                              │
           ▼                              ▼
┌──────────────────────┐    ┌──────────────────────────┐
│   API Gateway (NestJS)│    │  WhatsApp Gateway (Go)   │
│   :3000               │    │  :8084                   │
│   22 endpoints REST   │    │  /cobrar via WhatsApp    │
└──────────┬────────────┘    └────────────┬─────────────┘
           │                              │
           ▼                              ▼
┌──────────────────┐    ┌──────────────────┐
│ Cashin Engine (Go)│    │Cashout Engine (Go)│
│ :8080             │    │ :8081             │
│ Invoices/Pix      │    │ Transfers/Pix     │
└────────┬─────────┘    └────────┬─────────┘
         │                       │
         ▼                       ▼
┌──────────────────────────────────────────────┐
│           Provedores de Pagamento             │
│  Asaas (cobrança) / Stark Bank (transferência)│
└──────────────────────────────────────────────┘

┌──────────────────────────────────────────────┐
│           Serviços de Apoio                    │
│  Ledger Engine  → Contabilidade               │
│  Reconciliation → Conciliação                 │
│  Notification  → Webhooks para clientes       │
│  Kafka         → Event Sourcing               │
│  PostgreSQL    → Persistência                 │
│  Prometheus    → Métricas                     │
└──────────────────────────────────────────────┘
```

### 4.1. Fluxo de Eventos (Kafka)

```
Cashin Engine → payment.confirmed → Reconciliation → Ledger → Notification
                                                  ↓
                                             Webhook Cliente
```

---

## 5. Variáveis de Ambiente

### 5.1. Cashin Engine (`services/cashin-engine/.env`)

```env
SERVER_PORT=8080
DATABASE_URL=postgres://orkapay:orkapay@localhost:5432/cashin?sslmode=disable
KAFKA_BROKERS=localhost:9092
ASAAS_API_KEY=$aact_hmlg_...
ASAAS_WALLET_ID=b3abd3a9-...
ASAAS_BASE_URL=https://sandbox.asaas.com/api/v3
```

### 5.2. Cashout Engine (`services/cashout-engine/.env`)

```env
SERVER_PORT=8081
DATABASE_URL=postgres://orkapay:orkapay@localhost:5432/cashout?sslmode=disable
STARK_BANK_API_KEY=...
STARK_BANK_BASE_URL=https://sandbox.starkbank.com/v2
```

### 5.3. WhatsApp Gateway (`services/whatsapp-gateway/.env`)

```env
SERVER_PORT=8084
CASHIN_BASE_URL=http://localhost:8080
CASHIN_API_KEY=demo-key
```

### 5.4. API Gateway

As variáveis são configuradas no arquivo `configuration.ts`:

```env
# Nomes das env vars:
CASHIN_SERVICE_URL=http://localhost:8080
CASHIN_API_KEY=demo-key
CASHOUT_SERVICE_URL=http://localhost:8081
CASHOUT_API_KEY=demo-key
RECONCILIATION_SERVICE_URL=http://localhost:8082
NOTIFICATION_SERVICE_URL=http://localhost:8083
API_KEYS=key1,key2  # Chaves aceitas pelo gateway
```

---

## 6. Deploy no Render

### 6.1. Serviços no Render

O arquivo `render.yaml` define 3 serviços Web (Go) que fazem deploy automático ao fazer push no `main`:

| Serviço | Tipo | Build | Health Check |
|---------|------|-------|-------------|
| `orkapay-cashin` | Web (Go) | `go build -o cashin ./cmd/api` | `/health` |
| `orkapay-cashout` | Web (Go) | `go build -o cashout ./cmd/api` | `/health` |
| `orkapay-whatsapp` | Web (Go) | `go build -o whatsapp-gateway ./cmd/api` | `/health` |

### 6.2. Environment Groups

Todas as variáveis sensíveis são compartilhadas via **Environment Group** `orkapay-production` no Render. Para atualizar:

1. Acesse https://dashboard.render.com → Environment Groups → `orkapay-production`
2. Adicione/edite as variáveis (ex: `ASAAS_API_KEY`, `ASAAS_WALLET_ID`)
3. Salve e faça deploy manual de cada serviço afetado

### 6.3. Deploy Manual

```bash
# 1. Commit e push
git add .
git commit -m "descrição das mudanças"
git push origin main

# 2. Render faz auto-deploy (se configurado) OU:
#    Dashboard → Web Service → Manual Deploy → Deploy latest commit
```

### 6.4. Health Checks

Cada serviço expõe:
- `GET /health` → `{"status":"healthy"}`

Se o health check falhar 3 vezes seguidas, o Render reinicia o serviço automaticamente.

---

## 7. Testes

### 7.1. Testes Unitários

```bash
# Todos os serviços Go
cd services/cashin-engine && go test ./... -v && cd ../..
cd services/cashout-engine && go test ./... -v && cd ../..
cd services/whatsapp-gateway && go test ./... -v && cd ../..
```

### 7.2. Testes de Integração com WireMock

```bash
# Sobe mocks dos provedores
docker-compose -f tests/wiremock/docker-compose.wiremock.yml up -d

# Executa testes
cd tests/wiremock && go test -v
```

### 7.3. Testes de Carga (k6)

```bash
k6 run tests/load/k6-load-test.js
```

### 7.4. Testes de Caos (Toxiproxy)

```bash
cd tests/toxiproxy && go test -v
```

---

## 8. Troubleshooting

### 8.1. Erro: "context canceled"

**Causa:** O webhook do WhatsApp estava usando `r.Context()` (contexto da requisição HTTP), que é cancelado quando a resposta é enviada.

**Solução (já aplicada):** O handler agora usa `context.Background()` para processar mensagens em background. Faça deploy da versão mais recente.

### 8.2. Erro: "CPF/CNPJ informado é inválido"

**Causa:** O WhatsApp enviava `payer_tax_id: "00000000000"` que é inválido para o Asaas.

**Solução (já aplicada):** O client agora usa o CNPJ `11222333000181` como padrão (válido no Asaas Sandbox).

### 8.3. Erro: "Não é permitido split para sua própria carteira"

**Causa:** O código tentava fazer split de R$ 0,50 para a mesma carteira que emite a cobrança, o que o Asaas não permite.

**Solução (já aplicada):** Split desabilitado temporariamente. Será reativado com subcontas (wallets diferentes).

### 8.4. Erro: Webhook Asaas não processado

**Causa:** O webhook chega no endpoint `/webhooks/asaas`, mas o processamento (atualizar invoice + creditar ledger) ainda não foi implementado.

**Status:** O endpoint recebe e loga o webhook, mas o processamento do evento `PAYMENT_RECEIVED` está pendente (próximo sprint).

### 8.5. Serviço não conecta (Render)

```bash
# Verificar logs
curl https://orkapay-cashin.onrender.com/health

# Se timeout, o deploy pode estar em andamento (aguardar 1-2 min)
# Se erro 503, o serviço pode ter caído (verificar dashboard Render)
```

---

## 9. Roadmap e Próximos Passos

### Fase Atual — Tração (Sprint 1)

| Item | Status |
|------|--------|
| [x] Cashin Engine (criação/emissão de invoices) | ✅ Completo |
| [x] Cashout Engine (transferências Pix) | ✅ Completo |
| [x] WhatsApp Gateway (comando `/cobrar`) | ✅ Completo |
| [x] API Gateway (22 endpoints REST) | ✅ Completo |
| [x] Asaas Sandbox (invoices criadas e pagas) | ✅ Testado |
| [ ] Processamento de webhook Asaas (PAYMENT_RECEIVED) | ⬜ Próximo |
| [ ] Integração Ledger Engine (crédito automático) | ⬜ Próximo |
| [ ] Integração Stark Bank (transferências) | ⬜ Próximo |

### Próximas Fases

| Fase | Descrição | Previsão |
|------|-----------|----------|
| **Onboarding** | Cadastro, KYB/KYC, geração de API Key | Sprint 2 |
| **Dashboard** | Frontend web para clientes | Sprint 3 |
| **Boleto Híbrido** | Boleto com QR Code Pix integrado | Sprint 4 |
| **Cashback** | Programa de cashback e revenue share | Sprint 5 |

---

## Suporte

- **Issues técnicas:** Abra uma issue no repositório
- **Dúvidas de integração:** Consulte a documentação Swagger em `/api`
- **Emergência:** Contato direto com o time de operações

---

*Documento gerado em julho de 2026 para clientes e colaboradores do OrkaPay.*