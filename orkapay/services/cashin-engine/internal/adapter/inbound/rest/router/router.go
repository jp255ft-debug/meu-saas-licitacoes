package router

import (
	"net/http"
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/adapter/inbound/rest/handler"
	"github.com/jp255ft-debug/cashin-engine/internal/adapter/inbound/rest/middleware"
	"github.com/jp255ft-debug/cashin-engine/internal/service/invoice"
	"github.com/jp255ft-debug/cashin-engine/internal/service/webhook"
)

// New cria e configura o router HTTP com todas as rotas e middlewares.
func New(
	invoiceSvc *invoice.ServiceImpl,
	webhookSvc *webhook.WebhookService,
	authProviders []middleware.AuthProvider,
	rateLimiter *middleware.RateLimiter,
	asaasWebhookHandler *handler.AsaasWebhookHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Metrics
	metricsHandler := handler.NewMetrics()
	mux.HandleFunc("GET /metrics", metricsHandler.MetricsHandler)

	// Health check endpoints (sem autenticação)
	healthH := handler.NewHealthHandler()
	mux.HandleFunc("GET /health", healthH.Health)
	mux.HandleFunc("GET /ready", healthH.Ready)

	// Asaas webhook endpoint (público, validado por token no header)
	if asaasWebhookHandler != nil {
		mux.HandleFunc("POST /api/v1/webhooks/asaas", asaasWebhookHandler.HandleWebhook)
	}

	// Invoice endpoints (autenticados)
	invoiceH := handler.NewInvoiceHandler(invoiceSvc)
	mux.HandleFunc("POST /api/v1/invoices", invoiceH.CreateInvoice)
	mux.HandleFunc("GET /api/v1/invoices/{id}", invoiceH.GetInvoice)
	mux.HandleFunc("GET /api/v1/invoices", invoiceH.ListInvoices)
	mux.HandleFunc("POST /api/v1/invoices/{id}/issue", invoiceH.IssueInvoice)
	mux.HandleFunc("POST /api/v1/invoices/{id}/pay", invoiceH.PayInvoice)
	mux.HandleFunc("POST /api/v1/invoices/{id}/cancel", invoiceH.CancelInvoice)
	mux.HandleFunc("POST /api/v1/invoices/{id}/expire", invoiceH.ExpireInvoice)

	// Webhook endpoints (autenticados)
	webhookH := handler.NewWebhookHandler(webhookSvc)
	mux.HandleFunc("POST /api/v1/webhooks/configure", webhookH.ConfigureWebhook)
	mux.HandleFunc("GET /api/v1/webhooks", webhookH.ListWebhooks)
	mux.HandleFunc("DELETE /api/v1/webhooks/{id}", webhookH.DeleteWebhook)

	// Aplicar middleware (ordem inversa: último aplicado = primeiro executado)
	var h http.Handler = mux
	h = middleware.Timeout(30 * time.Second)(h)
	h = middleware.SecurityHeaders(h)
	h = middleware.RequestIDMiddleware(h)
	h = middleware.LoggingMiddleware(h)
	h = rateLimiter.Middleware(h)
	h = middleware.APIKeyAuth(authProviders...)(h)

	return h
}