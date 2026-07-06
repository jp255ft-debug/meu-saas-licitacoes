package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jp255ft-debug/cashin-engine/internal/adapter/inbound/rest/middleware"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/service/webhook"
)

// WebhookHandler expõe endpoints para configuração de webhooks.
type WebhookHandler struct {
	svc *webhook.WebhookService
}

// NewWebhookHandler cria uma nova instância do handler.
func NewWebhookHandler(svc *webhook.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

// ConfigureWebhookRequest representa o body da requisição.
type ConfigureWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// ConfigureWebhook godoc
// POST /api/v1/webhooks/configure
func (h *WebhookHandler) ConfigureWebhook(w http.ResponseWriter, r *http.Request) {
	var req ConfigureWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if req.URL == "" {
		respondError(w, http.StatusBadRequest, "missing_url", "url é obrigatório")
		return
	}
	if len(req.Events) == 0 {
		respondError(w, http.StatusBadRequest, "missing_events", "pelo menos um evento deve ser informado")
		return
	}

	clientID := middleware.GetClientID(r.Context())
	if clientID == "" {
		respondError(w, http.StatusUnauthorized, "unauthorized", "cliente não autenticado")
		return
	}

	wh := &domain.Webhook{
		ID:       uuid.New().String(),
		ClientID: clientID,
		URL:      req.URL,
		Events:   req.Events,
		Active:   true,
		Secret:   uuid.New().String(),
	}

	if err := h.svc.Save(r.Context(), wh); err != nil {
		slog.ErrorContext(r.Context(), "failed to save webhook", "error", err)
		respondError(w, http.StatusInternalServerError, "save_failed", "falha ao salvar webhook")
		return
	}

	respondJSON(w, http.StatusCreated, wh)
}

// ListWebhooks godoc
// GET /api/v1/webhooks
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	clientID := middleware.GetClientID(r.Context())
	if clientID == "" {
		respondError(w, http.StatusUnauthorized, "unauthorized", "cliente não autenticado")
		return
	}

	webhooks, err := h.svc.FindByClientID(r.Context(), clientID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list webhooks", "error", err)
		respondError(w, http.StatusInternalServerError, "list_failed", "falha ao listar webhooks")
		return
	}

	if webhooks == nil {
		webhooks = []*domain.Webhook{}
	}

	respondJSON(w, http.StatusOK, webhooks)
}

// DeleteWebhook godoc
// DELETE /api/v1/webhooks/{id}
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id do webhook é obrigatório")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		slog.ErrorContext(r.Context(), "failed to delete webhook", "error", err)
		respondError(w, http.StatusInternalServerError, "delete_failed", "falha ao deletar webhook")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}