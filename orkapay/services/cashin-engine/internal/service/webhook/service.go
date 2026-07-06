package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// WebhookService gerencia o dispatch de eventos para clientes.
type WebhookService struct {
	repo    outbound.WebhookRepository
	client  *http.Client
}

// NewWebhookService cria uma nova instância do serviço.
func NewWebhookService(repo outbound.WebhookRepository) *WebhookService {
	return &WebhookService{
		repo: repo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Save persiste um webhook.
func (s *WebhookService) Save(ctx context.Context, wh *domain.Webhook) error {
	return s.repo.Save(ctx, wh)
}

// FindByClientID busca webhooks de um cliente.
func (s *WebhookService) FindByClientID(ctx context.Context, clientID string) ([]*domain.Webhook, error) {
	return s.repo.FindByClientID(ctx, clientID)
}

// Delete remove um webhook.
func (s *WebhookService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// generateSignature cria uma assinatura HMAC-SHA256 do payload.
func generateSignature(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// Dispatch envia notificações para todos os webhooks ativos do cliente.
func (s *WebhookService) Dispatch(ctx context.Context, clientID, eventType string, data interface{}) {
	webhooks, err := s.repo.FindByClientID(ctx, clientID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find webhooks for client",
			"client_id", clientID, "error", err)
		return
	}

	var targets []*domain.Webhook
	for _, wh := range webhooks {
		if wh.ShouldDeliver(eventType) {
			targets = append(targets, wh)
		}
	}

	if len(targets) == 0 {
		slog.DebugContext(ctx, "no webhooks configured for event",
			"client_id", clientID, "event", eventType)
		return
	}

	for _, wh := range targets {
		go s.deliver(wh, eventType, data)
	}
}

// deliver realiza o HTTP POST com retry.
func (s *WebhookService) deliver(wh *domain.Webhook, eventType string, data interface{}) {
	payload := domain.WebhookEvent{
		Event:     eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal webhook payload", "webhook_id", wh.ID, "error", err)
		return
	}

	signature := generateSignature(wh.Secret, body)

	backoff := []time.Duration{1 * time.Second, 5 * time.Second, 30 * time.Second}
	maxAttempts := len(backoff) + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff[attempt-1])
		}

		req, err := http.NewRequest("POST", wh.URL, bytes.NewReader(body))
		if err != nil {
			slog.Error("failed to create webhook request", "webhook_id", wh.ID, "error", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-OrkaPay-Signature", signature)
		req.Header.Set("X-OrkaPay-Event", eventType)

		resp, err := s.client.Do(req)
		if err != nil {
			slog.Warn("webhook delivery failed",
				"webhook_id", wh.ID, "attempt", attempt+1, "error", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			slog.Info("webhook delivered successfully",
				"webhook_id", wh.ID, "event", eventType, "status", resp.StatusCode)
			return
		}

		slog.Warn("webhook returned non-2xx",
			"webhook_id", wh.ID, "event", eventType, "status", resp.StatusCode, "attempt", attempt+1)
	}

	slog.Error("webhook delivery failed after all retries",
		"webhook_id", wh.ID, "url", wh.URL, "event", eventType)
}