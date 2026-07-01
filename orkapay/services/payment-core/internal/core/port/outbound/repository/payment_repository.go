package port

import (
	"context"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
)

// PaymentRepository define as operaÃ§Ãµes de persistÃªncia do aggregate Payment.
// Interface de outbound port â€” implementada pelo adaptador PostgreSQL.
// Segue o Repository Pattern para desacoplar domÃ­nio de infraestrutura.
type PaymentRepository interface {
	// Save persiste um novo pagamento.
	Save(ctx context.Context, payment *domain.Payment) error

	// FindByID busca um pagamento pelo ID interno.
	FindByID(ctx context.Context, id string) (*domain.Payment, error)

	// FindByExternalID busca um pagamento pelo ID do gateway externo.
	FindByExternalID(ctx context.Context, externalID string) (*domain.Payment, error)

	// Update atualiza o estado de um pagamento existente.
	// Usa optimistic locking (version) para evitar conflitos.
	Update(ctx context.Context, payment *domain.Payment) error

	// ListByStatus lista pagamentos por status com paginaÃ§Ã£o.
	ListByStatus(ctx context.Context, status domain.PaymentStatus, limit, offset int) ([]*domain.Payment, error)

	// ListByPayerTaxID lista pagamentos de um pagador especÃ­fico.
	ListByPayerTaxID(ctx context.Context, taxID string, limit, offset int) ([]*domain.Payment, error)

	// ListExpired lista pagamentos pendentes que jÃ¡ expiraram.
	ListExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Payment, error)

	// CountByStatus retorna a contagem de pagamentos por status.
	CountByStatus(ctx context.Context, status domain.PaymentStatus) (int64, error)
}
