package port

import (
	"context"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
)

// EventPublisher define a interface para publicaÃ§Ã£o de eventos de domÃ­nio.
// Implementada pelo adaptador Kafka (ou mock em testes).
type EventPublisher interface {
	Publish(ctx context.Context, event domain.DomainEvent) error
	Close() error
}
