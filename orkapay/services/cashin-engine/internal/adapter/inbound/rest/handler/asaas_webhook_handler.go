package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/event"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

type AsaasWebhookHandler struct {
	eventPublisher outbound.EventPublisher
}

func NewAsaasWebhookHandler(publisher outbound.EventPublisher) *AsaasWebhookHandler {
	return &AsaasWebhookHandler{eventPublisher: publisher}
}

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

func (h *AsaasWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	)

	if payload.Event == "PAYMENT_RECEIVED" || payload.Event == "PAYMENT_CONFIRMED" {
		paymentEvent := &event.PaymentConfirmedEvent{
			PaymentID:  payload.Payment.ID,
			ExternalID: payload.Payment.ExternalID,
			Amount:     int64(payload.Payment.Value * 100),
			Status:     payload.Payment.Status,
			PixQRCode:  payload.Payment.PixQRCode,
			PixKey:     payload.Payment.PixKey,
			Barcode:    payload.Payment.Barcode,
			InvoiceURL: payload.Payment.InvoiceURL,
			OccurredAt: time.Now(),
		}

		if err := h.eventPublisher.PublishPaymentConfirmed(ctx, paymentEvent); err != nil {
			slog.ErrorContext(ctx, "failed to publish payment event", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}