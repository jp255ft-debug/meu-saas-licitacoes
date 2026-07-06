# Testes de Resiliência - OrkaPay

## Visão Geral

Este documento descreve a suíte de testes de resiliência do OrkaPay, cobrindo simulação de falhas de rede, chaos engineering em containers, testes de carga, e testes de integração com dependências reais.

---

## 1. Matriz de Cenários de Falha

| Cenário de Falha | Ferramenta | Como Simular | Comportamento Esperado | Teste |
|---|---|---|---|---|
| **Banco de dados offline** | Toxiproxy | `down` toxic no proxy PostgreSQL | API retorna 503; logs registram erro | `TestDatabaseConnectionDown_ReturnsError` |
| **Latência alta no banco** | Toxiproxy | `latency: 1500ms` | Timeout é respeitado; conexões não vazam | `TestDatabaseLatency_TimeoutRespected` |
| **Perda de pacotes (80%)** | Toxiproxy | `loss: 0.8` | Retries funcionam; maioria das queries falha | `TestDatabasePacketLoss_RetryLogic` |
| **Slow close** | Toxiproxy | `slow_close: 5000ms` | Conexões normais não são impactadas | `TestDatabaseSlowClose_NoConnectionLeak` |
| **Timeout de conexão** | Toxiproxy | `timeout: 0` | Conexão é recusada; erro tratado | `TestDatabaseTimeout_ConnectionRefused` |
| **Reset de peer (TCP RST)** | Toxiproxy | `reset_peer` | Conexão é resetada; erro tratado | `TestDatabaseResetPeer_ConnectionBroken` |
| **Recuperação pós-falha** | Toxiproxy | Remover toxic → reconectar | Sistema recupera automaticamente | `TestDatabaseRecovery_AfterToxicRemoval` |
| **Container da API morre** | Pumba | `kill --signal SIGTERM` | Health check falha; orquestrador reinicia | `test_kill_cashin_api` |
| **Container do banco morre** | Pumba | `kill --signal SIGKILL` | API fica indisponível; banco reinicia | `test_kill_cashin_db` |
| **API pausada por 15s** | Pumba | `stop --duration 15s` | API fica offline; recupera após pausa | `test_stop_cashout_api` |
| **Latência de rede (500ms)** | Pumba | `netem --delay 500ms` | API degrada mas continua respondendo | `test_network_latency_cashin` |
| **Estresse de CPU (70%)** | Pumba | `stress --cpu-load 70` | API pode degradar mas não crasha | `test_cpu_stress_cashout` |
| **Pico de requisições (50 usuarios)** | k6 | Ramp-up 10→20→50 req/s | p95 < 500ms; erro rate < 1% | `k6-load-test.js` |
| **Autenticação falha** | Teste unitário | API Key inválida | 401 Unauthorized | Testes de handler existentes |
| **Optimistic Locking** | Testcontainers | Duas requisições simultâneas | Uma falha com conflict; dados consistentes | `TestPaymentRepository_Update_OptimisticLock` |

---

## 2. Ferramentas

### 2.1 Toxiproxy (Simulação de Falhas de Rede)

**Propósito:** Proxy TCP que permite simular condições de rede adversárias.

**Setup automático via Testcontainers:**
- Container PostgreSQL + Container Toxiproxy
- Proxy configurado entre a aplicação e o banco
- Toxics injetados via API HTTP (`localhost:8474`)

**Arquivo:** `tests/toxiproxy/toxiproxy_test.go`

**Build tag:** `//go:build toxiproxy`

**Execução:**
```bash
go test ./tests/toxiproxy/... -tags=toxiproxy -v -timeout 300s
```

**Toxics disponíveis:** `latency`, `down`, `loss`, `slow_close`, `timeout`, `reset_peer`

### 2.2 Pumba (Chaos Engineering em Containers)

**Propósito:** Ferramenta de chaos engineering para containers Docker.

**Setup:**
```bash
docker pull gaiaadm/pumba:latest
```

**Arquivo:** `tests/chaos/chaos_test.sh`

**Execução:**
```bash
bash tests/chaos/chaos_test.sh
```

**Comandos de chaos:**
| Comando | Efeito |
|---------|--------|
| `pumba kill --signal SIGTERM <container>` | Mata container graciosamente |
| `pumba kill --signal SIGKILL <container>` | Mata container forçadamente |
| `pumba stop --duration 15s <container>` | Pausa container por 15s |
| `pumba netem --delay 500ms <container>` | Adiciona latência de rede |
| `pumba stress --cpu 2 --cpu-load 70 <container>` | Estresse de CPU |

### 2.3 k6 (Testes de Carga)

**Propósito:** Testes de carga e performance para APIs.

**Setup:**
```bash
# Windows (choco)
choco install k6

# Linux/Mac
brew install k6

# Go
go install go.k6.io/k6@latest
```

**Arquivo:** `tests/load/k6-load-test.js`

**Execução:**
```bash
# Executar testes
k6 run tests/load/k6-load-test.js

# Com relatório JSON
k6 run --out json=results.json tests/load/k6-load-test.js

# Usando URLs customizadas
CASHIN_URL=http://meu-host:8080 CASHOUT_URL=http://meu-host:8081 k6 run tests/load/k6-load-test.js
```

**Thresholds:**
- p(95) das requisições < 500ms
- Taxa de erro < 1%
- Erros customizados < 5%

### 2.4 Testcontainers (Testes de Integração)

**Propósito:** Testes de integração com dependências reais em containers efêmeros.

**Arquivos:**
- `services/cashin-engine/internal/adapter/outbound/postgres/repository_test.go`
- `services/cashout-engine/internal/adapter/outbound/postgres/repository_test.go`
- `services/payment-core/internal/adapter/outbound/postgres/payment_repository_integration_test.go`

**Build tag:** `//go:build integration`

**Execução:**
```bash
go test ./... -tags=integration -count=1 -v -timeout 120s
```

---

## 3. Pipeline CI/CD

### GitHub Actions Workflow

**Arquivo:** `.github/workflows/test-suite.yml`

**Jobs em paralelo:**
```
unit-tests → integration-tests
          → toxiproxy-tests
          → load-tests
          → chaos-tests
```

**Execução manual:**
```bash
bash scripts/test-suite.sh              # Todos os testes
bash scripts/test-suite.sh --unit       # Apenas unitários
bash scripts/test-suite.sh --skip-chaos # Todos exceto chaos
```

---

## 4. Comportamento Esperado do Sistema

### 4.1 Falhas de Banco de Dados

| Situação | Resposta HTTP | Log | Recuperação |
|----------|--------------|-----|-------------|
| Banco offline | 503 Service Unavailable | `ERRO: banco de dados indisponível` | Automática quando banco voltar |
| Latência alta | 503/Timeout | `AVISO: operação excedeu timeout` | Retry com backoff |
| Query lenta | 200 (lento) | `AVISO: query lenta (>1s)` | N/A (degradação) |
| Conexão perdida | 500/503 | `ERRO: conexão perdida` | Pool de conexões reconecta |

### 4.2 Falhas de Containers

| Situação | Observado | Recuperação |
|----------|-----------|-------------|
| API morta | Health check falha | Docker reinicia (restart: unless-stopped) |
| Banco morto | API retorna 503 | Docker reinicia; API reconecta |
| API pausada | Requisições falham | Docker retoma após pausa |

### 4.3 Sobrecarga

| Situação | Observado | Ação |
|----------|-----------|------|
| 50 req/s sustentado | p95 < 500ms | Normal |
| > 100 req/s | Latência aumenta | Rate limiting atua (se configurado) |
| Pico repentino | Algumas req falham | Retry com backoff |

---

## 5. Runbooks de Recuperação

### 5.1 Banco de Dados Indisponível

1. Verificar status: `docker compose ps`
2. Verificar logs: `docker compose logs cashin-db`
3. Se container parou: `docker compose restart cashin-db`
4. Aguardar health check: `docker compose ps cashin-db`
5. Verificar API: `curl http://localhost:8080/health`

### 5.2 API Indisponível

1. Verificar logs: `docker compose logs cashin-api`
2. Se erro de configuração:
   - Corrigir variáveis de ambiente
   - `docker compose up -d --build cashin-api`
3. Se OOM (Out of Memory):
   - Aumentar limites de memória no `docker-compose.yml`
   - `docker compose up -d cashin-api`

### 5.3 Latência Alta

1. Identificar gargalo:
   ```bash
   docker stats  # Verificar CPU/memória
   docker compose logs cashin-db | grep -i slow
   ```
2. Se banco lento: verificar índices e queries
3. Se API lenta: verificar conexões e goroutines

---

## 6. Executando Testes Localmente

### Pré-requisitos

- Docker Desktop (ou equivalente)
- Go 1.21+
- k6 CLI (para testes de carga)
- Conexão com internet (para baixar imagens Docker)

### Passo a Passo

```bash
# 1. Testes unitários (rápidos, sem Docker)
cd services/cashin-engine
go test ./... -short -count=1

# 2. Testes de integração (requer Docker)
go test ./... -tags=integration -count=1 -v -timeout 120s

# 3. Testes Toxiproxy (requer Docker, ~5 min)
cd ../../tests/toxiproxy
go test ./... -tags=toxiproxy -count=1 -v -timeout 300s

# 4. Testes de carga (requer Docker + k6)
cd ../..
k6 run tests/load/k6-load-test.js

# 5. Chaos tests (requer Docker + Pumba)
bash tests/chaos/chaos_test.sh

# 6. Tudo de uma vez
bash scripts/test-suite.sh --skip-chaos
```

---

## 7. Estrutura de Arquivos

```
orkapay/
├── tests/
│   ├── toxiproxy/
│   │   ├── go.mod
│   │   └── toxiproxy_test.go          # 6 cenários de falha de rede
│   ├── chaos/
│   │   └── chaos_test.sh              # 5 cenários de chaos em containers
│   └── load/
│       └── k6-load-test.js            # Teste de carga com 4 endpoints
├── scripts/
│   └── test-suite.sh                  # Script unificado de execução
├── .github/
│   └── workflows/
│       └── test-suite.yml             # CI/CD com 5 jobs paralelos
├── docs/
│   └── testes-resiliencia.md          # Este documento
└── services/
    ├── cashin-engine/
    │   └── .../repository_test.go     # Testcontainers (já existente)
    ├── cashout-engine/
    │   └── .../repository_test.go     # Testcontainers (já existente)
    └── payment-core/
        └── .../payment_repository_integration_test.go  # NOVO: Testcontainers
```

---

## 8. Métricas de Sucesso

| Métrica | Alvo | Ferramenta |
|---------|------|-----------|
| Cobertura de domínio | 100% | go test -cover |
| Cobertura de serviços | 90%+ | go test -cover |
| Cobertura de adaptadores | 80%+ | go test -cover |
| Tempo de resposta (p95) | < 500ms | k6 |
| Taxa de erro em carga | < 1% | k6 |
| Recuperação de container | < 30s | Pumba + health check |
| Cenários de caos | 100% pass | Pumba script |
| Cenários de rede | 100% pass | Toxiproxy |