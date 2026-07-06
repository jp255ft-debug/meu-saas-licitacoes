package outbound

import (
	"context"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
)

// WebhookRepository define o contrato para persistência de webhooks.
type WebhookRepository interface {
	Save(ctx context.Context, webhook *domain.Webhook) error
	FindByClientID(ctx context.Context, clientID string) ([]*domain.Webhook, error)
	FindByID(ctx context.Context, id string) (*domain.Webhook, error)
	Update(ctx context.Context, webhook *domain.Webhook) error
	Delete(ctx context.Context, id string) error
}