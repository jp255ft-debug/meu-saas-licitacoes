# 🎯 Plano de Ação Integrado — OrkaPay

> **Gerado em:** 02/07/2026
> **Baseado em:** Auditoria completa de arquivos + Análise de testes + Pesquisa de melhores práticas (Nubank, BaaS, Testcontainers, k6)

---

## 📊 Status Atual do Projeto

| Categoria | Status | Arquivos Implementados | Arquivos Pendentes |
|-----------|--------|:----------------------:|:------------------:|
| **🎯 Core (payment-core)** | ✅ Implementado | ~35 | ~15 (infra, pkg, gateways) |
| **📚 Documentação** | ✅ Completo | 22 | 0 |
| **📦 Infra Básica** | ✅ Pronto | 20 | 0 |
| **📨 Contratos Ávro** | ✅ Completo | 11 | 0 |
| **🌐 API Gateway** | 🏗️ Parcial | 5 | ~30 |
| **🔧 Demais Serviços (7)** | 🏗️ Scaffold | 28 | ~80 |
| **⏳ Pendentes** | 🏗️ Vazio | 0 | ~25 |
| **Total** | | **~121** | **~150** |

---

## 🚀 Sprint 1 — P0: Fundação Crítica (Semanas 1-3)

**Foco:** Domínios dos serviços scaffold + Gateways BaaS + Kafka real

### Prioridade: 🔴 ALTÍSSIMA

| # | Tarefa | Arquivos | Testes | Risco | Dependências |
|---|--------|:--------:|:------:|:-----:|:------------:|
| 1.1 | **Domínio: boleto-engine** (entity, event, port) | 6 🆕 | 6 🆕 | 🔴 | Nenhuma |
| 1.2 | **Domínio: cashin-engine** (entity, event, port) | 6 🆕 | 6 🆕 | 🔴 | Nenhuma |
| 1.3 | **Domínio: cashout-engine** (entity, event, port) | 6 🆕 | 6 🆕 | 🔴 | Nenhuma |
| 1.4 | **Gateways BaaS reais** (payment-core) — Asaas primeiro | 5 🆕 | 5 🆕 | 🔴 | 1.1-1.3 |
| 1.5 | **Kafka Producer real** (payment-core) | 1 🆕 + 1 ♻️ | 3 🆕 | 🔴 | 1.4 |
| 1.6 | **Adapters outbound** (boleto/provider+postgres, cashin/asaas+postgres, cashout/starkbank+postgres) | 6 🆕 | 6 🆕 | 🔴 | 1.1-1.3 |
| 1.7 | **Consumers Kafka** (payment-core) | 2 🆕 | 2 🆕 | 🔴 | 1.5 |
| 1.8 | **API Gateway: infrastructure/http** (proxies BaaS) | 5 🆕 | 5 🆕 | 🔴 | 1.4 |
| | **Total Sprint 1** | **~38 🆕** | **~12 🆕 + 1 ♻️** | | |

---

## 🚀 Sprint 2 — P1: Core Business + Segurança (Semanas 4-6)

**Foco:** Services + API Gateway modules + Testes de Contrato + Segurança

### Prioridade: 🟠 ALTA

| # | Tarefa | Arquivos | Testes | Risco |
|---|--------|:--------:|:------:|:-----:|
| 2.1 | **Services: boleto-engine** (use cases) | 1 🆕 | 2 🆕 | 🟠 |
| 2.2 | **Services: cashin-engine** (use cases) | 1 🆕 | 2 🆕 | 🟠 |
| 2.3 | **Services: cashout-engine** (use cases) | 1 🆕 | 2 🆕 | 🟠 |
| 2.4 | **Adapters inbound: gRPC+REST** (boleto, cashin, cashout) | 6 🆕 | 6 🆕 | 🟠 |
| 2.5 | **API Gateway: services + DTOs + entities** (todos módulos) | ~15 🆕 | 10 🆕 | 🟠 |
| 2.6 | **API Gateway: infrastructure** (gRPC, cache, database) | 3 🆕 | 3 🆕 | 🟠 |
| 2.7 | **rewards-engine: domínio completo** | 7 🆕 | 6 🆕 | 🟠 |
| 2.8 | **reconciliation: domínio + adapter** | 6 🆕 | 5 🆕 | 🟠 |
| 2.9 | **🧪 Testes de Contrato** (OpenAPI + Avro Schema Registry) | 2 🆕 | 🆕 | 🔴 |
| 2.10 | **🧪 Testes de Segurança** (SQL injection, XSS, webhook signature) | 2 🆕 | 🆕 | 🟠 |
| | **Total Sprint 2** | **~44 🆕** | **~40 🆕** | |

---

## 🚀 Sprint 3 — P2: Suporte Operacional + Resiliência (Semanas 7-9)

**Foco:** Notification, Smart-Retry, Observabilidade, Chaos Engineering

### Prioridade: 🟡 MÉDIA

| # | Tarefa | Arquivos | Testes | Risco |
|---|--------|:--------:|:------:|:-----:|
| 3.1 | **notification: domínio + adapters (email, webhook)** | 5 🆕 | 4 🆕 | 🟡 |
| 3.2 | **smart-retry: domínio + algoritmo retry + ML holder** | 6 🆕 | 5 🆕 | 🟡 |
| 3.3 | **payment-core: infrastructure** (cache, config, logger, metrics) | 5 🆕 | 3 🆕 | 🟡 |
| 3.4 | **payment-core: pkg** (config, crypto, errors, utils) | 4 🆕 | 4 🆕 | 🟡 |
| 3.5 | **Terraform modules** (database, kubernetes, networking) | 3 🆕 | — | 🟡 |
| 3.6 | **Monitoring** (alerts, dashboards, logging) | 3 🆕 | — | 🟡 |
| 3.7 | **🧪 Testes E2E** (3 fluxos críticos) | 3 🆕 | 🆕 | 🔴 |
| 3.8 | **🧪 Testes de Carga (k6)** | 2 🆕 | 🆕 | 🟠 |
| 3.9 | **🧪 Chaos Engineering** (Chaos Mesh) | 1 🆕 | 🆕 | 🟡 |
| | **Total Sprint 3** | **~32 🆕** | **~19 🆕** | |

---

## 🚀 Sprint 4 — P3: Refinamentos (Semanas 10-12)

**Foco:** Infra Cloud, Scripts, Provedores extras

### Prioridade: 🟢 BAIXA

| # | Tarefa | Arquivos | Testes | Risco |
|---|--------|:--------:|:------:|:-----:|
| 4.1 | **K8s overlays** (dev, staging, production) | 3 🆕 | — | 🟢 |
| 4.2 | **Scripts** (deploy, migration, seed) | 3 🆕 | — | 🟢 |
| 4.3 | **shared/baas** (celcoin, fitbank) | 2 🆕 | 2 🆕 | 🟢 |
| 4.4 | **shared/types** | 1 🆕 | — | 🟢 |
| 4.5 | **API Gateway: common** (config, decorators, filters, guards, interceptors, logger, middlewares) | 7 🆕 | 5 🆕 | 🟢 |
| 4.6 | **Testes e2e API Gateway** | 1 🆕 | 🆕 | 🟢 |
| | **Total Sprint 4** | **~17 🆕** | **~8 🆕** | |

---

## 📊 Resumo Macro

| Sprint | Arquivos 🆕 | Testes 🆕 | ♻️ Refatorações | Cobertura Final Estimada |
|:-----:|:-----------:|:---------:|:---------------:|:------------------------:|
| **Sprint 1** | 38 | 12 | 1 | ~55% |
| **Sprint 2** | 44 | 40 | 0 | ~68% |
| **Sprint 3** | 32 | 19 | 0 | ~78% |
| **Sprint 4** | 17 | 8 | 0 | ~85% |
| **Total** | **131 🆕** | **79 🆕** | **1 ♻️** | **🎯 85%** |

---

## 📈 Matriz Riscos x Testes

```
RISCO ATUAL                    TESTES EXISTENTES       AÇÃO              SPRINT
─────────────────────────────────────────────────────────────────────────────────
🔴 Domínios scaffold (3)       ❌ 0%                    Implementar       1
🔴 Gateways BaaS               ❌ Mocks only             Implementar       1
🔴 Kafka/Messaging             ❌ Mock apenas            Implementar       1
🔴 Testes de Contrato          ❌ 0%                     Implementar       2
🔴 Testes de Segurança         ❌ 0%                     Implementar       2
🟠 Services (boleto,cashin,cashout) ❌ 0%               Implementar       1-2
🟠 Multitenancy                ❌ Não implementado       Implementar       3
🟡 Observabilidade             ❌ 0%                     Implementar       3
🟡 Chaos Engineering           ❌ 0%                     Implementar       3
🟡 Infra Cloud (Terraform)     ❌ 0%                     Implementar       3
🟢 Documentação                ✅ 100%                   OK               —
🟢 payment-core domain         ✅ ~85%                   Manter           —
🟢 payment-core application    ✅ ~88%                   Manter           —
🟢 payment-core handlers       ✅ ~100%                  Manter           —
🟢 payment-core postgres       ✅ Testcontainers         Expandir         3
```

---

## 🧠 Decisões Arquiteturais Validadas

| Decisão | Status | Referência |
|---------|--------|------------|
| **E2E limitado a 3-5 fluxos** | ✅ Aprovado | Nubank matou suíte E2E por flakiness |
| **Testes de Contrato (Pact/OpenAPI)** | ✅ Essencial | Estratégia que substitui E2E em fintechs |
| **Sandbox BaaS obrigatório** | ✅ Aprovado | Asaas e Stark Bank documentam sandbox |
| **Testcontainers para integração** | ✅ Padrão | Ferramenta oficial para Go |
| **k6 para carga** | ✅ Recomendado | Escrito em Go, ideal para APIs financeiras |
| **Chaos Engineering** | ✅ Adicionado Sprint 3 | Essencial para sistemas financeiros |
| **ngrok para webhook local** | ✅ Planejado | Asaas recomenda para testes em localhost |

---

## 🚨 Riscos Não Mapeados (Identificados na Pesquisa)

| Risco | Fonte | Mitigação | Sprint |
|-------|-------|-----------|:------:|
| **Flakiness em testes E2E** | Nubank | Manter E2E < 5, rodar apenas em staging | 3 |
| **Subestimar regulação** | Banco Central | Consultoria jurídica na Sprint 2 | 2 |
| **Webhook em localhost** | Asaas | ngrok/Cloudflare Tunnel desde o início | 1 |
| **Dependência de BaaS** | Risco geral | Fallback + múltiplos provedores | 1 |

---

## 🏁 Próximos Passos Imediatos (Sprint 1)

1. **Implementar domínios:** boleto-engine, cashin-engine, cashout-engine
2. **Implementar gateways BaaS reais** (Asaas primeiro)
3. **Implementar Kafka Producer real** no payment-core
4. **Implementar adapters outbound** para cada engine
5. **Implementar consumers Kafka** no payment-core
6. **Implementar proxies BaaS** no API Gateway
7. **Escrever testes** para cada novo componente
