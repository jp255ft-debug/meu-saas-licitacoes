package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// AsaasWebhookHandler processa webhooks recebidos do Asaas.
type AsaasWebhookHandler struct {
	webhookToken string
}

// NewAsaasWebhookHandler cria uma nova instância do handler.
func NewAsaasWebhookHandler(webhookToken string) *AsaasWebhookHandler {
	return &AsaasWebhookHandler{webhookToken: webhookToken}
}

// AsaasWebhookPayload representa o payload enviado pelo Asaas.
type AsaasWebhookPayload struct {
	Event  string `json:"event"`
	Payment struct {
		ID          string  `json:"id"`
		Customer    string  `json:"customer"`
		BillingType string  `json:"billingType"`
		DueDate     string  `json:"dueDate"`
		Value       float64 `json:"value"`
		Status      string  `json:"status"`
		ExternalID  string  `json:"externalReference"`
		PixQRCode   string  `json:"pixQrCode"`
		PixKey      string  `json:"pixKey"`
		Barcode     string  `json:"barcode"`
		InvoiceURL  string  `json:"invoiceUrl"`
	} `json:"payment"`
	Split []struct {
		WalletId    string  `json:"walletId"`
		FixedValue  float64 `json:"fixedValue"`
		TotalAmount float64 `json:"totalAmount"`
	} `json:"split,omitempty"`
}

// HandleWebhook processa o webhook recebido.
func (h *AsaasWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Valida o token de autenticação enviado pelo Asaas no header
	token := r.Header.Get("asaas-access-token")
	if token == "" || token != h.webhookToken {
		slog.WarnContext(ctx, "unauthorized webhook attempt",
			"provided_token", token != "",
			"has_token_configured", h.webhookToken != "",
		)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload AsaasWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.WarnContext(ctx, "failed to decode asaas webhook payload", "error", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "received webhook from Asaas",
		"event", payload.Event,
		"payment_id", payload.Payment.ID,
		"external_id", payload.Payment.ExternalID,
		"status", payload.Payment.Status,
		"value", payload.Payment.Value,
	)

	// Nota: o processamento do webhook (atualizar invoice, creditar ledger)
	// será implementado quando o fluxo de eventos estiver completo.
	// Atualmente, o Asaas envia webhooks diretamente e o cashin já
	// registra o pagamento no Asaas. A atualização local será feita
	// via reconciliação (reconciliation service) ou em próxima versão.

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}