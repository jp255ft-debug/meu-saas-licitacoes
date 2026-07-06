# =============================================================================
# Dockerfile — OrkaPay Payment Core
# Multi-stage build: compile → distroless runtime
# =============================================================================

# ---- Stage 1: Build ----
FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /src

# Cache de dependências (layer separado para acelerar rebuilds)
COPY go.mod go.sum ./
RUN go mod download

# Código fonte
COPY . .

# Compilação otimizada:
#   - CGO_ENABLED=0  → binário estático (sem libc)
#   - ldflags        → stripping debug symbols, menor binário
#   - -v             → verbose para debug
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -v -ldflags="-s -w -extldflags=-static" \
    -o /app/payment-core ./cmd/server

# ---- Stage 2: Runtime ----
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

# Timezone (herdado do builder)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Binário compilado
COPY --from=builder /app/payment-core /app/payment-core

# Porta do servidor HTTP
EXPOSE 8080

# Health check
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/payment-core", "-health"]

# Usuário não-root (distroless já usa nonroot)
USER 65532:65532

ENTRYPOINT ["/app/payment-core"]
