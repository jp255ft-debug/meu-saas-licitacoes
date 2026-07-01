package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// PaymentHandler expÃµe os casos de uso via HTTP REST.
type PaymentHandler struct {
	creator   port.PaymentCreator
	processor port.PaymentProcessor
	canceller port.PaymentCanceller
	webhook   port.PaymentWebhookHandler
}

// NewPaymentHandler cria um novo handler.
func NewPaymentHandler(
	creator port.PaymentCreator,
	processor port.PaymentProcessor,
	canceller port.PaymentCanceller,
	webhook port.PaymentWebhookHandler,
) *PaymentHandler {
	return &PaymentHandler{
		creator:   creator,
		processor: processor,
		canceller: canceller,
		webhook:   webhook,
	}
}

// CreatePaymentRequest representa o body da requisiÃ§Ã£o POST /payments.
type CreatePaymentRequest struct {
	PaymentType string            `json:"payment_type"`
	Method      string            `json:"method"`
	AmountCents int64             `json:"amount_cents"`
	PayerName   string            `json:"payer_name"`
	PayerTaxID  string            `json:"payer_tax_id"`
	PayerEmail  string            `json:"payer_email"`
	PayerPixKey string            `json:"payer_pix_key"`
	PayeeName   string            `json:"payee_name"`
	PayeeTaxID  string            `json:"payee_tax_id"`
	PayeePixKey string            `json:"payee_pix_key"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreatePaymentResponse representa a resposta da criaÃ§Ã£o.
type CreatePaymentResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Amount      string `json:"amount"`
	Fee         string `json:"fee"`
	NetAmount   string `json:"net_amount"`
	PaymentType string `json:"payment_type"`
	Method      string `json:"method"`
	Description string `json:"description"`
	ExpiresAt   string `json:"expires_at"`
	CreatedAt   string `json:"created_at"`
}

// CreatePayment godoc
// POST /api/v1/payments
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	input := port.CreatePaymentInput{
		PaymentType: req.PaymentType,
		Method:      req.Method,
		AmountCents: req.AmountCents,
		PayerName:   req.PayerName,
		PayerTaxID:  req.PayerTaxID,
		PayerEmail:  req.PayerEmail,
		PayerPixKey: req.PayerPixKey,
		PayeeName:   req.PayeeName,
		PayeeTaxID:  req.PayeeTaxID,
		PayeePixKey: req.PayeePixKey,
		Description: req.Description,
		Metadata:    req.Metadata,
	}

	output, err := h.creator.Create(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "creation_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, CreatePaymentResponse{
		ID:          output.ID,
		Status:      output.Status,
		Amount:      output.Amount,
		Fee:         output.Fee,
		NetAmount:   output.NetAmount,
		PaymentType: output.PaymentType,
		Method:      output.Method,
		Description: output.Description,
		ExpiresAt:   output.ExpiresAt,
		CreatedAt:   output.CreatedAt,
	})
}

// GetPayment godoc
// GET /api/v1/payments/{id}
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id do pagamento Ã© obrigatÃ³rio")
		return
	}

	// Para o MVP, usamos o processor para buscar (ou criar um getter separado)
	// Aqui simplificamos retornando 501 â€” o getter serÃ¡ implementado no futuro
	respondError(w, http.StatusNotImplemented, "not_implemented", "busca por ID serÃ¡ implementada em breve")
}

// CancelPaymentRequest representa o body da requisiÃ§Ã£o de cancelamento.
type CancelPaymentRequest struct {
	Reason string `json:"reason"`
}

// CancelPayment godoc
// POST /api/v1/payments/{id}/cancel
func (h *PaymentHandler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id do pagamento Ã© obrigatÃ³rio")
		return
	}

	var req CancelPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	output, err := h.canceller.Cancel(r.Context(), id, req.Reason)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "cancel_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, CreatePaymentResponse{
		ID:        output.ID,
		Status:    output.Status,
		Amount:    output.Amount,
		Fee:       output.Fee,
		NetAmount: output.NetAmount,
		CreatedAt: output.CreatedAt,
	})
}

// WebhookAsaas godoc
// POST /api/v1/webhooks/asaas
func (h *PaymentHandler) WebhookAsaas(w http.ResponseWriter, r *http.Request) {
	var event port.WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_webhook", err.Error())
		return
	}

	if err := h.webhook.HandleWebhook(r.Context(), event); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "webhook_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

// respondJSON envia uma resposta JSON.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// respondError envia uma resposta de erro padronizada.
func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
