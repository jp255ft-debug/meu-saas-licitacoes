package application

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// WebhookPaymentUseCase orquestra o processamento de webhooks do gateway.
// Fluxo: validar evento â†’ buscar pagamento â†’ atualizar status conforme evento â†’ persistir â†’ publicar evento.
type WebhookPaymentUseCase struct {
	repo port.PaymentRepository
	pub  port.EventPublisher
}

// NewWebhookPaymentUseCase cria uma nova instÃ¢ncia do caso de uso.
func NewWebhookPaymentUseCase(repo port.PaymentRepository, pub port.EventPublisher) *WebhookPaymentUseCase {
	return &WebhookPaymentUseCase{
		repo: repo,
		pub:  pub,
	}
}

// HandleWebhook processa um evento de webhook recebido.
func (uc *WebhookPaymentUseCase) HandleWebhook(ctx context.Context, event port.WebhookEvent) error {
	// 1. Buscar pagamento pelo ID externo
	payment, err := uc.repo.FindByExternalID(ctx, event.ExternalID)
	if err != nil {
		return fmt.Errorf("pagamento nÃ£o encontrado para externalID %s: %w", event.ExternalID, err)
	}

	// 2. Atualizar status conforme o tipo do evento
	switch event.Status {
	case "RECEIVED", "CONFIRMED":
		// Pagamento recebido/confirmado â†’ liquidar
		if err := payment.Settle(); err != nil {
			return fmt.Errorf("falha ao liquidar pagamento: %w", err)
		}
	case "OVERDUE":
		// Pagamento vencido â†’ falhar
		if err := payment.Fail("pagamento vencido"); err != nil {
			return fmt.Errorf("falha ao marcar pagamento como vencido: %w", err)
		}
	case "CANCELLED", "REFUNDED":
		// Pagamento cancelado/estornado
		if err := payment.Cancel("cancelado pelo gateway"); err != nil {
			return fmt.Errorf("falha ao cancelar pagamento: %w", err)
		}
	default:
		// Status nÃ£o mapeado â€” ignorar ou logar
		return nil
	}

	// 3. Persistir atualizaÃ§Ã£o
	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("falha ao atualizar pagamento: %w", err)
	}

	// 4. Publicar eventos
	for _, evt := range payment.Events() {
		if err := uc.pub.Publish(ctx, evt); err != nil {
			_ = err
		}
	}

	return nil
}
