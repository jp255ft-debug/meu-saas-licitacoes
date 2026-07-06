package application

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// CancelPaymentUseCase orquestra o cancelamento de um pagamento pendente.
// Fluxo: buscar no repo â†’ cancelar no gateway â†’ cancelar no domain â†’ persistir â†’ publicar evento.
type CancelPaymentUseCase struct {
	repo port.PaymentRepository
	pub  port.EventPublisher
	gw   port.PaymentGateway
}

// NewCancelPaymentUseCase cria uma nova instÃ¢ncia do caso de uso.
func NewCancelPaymentUseCase(repo port.PaymentRepository, pub port.EventPublisher, gw port.PaymentGateway) *CancelPaymentUseCase {
	return &CancelPaymentUseCase{
		repo: repo,
		pub:  pub,
		gw:   gw,
	}
}

// Cancel executa o fluxo de cancelamento de pagamento.
func (uc *CancelPaymentUseCase) Cancel(ctx context.Context, paymentID string, reason string) (*port.PaymentOutput, error) {
	// 1. Buscar pagamento no repositÃ³rio
	payment, err := uc.repo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("pagamento nÃ£o encontrado: %w", err)
	}

	// 2. Verificar se pode ser cancelado (regra de negÃ³cio no domain)
	if payment.Status() != domain.PaymentStatusPending {
		return nil, domain.ErrPaymentCannotBeCancelled
	}

	// 3. Cancelar no gateway (se jÃ¡ tiver externalID)
	if payment.ExternalID() != "" {
		if err := uc.gw.CancelPayment(ctx, payment.ExternalID(), reason); err != nil {
			// Se o gateway falhar, ainda tenta cancelar localmente
			// (o gateway pode jÃ¡ ter cancelado ou nÃ£o ter recebido a cobranÃ§a)
			_ = err
		}
	}

	// 4. Cancelar no domain (aplica regras de negÃ³cio e emite evento)
	if err := payment.Cancel(reason); err != nil {
		return nil, fmt.Errorf("falha ao cancelar pagamento: %w", err)
	}

	// 5. Persistir atualizaÃ§Ã£o
	if err := uc.repo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("falha ao atualizar pagamento: %w", err)
	}

	// 6. Publicar eventos
	uc.publishEvents(ctx, payment)

	return port.MapToOutput(payment), nil
}

// publishEvents publica os eventos pendentes do aggregate.
func (uc *CancelPaymentUseCase) publishEvents(ctx context.Context, payment *domain.Payment) {
	for _, event := range payment.Events() {
		if err := uc.pub.Publish(ctx, event); err != nil {
			// Log e continua
			_ = err
		}
	}
}
