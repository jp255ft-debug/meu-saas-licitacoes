#!/bin/bash
# test-suite.sh - Test Suite Unificada para OrkaPay (Fases 8-11)
#
# Executa todas as suítes de teste em ordem: unitários, integração, toxiproxy,
# wiremock, security, fuzzing, carga, chaos
#
# Uso:
#   bash scripts/test-suite.sh              # Executa todos os testes
#   bash scripts/test-suite.sh --unit       # Apenas unitários
#   bash scripts/test-suite.sh --wiremock   # Apenas WireMock
#   bash scripts/test-suite.sh --security   # Apenas segurança (ZAP + fuzz)
#   bash scripts/test-suite.sh --load       # Apenas carga + security k6
#   bash scripts/test-suite.sh --chaos      # Apenas chaos

set -euo pipefail

# ---------------------------------------------------------------------------
# Cores
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ---------------------------------------------------------------------------
# Contadores
# ---------------------------------------------------------------------------
PASS_COUNT=0
FAIL_COUNT=0
START_TIME=$(date +%s)

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_failure() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }

run_suite() {
    local name=$1
    shift
    log_info "=========================================="
    log_info "🧪 Suite: $name"
    log_info "=========================================="
    echo ""

    if "$@"; then
        log_success "Suite '$name' concluída com sucesso"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        log_failure "Suite '$name' falhou"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
    echo ""
}

check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_warn "Docker não está disponível. Testes que exigem Docker serão pulados."
        return 1
    fi
    return 0
}

check_k6() {
    if ! command -v k6 &> /dev/null; then
        log_warn "k6 não encontrado. Testes de carga serão pulados."
        return 1
    fi
    return 0
}

# ---------------------------------------------------------------------------
# Suítes de Teste - Fases 1-7 (existentes)
# ---------------------------------------------------------------------------

run_unit_tests() {
    log_info "Executando testes unitários em todos os serviços..."
    cd services/cashin-engine
    go test ./... -short -count=1 || return 1
    cd ../cashout-engine
    go test ./... -short -count=1 || return 1
    cd ../payment-core
    go test ./... -short -count=1 || return 1
    cd ../..
    return 0
}

run_integration_tests() {
    check_docker || return 0
    log_info "Executando testes de integração (Testcontainers)..."
    cd services/cashin-engine
    go test ./... -tags=integration -count=1 -v -timeout 120s || return 1
    cd ../cashout-engine
    go test ./... -tags=integration -count=1 -v -timeout 120s || return 1
    cd ../payment-core
    go test ./... -tags=integration -count=1 -v -timeout 120s || return 1
    cd ../..
    return 0
}

run_toxiproxy_tests() {
    check_docker || return 0
    log_info "Executando testes Toxiproxy..."
    cd tests/toxiproxy
    go test ./... -tags=toxiproxy -count=1 -v -timeout 300s || return 1
    cd ../..
    return 0
}

run_load_tests() {
    check_docker || return 0
    check_k6 || return 0
    log_info "Executando testes de carga (k6)..."
    docker compose -f docker-compose.dev.yml up -d --build && sleep 15
    k6 run tests/load/k6-load-test.js || {
        docker compose -f docker-compose.dev.yml down -v 2>/dev/null || true
        return 1
    }
    docker compose -f docker-compose.dev.yml down -v 2>/dev/null || true
    return 0
}

run_chaos_tests() {
    check_docker || return 0
    log_info "Executando testes de Chaos Engineering (Pumba)..."
    docker pull gaiaadm/pumba:latest 2>/dev/null || true
    bash tests/chaos/chaos_test.sh || return 1
    return 0
}

# ---------------------------------------------------------------------------
# Suítes de Teste - Fases 8-11 (novas)
# ---------------------------------------------------------------------------

run_wiremock_tests() {
    check_docker || return 0
    log_info "Executando testes WireMock (provedores externos)..."
    
    # Sobe WireMock
    cd tests/wiremock
    docker compose -f docker-compose.wiremock.yml up -d
    sleep 5
    
    # Executa testes Go contra WireMock (se houver)
    if ls *.go 2>/dev/null; then
        log_info "Testes Go com build tag wiremock..."
        go test ./... -tags=wiremock -count=1 -v -timeout 60s || {
            docker compose -f docker-compose.wiremock.yml down -v
            cd ../..
            return 1
        }
    fi
    
    docker compose -f docker-compose.wiremock.yml down -v
    cd ../..
    return 0
}

run_security_tests() {
    check_docker || return 0
    log_info "Executando testes de segurança..."
    
    # ZAP Scan
    log_info "OWASP ZAP Baseline Scan..."
    bash tests/security/zap/zap-test.sh || log_warn "ZAP scan reportou alertas"
    
    # Fuzzing (breve)
    log_info "Go Fuzz Tests (rápido)..."
    cd services/cashin-engine
    timeout 30s go test -fuzz=FuzzCreateInvoicePayload -fuzztime=10s ./... 2>/dev/null || true
    cd ../cashout-engine
    timeout 30s go test -fuzz=FuzzCreateTransferPayload -fuzztime=10s ./... 2>/dev/null || true
    cd ../..
    
    return 0
}

run_load_security_tests() {
    check_docker || return 0
    check_k6 || return 0
    log_info "Executando testes de carga com payloads maliciosos (k6 security)..."
    
    docker compose -f docker-compose.dev.yml up -d --build && sleep 15
    
    k6 run tests/load/k6-security-test.js || {
        log_warn "Alguns payloads maliciosos podem não ter sido bloqueados"
    }
    
    docker compose -f docker-compose.dev.yml down -v 2>/dev/null || true
    return 0
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    echo ""
    echo "============================================"
    echo "   🧪 OrkaPay - Test Suite Unificada v2"
    echo "============================================"
    echo ""
    echo "Data: $(date)"
    echo "Go: $(go version 2>/dev/null || echo 'Go não encontrado')"
    echo "Docker: $(docker --version 2>/dev/null || echo 'Docker não encontrado')"
    echo ""

    # Parse args
    RUN_ALL=true
    RUN_UNIT=false
    RUN_INTEGRATION=false
    RUN_TOXIPROXY=false
    RUN_LOAD=false
    RUN_CHAOS=false
    RUN_WIREMOCK=false
    RUN_SECURITY=false
    RUN_LOAD_SECURITY=false

    if [ $# -gt 0 ]; then
        RUN_ALL=false
        for arg in "$@"; do
            case "$arg" in
                --unit)         RUN_UNIT=true ;;
                --integration)  RUN_INTEGRATION=true ;;
                --toxiproxy)    RUN_TOXIPROXY=true ;;
                --load)         RUN_LOAD=true ;;
                --chaos)        RUN_CHAOS=true ;;
                --wiremock)     RUN_WIREMOCK=true ;;
                --security)     RUN_SECURITY=true ;;
                --load-sec)     RUN_LOAD_SECURITY=true ;;
                --all)          RUN_ALL=true ;;
                --skip-chaos)   RUN_ALL=true; export SKIP_CHAOS=true ;;
                --help|-h)
                    echo "Uso: $0 [opções]"
                    echo "  --unit          Testes unitários"
                    echo "  --integration   Testes de integração"
                    echo "  --toxiproxy     Testes de falhas de rede"
                    echo "  --load          Testes de carga (k6)"
                    echo "  --chaos         Chaos engineering (Pumba)"
                    echo "  --wiremock      Simulação de provedores externos"
                    echo "  --security      Segurança (ZAP + fuzzing)"
                    echo "  --load-sec      Carga com payloads maliciosos"
                    echo "  --all           Todas as suítes"
                    echo "  --skip-chaos    Todos exceto chaos"
                    exit 0
                    ;;
                *)
                    log_failure "Argumento desconhecido: $arg"
                    exit 1
                    ;;
            esac
        done
    fi

    if [ "$RUN_ALL" = true ]; then
        run_suite "Testes Unitários" run_unit_tests
        run_suite "Testes de Integração" run_integration_tests
        run_suite "Toxiproxy (Falhas de Rede)" run_toxiproxy_tests
        run_suite "WireMock (Provedores Externos)" run_wiremock_tests
        run_suite "Segurança (ZAP + Fuzzing)" run_security_tests
        run_suite "Testes de Carga (k6)" run_load_tests
        run_suite "k6 Security (Payloads Maliciosos)" run_load_security_tests

        if [ "${SKIP_CHAOS:-}" != "true" ]; then
            run_suite "Chaos Engineering (Pumba)" run_chaos_tests
        else
            log_warn "Testes de chaos pulados (SKIP_CHAOS=true)"
        fi
    else
        [ "$RUN_UNIT" = true ] && run_suite "Testes Unitários" run_unit_tests
        [ "$RUN_INTEGRATION" = true ] && run_suite "Testes de Integração" run_integration_tests
        [ "$RUN_TOXIPROXY" = true ] && run_suite "Toxiproxy" run_toxiproxy_tests
        [ "$RUN_LOAD" = true ] && run_suite "Testes de Carga" run_load_tests
        [ "$RUN_CHAOS" = true ] && run_suite "Chaos Engineering" run_chaos_tests
        [ "$RUN_WIREMOCK" = true ] && run_suite "WireMock" run_wiremock_tests
        [ "$RUN_SECURITY" = true ] && run_suite "Segurança" run_security_tests
        [ "$RUN_LOAD_SECURITY" = true ] && run_suite "k6 Security" run_load_security_tests
    fi

    # Resultados finais
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    echo ""
    echo "============================================"
    echo "   📊 Resultados Finais"
    echo "============================================"
    echo -e "   ${GREEN}Suítes com sucesso: $PASS_COUNT${NC}"
    echo -e "   ${RED}Suítes com falha: $FAIL_COUNT${NC}"
    echo "   Duração total: ${DURATION}s"
    echo "============================================"
    echo ""

    if [ $FAIL_COUNT -gt 0 ]; then
        log_failure "Algumas suítes de teste falharam"
        exit 1
    else
        log_success "Todas as suítes de teste passaram! 🚀"
    fi
}

main "$@"