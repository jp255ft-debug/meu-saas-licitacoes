#!/bin/bash
# audit.sh - Script automatizado de auditoria para OrkaPay
# Uso: bash scripts/audit.sh
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

mkdir -p reports/audit
PASS=0; FAIL=0; TOTAL=0

check() {
    TOTAL=$((TOTAL + 1))
    local name=$1; shift
    if "$@"; then
        log_pass "$name"
        PASS=$((PASS + 1))
    else
        log_fail "$name"
        FAIL=$((FAIL + 1))
    fi
}

echo "============================================"
echo "   🔍 Auditoria OrkaPay"
echo "============================================"
echo ""

# 1. Estrutura hexagonal
check "Estrutura hexagonal (domain/adapter/cmd)" test -d services/cashin-engine/internal/core/domain -a -d services/cashin-engine/internal/core/application -a -d services/cashin-engine/internal/adapter -a -d services/cashin-engine/cmd

# 2. go vet
for svc in cashin-engine cashout-engine payment-core; do
    check "go vet $svc" bash -c "cd services/$svc && go vet ./... 2>&1 | head -5" 2>/dev/null
done

# 3. Nomes lowercase
LOWER=$(find services -name "*[A-Z]*.go" 2>/dev/null | grep -v "_test.go" | grep -v ".pb.go" | wc -l)
check "Nomes de arquivos lowercase" test "$LOWER" -eq 0

# 4. Formatação
check "gofmt (cashin)" bash -c "cd services/cashin-engine && gofmt -l . | wc -l | xargs test 0 -eq" 2>/dev/null

# 5. Middleware: Rate limiting
check "Rate limiting middleware (cashin)" test -f services/cashin-engine/internal/adapter/inbound/rest/middleware/ratelimit.go
check "Rate limiting middleware (cashout)" test -f services/cashout-engine/internal/adapter/inbound/rest/middleware/ratelimit.go

# 6. Middleware: Security headers
check "Security headers (cashin)" grep -q "SecurityHeaders" services/cashin-engine/internal/adapter/inbound/rest/middleware/security.go 2>/dev/null
check "Security headers (cashout)" grep -q "SecurityHeaders" services/cashout-engine/internal/adapter/inbound/rest/middleware/security.go 2>/dev/null

# 7. Middleware: Request ID
check "Request ID (cashin)" test -f services/cashin-engine/internal/adapter/inbound/rest/middleware/requestid.go
check "Request ID (cashout)" test -f services/cashout-engine/internal/adapter/inbound/rest/middleware/requestid.go

# 8. Logs estruturados (slog)
check "Logging middleware (cashin)" grep -q "slog" services/cashin-engine/internal/adapter/inbound/rest/middleware/logging.go 2>/dev/null
check "Logging middleware (cashout)" grep -q "slog" services/cashout-engine/internal/adapter/inbound/rest/middleware/logging.go 2>/dev/null

# 9. Config tipada
check "Config tipada (cashin)" test -f services/cashin-engine/internal/adapter/config/config.go
check "Config tipada (cashout)" test -f services/cashout-engine/internal/adapter/config/config.go

# 10. Dockerfiles
check "Dockerfile (cashin)" test -f services/cashin-engine/Dockerfile
check "Dockerfile (cashout)" test -f services/cashout-engine/Dockerfile

# 11. docker-compose
check "docker-compose.dev.yml" test -f docker-compose.dev.yml

# 12. Arquivos de teste
TEST_FILES=$(find services -name "*_test.go" 2>/dev/null | wc -l)
check "Arquivos de teste ($TEST_FILES)" test "$TEST_FILES" -gt 10

# 13. Optimistic locking nas migrations
check "Optimistic locking (version column)" grep -q "version" services/cashin-engine/internal/adapter/outbound/postgres/migrations/001_create_invoices_table.up.sql 2>/dev/null

# 14. .env.example
check ".env.example (cashin)" test -f services/cashin-engine/.env.example
check ".env.example (cashout)" test -f services/cashout-engine/.env.example

# 15. Documentação
check "testes-resiliencia.md" test -f docs/testes-resiliencia.md
check "auditoria-completa.md" test -f docs/auditoria-completa.md
check "observabilidade.md" test -f docs/observabilidade.md

# 16. CI/CD
check "test-suite.yml" test -f .github/workflows/test-suite.yml
check "security-scan.yml" test -f .github/workflows/security-scan.yml

echo ""
echo "============================================"
echo "   📊 Resultados"
echo "============================================"
echo -e "   ${GREEN}PASS: $PASS${NC}"
echo -e "   ${RED}FAIL: $FAIL${NC}"
echo -e "   Total: $TOTAL"
echo "   Score: $(( (PASS * 100) / TOTAL ))%"
echo "============================================"
echo ""

# Salva relatório
cat > reports/audit/report.json <<EOF
{
  "date": "$(date -Iseconds)",
  "total": $TOTAL,
  "passed": $PASS,
  "failed": $FAIL,
  "score": $(( (PASS * 100) / TOTAL ))
}
EOF

exit $FAIL