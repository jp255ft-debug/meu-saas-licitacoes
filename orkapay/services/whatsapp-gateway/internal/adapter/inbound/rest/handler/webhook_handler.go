package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/domain"
	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/service/whatsapp"
)

type WebhookHandler struct {
	service *whatsapp.Service
}

func NewWebhookHandler(service *whatsapp.Service) *WebhookHandler {
	return &WebhookHandler{service: service}
}

type WebhookPayload struct {
	From string `json:"from"`
	To   string `json:"to"`
	Body string `json:"body"`
	Type string `json:"type"`
	Time string `json:"time"`
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.WarnContext(ctx, "failed to decode webhook payload", "error", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	ts, _ := time.Parse(time.RFC3339, payload.Time)
	if ts.IsZero() {
		ts = time.Now()
	}

	msg := &domain.WhatsAppMessage{
		FromNumber:  payload.From,
		ToNumber:    payload.To,
		Body:        payload.Body,
		MessageType: payload.Type,
		Timestamp:   ts,
	}

	go func() {
		bgCtx := context.Background()
		if err := h.service.ProcessMessage(bgCtx, msg); err != nil {
			slog.ErrorContext(bgCtx, "failed to process message", "error", err, "from", msg.FromNumber)
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}