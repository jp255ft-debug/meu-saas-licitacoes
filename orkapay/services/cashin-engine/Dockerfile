# =============================================================================
# Dockerfile — OrkaPay Cashin Engine
# Multi-stage build: compile → distroless runtime
# =============================================================================

# ---- Stage 1: Build ----
FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags=-static" \
    -o /app/cashin-engine ./cmd/api

# ---- Stage 2: Runtime ----
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/cashin-engine /app/cashin-engine

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/cashin-engine", "-health"]

USER 65532:65532

ENTRYPOINT ["/app/cashin-engine"]