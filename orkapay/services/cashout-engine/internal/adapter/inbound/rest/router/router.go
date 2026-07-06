package router

import (
	"net/http"
	"time"

	"github.com/jp255ft-debug/cashout-engine/internal/adapter/inbound/rest/handler"
	"github.com/jp255ft-debug/cashout-engine/internal/adapter/inbound/rest/middleware"
	"github.com/jp255ft-debug/cashout-engine/internal/service/transfer"
)

// New cria e configura o router HTTP com todas as rotas e middlewares.
func New(
	transferSvc *transfer.ServiceImpl,
	apiKeys map[string]string,
	rateLimiter *middleware.RateLimiter,
) http.Handler {
	mux := http.NewServeMux()

	healthH := handler.NewHealthHandler()
	mux.HandleFunc("GET /health", healthH.Health)
	mux.HandleFunc("GET /ready", healthH.Ready)

	transferH := handler.NewTransferHandler(transferSvc)
	mux.HandleFunc("POST /api/v1/transfers", transferH.CreateTransfer)
	mux.HandleFunc("GET /api/v1/transfers/{id}", transferH.GetTransfer)
	mux.HandleFunc("GET /api/v1/transfers", transferH.ListTransfers)
	mux.HandleFunc("POST /api/v1/transfers/{id}/approve", transferH.ApproveTransfer)
	mux.HandleFunc("POST /api/v1/transfers/{id}/settle", transferH.SettleTransfer)
	mux.HandleFunc("POST /api/v1/transfers/{id}/fail", transferH.FailTransfer)
	mux.HandleFunc("POST /api/v1/transfers/{id}/cancel", transferH.CancelTransfer)

	var h http.Handler = mux
	h = middleware.Timeout(30 * time.Second)(h)
	h = middleware.SecurityHeaders(h)
	h = middleware.RequestIDMiddleware(h)
	h = middleware.LoggingMiddleware(h)
	h = rateLimiter.Middleware(h)
	h = middleware.APIKeyAuth(apiKeys)(h)

	return h
}