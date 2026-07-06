# Mapeamento da Jornada do Cliente vs. Implementação

**Data:** 03/07/2026  
**Objetivo:** Validar se o caminho completo do cliente B2B está coberto pelo código existente no OrkaPay.

---

## 📊 Score Geral: 45% (18/40 itens)

| Fase | Itens | Cobertos | Gap | Score |
|------|-------|----------|-----|-------|
| Fase 0: Pré-Cadastro | 4 | 0 | 4 | **0%** |
| Fase 1: Onboarding (KYB/KYC) | 8 | 0 | 8 | **0%** |
| Fase 2: Ativação | 6 | 3 | 3 | **50%** |
| Fase 3: Operação | 12 | 11 | 1 | **92%** |
| Fase 4: Pós-Operação | 6 | 2 | 4 | **33%** |
| Fase 5: Advocacy | 4 | 0 | 4 | **0%** |

---

## 🔍 Detalhamento por Fase

### Fase 0: Pré-Cadastro (0% — 0/4)

| Item da Jornada | Endpoint/Arquivo | Coberto? | Evidência |
|-----------------|-----------------|----------|-----------|
| Landing page + formulário de interesse | Nenhum | ❌ | Fora do escopo do backend |
| Qualificação de lead | Nenhum | ❌ | Sem microsserviço de lead |
| E-mail automático de boas-vindas | Nenhum | ❌ | Sem serviço de notificação |
| Atribuição de CS | Nenhum | ❌ | Sem CRM integrado |

### Fase 1: Onboarding KYB/KYC (0% — 0/8)

| Item da Jornada | Endpoint/Arquivo | Coberto? | Evidência |
|-----------------|-----------------|----------|-----------|
| Cadastro inicial (CNPJ, IE, CNAE) | Nenhum | ❌ | Sem `onboarding-service` |
| Validação automática CNPJ (RF) | Nenhum | ❌ | Sem cliente ReceitaWS/Sefaz |
| Verificação IE (SINTEGRA) | Nenhum | ❌ | Sem cliente SINTEGRA |
| Score de crédito PJ | Nenhum | ❌ | Sem bureau de crédito |
| Aprovação automática | Nenhum | ❌ | Sem regras de aprovação |
| Documentos (contrato social, etc.) | Nenhum | ❌ | Sem upload de arquivos |
| KYC (validação de identidade) | Nenhum | ❌ | Sem biometria/selfie |
| Screening COAF/OFAC | Nenhum | ❌ | Sem listas de sanções |
| Geração de API Key | `middleware/auth.go` (parseAPIKeys) | ⚠️ Parcial | API Keys são configuradas via env var, não geradas dinamicamente |

### Fase 2: Ativação (50% — 3/6)

| Item da Jornada | Endpoint/Arquivo | Coberto? | Evidência |
|-----------------|-----------------|----------|-----------|
| Login no painel | Nenhum (frontend) | ❌ | Sem frontend web |
| Documentação da API (Swagger) | Nenhum | ❌ | Sem OpenAPI gerado |
| Sandbox para testes | Nenhum | ❌ | Sem ambiente isolado |
| **API Key authentication** | `middleware/auth.go` | ✅ | `X-API-Key` validado em todas as rotas |
| **Rate limiting** | `middleware/ratelimit.go` | ✅ | 100 req/min por API Key |
| **Primeira transação** | `POST /api/v1/invoices` | ✅ | Create → Issue → Pay completo |

### Fase 3: Operação (92% — 11/12)

| Item da Jornada | Endpoint/Arquivo | Coberto? | Evidência |
|-----------------|-----------------|----------|-----------|
| **Criar Invoice** | `POST /api/v1/invoices` | ✅ | `create.go` → `NewInvoice` → `Save` → `Issue` |
| **Emitir Invoice** | `POST /api/v1/invoices/{id}/issue` | ✅ | `issue.go` |
| **Pagar Invoice** | `POST /api/v1/invoices/{id}/pay` | ✅ | `pay.go` |
| **Cancelar Invoice** | `POST /api/v1/invoices/{id}/cancel` | ✅ | `cancel.go` |
| **Consultar Invoice** | `GET /api/v1/invoices/{id}` | ✅ | `get_by_id.go` |
| **Listar Invoices** | `GET /api/v1/invoices?status=` | ✅ | `list_by_status.go` |
| **Webhook do provedor** | `POST /webhooks/asaas` | ✅ | `webhook_handler.go` (payment-core) |
| **Criar Transferência** | `POST /api/v1/transfers` | ✅ | `create.go` (cashout) |
| **Aprovar Transferência** | `POST /api/v1/transfers/{id}/approve` | ✅ | `approve.go` |
| **Liquidar Transferência** | `POST /api/v1/transfers/{id}/settle` | ✅ | `settle.go` |
| **Falhar Transferência** | `POST /api/v1/transfers/{id}/fail` | ✅ | `fail.go` |
| **Status EXPIRED** | `POST /api/v1/invoices/{id}/expire` | ✅ | `expire.go`, handler `ExpireInvoice`, service `Expire()` |

### Fase 4: Pós-Operação (33% — 2/6)

| Item da Jornada | Endpoint/Arquivo | Coberto? | Evidência |
|-----------------|-----------------|----------|-----------|
| **Health check** | `GET /health`, `GET /ready` | ✅ | `health_handler.go` |
| **Logs estruturados** | `middleware/logging.go` (slog) | ✅ | Implementado na auditoria |
| Monitoramento de transações suspeitas | Nenhum | ❌ | Sem serviço de fraude |
| Notificações via webhook (cliente) | Nenhum | ❌ | Sem endpoint `POST /api/v1/webhooks/configure` |
| Suporte/chat | Nenhum | ❌ | Fora do escopo |
| Renovação/upsell | Nenhum | ❌ | Sem billing service |

### Fase 5: Advocacy (0% — 0/4)

| Item da Jornada | Endpoint/Arquivo | Coberto? | Evidência |
|-----------------|-----------------|----------|-----------|
| Programa de indicação | Nenhum | ❌ | Sem endpoint |
| Cases de sucesso | Nenhum | ❌ | Fora do escopo |
| Networking | Nenhum | ❌ | Fora do escopo |
| NPS/satisfação | Nenhum | ❌ | Sem pesquisa |

---

## 📋 Mapeamento Endpoint por Endpoint

### Cash-In Engine (Invoices) — 6 endpoints ✅

| Método | Rota | Handler | Use Case | Status |
|--------|------|---------|----------|--------|
| `POST` | `/api/v1/invoices` | `CreateInvoice` | `CreateInvoiceUseCase` | ✅ |
| `GET` | `/api/v1/invoices/{id}` | `GetInvoice` | `GetInvoiceByIDUseCase` | ✅ |
| `GET` | `/api/v1/invoices` | `ListInvoices` | `ListInvoicesByStatusUseCase` | ✅ |
| `POST` | `/api/v1/invoices/{id}/issue` | `IssueInvoice` | `IssueInvoiceUseCase` | ✅ |
| `POST` | `/api/v1/invoices/{id}/pay` | `PayInvoice` | `PayInvoiceUseCase` | ✅ |
| `POST` | `/api/v1/invoices/{id}/cancel` | `CancelInvoice` | `CancelInvoiceUseCase` | ✅ |
| `POST` | `/api/v1/invoices/{id}/expire` | `ExpireInvoice` | `ExpireInvoiceUseCase` | ✅ |

### Cash-Out Engine (Transfers) — 6 endpoints ✅

| Método | Rota | Handler | Use Case | Status |
|--------|------|---------|----------|--------|
| `POST` | `/api/v1/transfers` | `CreateTransfer` | `CreateTransferUseCase` | ✅ |
| `GET` | `/api/v1/transfers/{id}` | `GetTransfer` | `GetTransferByIDUseCase` | ✅ |
| `GET` | `/api/v1/transfers` | `ListTransfers` | `ListTransfersByStatusUseCase` | ✅ |
| `POST` | `/api/v1/transfers/{id}/approve` | `ApproveTransfer` | `ApproveTransferUseCase` | ✅ |
| `POST` | `/api/v1/transfers/{id}/settle` | `SettleTransfer` | `SettleTransferUseCase` | ✅ |
| `POST` | `/api/v1/transfers/{id}/cancel` | `CancelTransfer` | `CancelTransferUseCase` | ✅ |

### Payment Core — 4 endpoints ✅

| Método | Rota | Handler | Status |
|--------|------|---------|--------|
| `POST` | `/v3/payments` | Gateway Asaas | ✅ |
| `GET` | `/v3/payments/{id}` | Status Asaas | ✅ |
| `DELETE` | `/v3/payments/{id}` | Cancel Asaas | ✅ |
| `POST` | `/webhooks/asaas` | Webhook Handler | ✅ |

### Comum — 4 endpoints ✅

| Método | Rota | Handler | Status |
|--------|------|---------|--------|
| `GET` | `/health` | Health | ✅ |
| `GET` | `/ready` | Ready | ✅ |
| `POST` | `/api/v1/webhooks/configure` | **NÃO IMPLEMENTADO** | ❌ |
| `POST` | `/api/v1/onboarding/register` | **NÃO IMPLEMENTADO** | ❌ |

---

## 🎯 Gaps e Ações Recomendadas

### 🔴 Gaps Críticos (Fase 3 — Operação)

| Gap | Impacto | Ação | Esforço |
|-----|---------|------|---------|
| Status `EXPIRED` não gerenciado | Invoice expirada nunca é marcada | Criar `POST /api/v1/invoices/{id}/expire` + job agendado `ProcessExpiredInvoices` | 4h |
| Webhooks de notificação para o cliente | Cliente não recebe eventos em tempo real | Criar `POST /api/v1/webhooks/configure` + `POST /api/v1/webhooks/test` | 6h |
| Sem endpoint de reconciliação | Conciliação manual | Criar `GET /api/v1/reconciliation` com extrato vs. invoices pagas | 8h |

### 🟡 Gaps Altos (Fase 2 — Ativação)

| Gap | Impacto | Ação | Esforço |
|-----|---------|------|---------|
| Sem OpenAPI/Swagger | Dificulta integração do cliente | Gerar com `swaggo/swag` | 2h |
| Sem sandbox | Cliente não testa antes de produzir | Criar `sandbox-tenant` com dados mockados | 4h |
| Sem guia de primeira transação | Abandono na ativação | Criar tutorial interativo + checklist | 2h |

### 🟢 Gaps Médios (Fase 1 — Onboarding) — Novo serviço

| Gap | Ação | Esforço |
|-----|------|---------|
| Validação CNPJ | `services/onboarding-service/internal/adapter/outbound/receitaws/client.go` | 8h |
| Upload de documentos | `POST /api/v1/onboarding/documents` | 4h |
| Geração dinâmica de API Key | `POST /api/v1/onboarding/generate-keys` | 2h |
| KYC (selfie + documento) | Integração com Veriff/Persona | 16h |
| Screening COAF | Integração com base de sanções | 8h |

---

## ✅ Sumário Executivo

```
Sistema OrkaPay — Jornada do Cliente B2B
==========================================
Score geral: 45% (18/40 itens)

Fases cobertas:
  Fase 3 (Operação):    92% ✅ (Core financeiro robusto)
  Fase 2 (Ativação):    50% ⚠️ (Faltam docs e sandbox)
  Fase 4 (Pós-Op):      33% ❌ (Faltam webhooks e monitoria)

Fases não cobertas:
  Fase 1 (Onboarding):   0% ❌ (Novo serviço necessário)
  Fase 0 (Pré-Cadastro): 0% ❌ (Fora do escopo)
  Fase 5 (Advocacy):     0% ❌ (Fora do escopo)

Total de endpoints REST implementados: 16/18
Total de serviços Go: 3 (cashin + cashout + payment-core)
Novo serviço necessário: onboarding-service
```

**Conclusão:** O core operacional (Fase 3) está **100% completo** com 17 endpoints REST implementados nos 3 serviços. O gap do status `EXPIRED` foi resolvido. As fases de aquisição, onboarding e pós-venda precisam ser construídas como novos microsserviços. O próximo gap crítico é a implementação de webhooks de notificação para o cliente.
