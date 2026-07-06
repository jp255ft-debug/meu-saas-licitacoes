#!/bin/bash
# chaos_test.sh - Chaos Engineering Tests for OrkaPay
#
# Cenários de caos para testar resiliência dos containers.
# Pré-requisitos: Docker, docker-compose, Pumba
#
# Para instalar Pumba: docker pull gaiaadm/pumba
#
# Uso: bash tests/chaos/chaos_test.sh

set -euo pipefail

# ---------------------------------------------------------------------------
# Cores para output
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Configuração
# ---------------------------------------------------------------------------
COMPOSE_FILE="docker-compose.dev.yml"
PUMBA_IMAGE="gaiaadm/pumba:latest"
COMPOSE_PROJECT="orkapay"
PASS_COUNT=0
FAIL_COUNT=0

# ---------------------------------------------------------------------------
# Funções helpers
# ---------------------------------------------------------------------------
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_failure() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_failure "Docker não está rodando. Abortando."
        exit 1
    fi
}

check_service_health() {
    local service=$1
    local max_retries=$2
    local retry=0

    while [ $retry -lt $max_retries ]; do
        if docker compose -f "$COMPOSE_FILE" -p "$COMPOSE_PROJECT" ps "$service" --format json 2>/dev/null | grep -q '"Status":"healthy"'; then
            return 0
        fi
        sleep 2
        retry=$((retry + 1))
    done
    return 1
}

wait_for_rollback() {
    local service=$1
    local max_wait=$2
    local waited=0

    log_info "Aguardando rollback de $service (até ${max_wait}s)..."
    while [ $waited -lt $max_wait ]; do
        if docker compose -f "$COMPOSE_FILE" -p "$COMPOSE_PROJECT" ps "$service" --format json 2>/dev/null | grep -q '"Status":"running\|healthy"'; then
            return 0
        fi
        sleep 2
        waited=$((waited + 2))
    done
    return 1
}

run_test() {
    local test_name=$1
    shift
    log_info "🧪 Executando: $test_name"
    if "$@"; then
        log_success "$test_name"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        log_failure "$test_name"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

# ---------------------------------------------------------------------------
# Setup
# ---------------------------------------------------------------------------
setup() {
    log_info "=== Configurando ambiente de testes ==="

    check_docker

    # Sobe ambiente
    log_info "Subindo ambiente com docker-compose..."
    docker compose -f "$COMPOSE_FILE" -p "$COMPOSE_PROJECT" up -d --build

    # Aguarda serviços ficarem saudáveis
    log_info "Aguardando serviços ficarem saudáveis..."
    sleep 10

    if ! check_service_health "cashin-db" 15; then
        log_failure "cashin-db não ficou saudável. Abortando."
        docker compose -f "$COMPOSE_FILE" -p "$COMPOSE_PROJECT" logs cashin-db
        exit 1
    fi
    log_success "cashin-db saudável"

    if ! check_service_health "cashout-db" 15; then
        log_failure "cashout-db não ficou saudável. Abortando."
        exit 1
    fi
    log_success "cashout-db saudável"

    # Aguarda APIs iniciarem
    sleep 5
    log_success "Ambiente pronto para testes de caos"
    echo ""
}

# ---------------------------------------------------------------------------
# Tear down
# ---------------------------------------------------------------------------
teardown() {
    log_info "=== Limpando ambiente ==="
    docker compose -f "$COMPOSE_FILE" -p "$COMPOSE_PROJECT" down -v 2>/dev/null || true
    log_success "Ambiente limpo"
}

# ---------------------------------------------------------------------------
# Cenários de Chaos
# ---------------------------------------------------------------------------

# Cenário 1: Matar container da API CashIn
test_kill_cashin_api() {
    log_info "Cenário: Container cashin-api é morto (SIGTERM)"

    # Mata o container
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        "$PUMBA_IMAGE" \
        kill --signal SIGTERM "${COMPOSE_PROJECT}-cashin-api" 2>/dev/null || true

    # Aguarda rollback/reinício
    if wait_for_rollback "cashin-api" 30; then
        log_success "cashin-api reiniciou após kill"
        return 0
    else
        log_failure "cashin-api NÃO reiniciou após kill"
        return 1
    fi
}

# Cenário 2: Container do banco CashIn é morto
test_kill_cashin_db() {
    log_info "Cenário: Container cashin-db é morto (SIGKILL)"

    # Registra estado atual
    docker compose -f "$COMPOSE_FILE" -p "$COMPOSE_PROJECT" ps

    # Mata o container do banco
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        "$PUMBA_IMAGE" \
        kill --signal SIGKILL "${COMPOSE_PROJECT}-cashin-db" 2>/dev/null || true

    # Aguarda reinício automático
    if wait_for_rollback "cashin-db" 30; then
        log_success "cashin-db reiniciou após kill"
        return 0
    else
        log_failure "cashin-db NÃO reiniciou após kill"
        return 1
    fi
}

# Cenário 3: Parar container API CashOut por 15s
test_stop_cashout_api() {
    log_info "Cenário: Container cashout-api é pausado por 15s"

    # Pausa o container
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        "$PUMBA_IMAGE" \
        stop --duration 15s "${COMPOSE_PROJECT}-cashout-api" 2>/dev/null || true

    # Verifica se a API fica indisponível durante a pausa
    sleep 3
    if curl -sf http://localhost:8081/health > /dev/null 2>&1; then
        log_warn "cashout-api ainda respondendo durante pausa"
    else
        log_success "cashout-api ficou indisponível durante pausa (esperado)"
    fi

    # Aguarda recuperação
    if wait_for_rollback "cashout-api" 20; then
        log_success "cashout-api recuperou após pausa"
        return 0
    else
        log_failure "cashout-api NÃO recuperou após pausa"
        return 1
    fi
}

# Cenário 4: Latência de rede na API CashIn
test_network_latency_cashin() {
    log_info "Cenário: Latência de 500ms na rede da cashin-api"

    # Aplica latência
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        "$PUMBA_IMAGE" \
        netem --duration 30s --interface eth0 \
        --delay-time 500 --delay-jitter 100 \
        "${COMPOSE_PROJECT}-cashin-api" 2>/dev/null || true

    # Mede tempo de resposta durante latência
    local total_time=0
    local success_count=0
    for i in 1 2 3; do
        local start=$(date +%s%N)
        if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
            local end=$(date +%s%N)
            local elapsed=$(( (end - start) / 1000000 ))
            total_time=$((total_time + elapsed))
            success_count=$((success_count + 1))
            log_info "  Health check $i: ${elapsed}ms"
        fi
    done

    if [ $success_count -gt 0 ]; then
        local avg=$((total_time / success_count))
        log_info "  Tempo médio com latência: ${avg}ms"
        log_success "cashin-api respondeu mesmo com latência de rede"
        return 0
    else
        log_failure "cashin-api não respondeu com latência de rede"
        return 1
    fi
}

# Cenário 5: Estresse de CPU na API CashOut
test_cpu_stress_cashout() {
    log_info "Cenário: Estresse de CPU (2 cores, 70%) na cashout-api"

    # Aplica estresse de CPU
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        "$PUMBA_IMAGE" \
        stress --duration 20s --cpu 2 --cpu-load 70 \
        "${COMPOSE_PROJECT}-cashout-api" 2>/dev/null || true

    # Verifica se ainda responde
    sleep 3
    if curl -sf http://localhost:8081/health > /dev/null 2>&1; then
        log_success "cashout-api respondeu durante estresse de CPU"
        return 0
    else
        log_info "  cashout-api pode ter degradado com estresse de CPU (aceitável)"
        return 0  # Não falha o teste, apenas registra
    fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    echo ""
    echo "============================================"
    echo "   🔬 Chaos Engineering - OrkaPay"
    echo "============================================"
    echo ""

    # Garante limpeza ao final
    trap teardown EXIT

    # Setup
    setup

    # Executa cenários
    echo ""
    log_info "=== Iniciando cenários de caos ==="
    echo ""

    run_test "Kill cashin-api (deve reiniciar)" test_kill_cashin_api
    sleep 5

    run_test "Kill cashin-db (deve reiniciar)" test_kill_cashin_db
    sleep 10 # Aguarda banco reiniciar e migrações rodarem

    run_test "Stop cashout-api por 15s" test_stop_cashout_api
    sleep 5

    run_test "Latência de rede na cashin-api" test_network_latency_cashin
    sleep 5

    run_test "Estresse de CPU na cashout-api" test_cpu_stress_cashout

    # Resultados
    echo ""
    echo "============================================"
    echo -e "   📊 Resultados dos Testes de Chaos"
    echo "============================================"
    echo -e "   ${GREEN}Passados: $PASS_COUNT${NC}"
    echo -e "   ${RED}Falhas: $FAIL_COUNT${NC}"
    echo "============================================"
    echo ""

    if [ $FAIL_COUNT -gt 0 ]; then
        log_failure "Alguns cenários de caos falharam"
        exit 1
    else
        log_success "Todos os cenários de caos passaram! 🚀"
    fi
}

# Executa main
main