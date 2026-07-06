# Observabilidade - OrkaPay

## Visão Geral

Este documento descreve a stack de observabilidade do OrkaPay, que inclui rastreamento distribuído (tracing), métricas e dashboards.

---

## Stack

| Componente | Propósito | Porta | Acesso |
|-----------|-----------|-------|--------|
| **Jaeger** | Rastreamento distribuído (tracing) | `16686` | http://localhost:16686 |
| **Prometheus** | Coleta de métricas | `9090` | http://localhost:9090 |
| **Grafana** | Dashboards e visualização | `3000` | http://localhost:3000 |

---

## Como Subir

### Ambiente completo (APIs + Observabilidade)

```bash
# Cria rede compartilhada
docker network create orkapay-network 2>/dev/null || true

# Sobe tudo
docker compose -f docker-compose.dev.yml -f docker-compose.observability.yml up -d
```

### Apenas observabilidade

```bash
docker compose -f docker-compose.observability.yml up -d
```

---

## Jaeger - Rastreamento Distribuído

### Acessar UI

Abra http://localhost:16686

### Como funciona

O middleware de tracing nos serviços captura:

- **Span por requisição HTTP** (método, URL, status code, duração)
- **Propagação de contexto** entre serviços via headers
- **Atributos customizados** como client_id, operação, etc.

### Para ativar o tracing real

1. Adicione as dependências ao `go.mod`:
   ```bash
   go get go.opentelemetry.io/otel \
          go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp \
          go.opentelemetry.io/otel/sdk
   ```

2. Descomente o código no middleware `tracing.go` em ambos os serviços

3. Adicione a variável de ambiente nos serviços:
   ```yaml
   OTEL_EXPORTER_OTLP_ENDPOINT: http://jaeger:4318
   ```

### Exemplo de busca no Jaeger

1. Abra http://localhost:16686
2. Selecione o serviço: `orkapay/cashin` ou `orkapay/cashout`
3. Clique em "Find Traces"
4. Explore os spans de cada requisição

---

## Prometheus - Métricas

### Acessar

http://localhost:9090

### Targets configurados

| Job | Target | Métricas |
|-----|--------|----------|
| `orkapay-cashin` | `cashin-api:8080/metrics` | Goroutines, memória, requisições |
| `orkapay-cashout` | `cashout-api:8080/metrics` | Goroutines, memória, requisições |

### Queries úteis

```promql
# Taxa de requisições por segundo
rate(http_requests_total[5m])

# Latência média (se exposta)
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

# Uso de memória
go_memstats_alloc_bytes

# Goroutines ativas
go_goroutines
```

---

## Grafana - Dashboards

### Acessar

http://localhost:3000
- **Usuário:** `admin`
- **Senha:** `admin`

### Configurar datasource Prometheus

1. Vá em Configuration → Data Sources
2. Add data source → Prometheus
3. URL: `http://prometheus:9090`
4. Save & Test

### Dashboard sugerido - "OrkaPay - Resiliência"

Crie um dashboard com os seguintes painéis:

| Painel | Query | Descrição |
|--------|-------|-----------|
| **Taxa de Erro** | `rate(http_requests_total{status=~"5.."}[5m])` | Erros 5xx nos serviços |
| **Latência p95** | `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))` | Latência no percentil 95 |
| **Goroutines** | `go_goroutines{job=~"orkapay-.*"}` | Goroutines ativas por serviço |
| **Memória** | `go_memstats_alloc_bytes{job=~"orkapay-.*"}` | Alocação de memória |
| **Requisições/s** | `rate(http_requests_total[5m])` | Throughput |
| **Containers** | Contagem de containers rodando | (via docker_exporter ou similar) |

---

## Integração com Chaos Tests

Durante os testes de chaos (Pumba), use a stack de observabilidade para:

1. **Antes do caos:** Capture baseline de métricas (latência, taxa de erro)
2. **Durante o caos:** Observe o impacto em tempo real nos dashboards
3. **Após o caos:** Verifique se as métricas retornaram ao baseline

### Comandos úteis

```bash
# Subir ambiente completo: APIs + observabilidade
docker compose -f docker-compose.dev.yml -f docker-compose.observability.yml up -d

# Executar chaos tests enquanto monitora
bash tests/chaos/chaos_test.sh

# Ver traces no Jaeger durante o caos
open http://localhost:16686

# Ver métricas no Grafana durante o caos
open http://localhost:3000
```

---

## Arquivos

| Arquivo | Descrição |
|---------|-----------|
| `docker-compose.observability.yml` | Jaeger + Prometheus + Grafana |
| `prometheus.yml` | Configuração do Prometheus |
| `services/cashin-engine/.../middleware/tracing.go` | Middleware de tracing (cashin) |
| `services/cashout-engine/.../middleware/tracing.go` | Middleware de tracing (cashout) |

---

## Troubleshooting

### Jaeger não mostra traces

1. Verifique se o Jaeger está rodando: `docker ps | grep jaeger`
2. Verifique as variáveis de ambiente:
   ```bash
   docker compose logs cashin-api | grep OTEL
   ```
3. O middleware de tracing atual é um **placeholder**. Para ativar o tracing real, descomente o código e adicione as dependências OpenTelemetry.

### Prometheus não coleta métricas

1. Verifique se os serviços expõem `/metrics`:
   ```bash
   curl http://localhost:8080/metrics
   ```
2. Verifique o target no Prometheus:
   http://localhost:9090/targets

### Grafana não conecta ao Prometheus

1. Verifique se o datasource está configurado corretamente
2. O Prometheus deve ser acessível via nome do container: `http://prometheus:9090`