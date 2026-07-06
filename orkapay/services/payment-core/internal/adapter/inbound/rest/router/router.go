package router

import (
	"net/http"

	"github.com/jp255ft-debug/payment-core/internal/adapter/inbound/rest/handler"
)

// New cria e configura o router HTTP.
func New(
	paymentHandler *handler.PaymentHandler,
	healthHandler *handler.HealthHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Health checks
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	// API v1 â€” Payments
	mux.HandleFunc("POST /api/v1/payments", paymentHandler.CreatePayment)
	mux.HandleFunc("GET /api/v1/payments/{id}", paymentHandler.GetPayment)
	mux.HandleFunc("POST /api/v1/payments/{id}/cancel", paymentHandler.CancelPayment)

	// Webhooks
	mux.HandleFunc("POST /api/v1/webhooks/asaas", paymentHandler.WebhookAsaas)

	return mux
}
