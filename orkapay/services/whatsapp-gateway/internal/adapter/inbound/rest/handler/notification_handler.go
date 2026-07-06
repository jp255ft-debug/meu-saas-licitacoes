package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/service/notification"
)

type NotificationHandler struct {
	service *notification.Service
}

func NewNotificationHandler(service *notification.Service) *NotificationHandler {
	return &NotificationHandler{service: service}
}

type NotificationRequest struct {
	ToNumber  string `json:"to_number"`
	InvoiceID string `json:"invoice_id"`
	Amount    int64  `json:"amount"`
}

func (h *NotificationHandler) SendPaymentConfirmation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.service.SendPaymentConfirmation(ctx, req.ToNumber, req.InvoiceID, req.Amount); err != nil {
		slog.ErrorContext(ctx, "failed to send notification", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}