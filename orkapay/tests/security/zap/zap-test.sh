#!/bin/bash
# zap-test.sh - OWASP ZAP Security Scan for OrkaPay
#
# Executa um baseline scan de segurança contra as APIs do OrkaPay.
# Uso: bash tests/security/zap/zap-test.sh

set -euo pipefail

# ---------------------------------------------------------------------------
# Cores
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_failure() { echo -e "${RED}[FAIL]${NC} $1"; }

# ---------------------------------------------------------------------------
# Configuração
# ---------------------------------------------------------------------------
REPORT_DIR="$(dirname "$0")/reports"
ZAP_IMAGE="owasp/zap2docker-stable"
API_URL="http://host.docker.internal:8080"
COMPOSE_FILE="docker-compose.dev.yml"

# ---------------------------------------------------------------------------
# Verifica pré-requisitos
# ---------------------------------------------------------------------------
if ! docker info > /dev/null 2>&1; then
    log_failure "Docker não está rodando"
    exit 1
fi

# ---------------------------------------------------------------------------
# Sobe ambiente
# ---------------------------------------------------------------------------
log_info "Subindo ambiente OrkaPay..."
cd "$(git rev-parse --show-toplevel)" 2>/dev/null || cd "$(dirname "$0")/../../.."

docker compose -f "$COMPOSE_FILE" up -d --build cashin-api cashout-api
sleep 10

# Verifica se a API está respondendo
MAX_RETRIES=15
RETRY=0
while [ $RETRY -lt $MAX_RETRIES ]; do
    if curl -sf "$API_URL/health" > /dev/null 2>&1; then
        break
    fi
    sleep 2
    RETRY=$((RETRY + 1))
done

if [ $RETRY -eq $MAX_RETRIES ]; then
    log_failure "API não respondeu após $MAX_RETRIES tentativas"
    exit 1
fi

log_success "API respondendo em $API_URL"

# ---------------------------------------------------------------------------
# Cria diretório de relatório
# ---------------------------------------------------------------------------
mkdir -p "$REPORT_DIR"
REPORT_FILE="$REPORT_DIR/zap-report-$(date +%Y%m%d-%H%M%S).html"

# ---------------------------------------------------------------------------
# Executa ZAP Baseline Scan
# ---------------------------------------------------------------------------
log_info "Executando OWASP ZAP Baseline Scan..."
log_info "Alvo: $API_URL"

# O ZAP baseline scan é passivo (não modifica dados)
docker run --rm \
    -v "$(pwd)/$REPORT_DIR:/zap/wrk:rw" \
    --add-host host.docker.internal:host-gateway \
    "$ZAP_IMAGE" \
    zap-baseline.py \
    -t "$API_URL" \
    -r "zap-report.html" \
    -j \
    -I

# Move relatório para nome com timestamp
if [ -f "$(pwd)/$REPORT_DIR/zap-report.html" ]; then
    mv "$(pwd)/$REPORT_DIR/zap-report.html" "$(pwd)/$REPORT_FILE"
fi

log_info "Relatório salvo em: $REPORT_FILE"

# ---------------------------------------------------------------------------
# Verifica resultado
# ---------------------------------------------------------------------------
if [ -f "$(pwd)/$REPORT_FILE" ]; then
    # Verifica se há alertas de alta severidade
    if grep -q '"High"' "$(pwd)/$REPORT_FILE" 2>/dev/null; then
        log_failure "⚠️  Alertas de ALTA severidade encontrados!"
        log_info "Consulte o relatório para detalhes: $REPORT_FILE"
        exit 1
    else
        log_success "Nenhum alerta de alta severidade encontrado"
    fi
else
    log_warn "Relatório não encontrado, mas scan pode ter sido executado"
fi

# ---------------------------------------------------------------------------
# Limpeza
# ---------------------------------------------------------------------------
log_info "Limpando ambiente..."
docker compose -f "$COMPOSE_FILE" down -v 2>/dev/null || true

log_success "ZAP Scan concluído!"
exit 0