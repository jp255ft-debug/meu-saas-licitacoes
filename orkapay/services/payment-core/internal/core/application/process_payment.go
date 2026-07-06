package application

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// ProcessPaymentUseCase orquestra o processamento de um pagamento pendente.
// Fluxo: buscar no repo â†’ chamar gateway â†’ aprovar â†’ persistir â†’ publicar evento.
type ProcessPaymentUseCase struct {
	repo port.PaymentRepository
	pub  port.EventPublisher
	gw   port.PaymentGateway
}

// NewProcessPaymentUseCase cria uma nova instÃ¢ncia do caso de uso.
func NewProcessPaymentUseCase(repo port.PaymentRepository, pub port.EventPublisher, gw port.PaymentGateway) *ProcessPaymentUseCase {
	return &ProcessPaymentUseCase{
		repo: repo,
		pub:  pub,
		gw:   gw,
	}
}

// Process executa o fluxo de processamento de pagamento.
func (uc *ProcessPaymentUseCase) Process(ctx context.Context, paymentID string) (*port.PaymentOutput, error) {
	// 1. Buscar pagamento no repositÃ³rio
	payment, err := uc.repo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("pagamento nÃ£o encontrado: %w", err)
	}

	// 2. Verificar se pode ser processado
	if payment.Status() != domain.PaymentStatusPending {
		return nil, domain.ErrPaymentNotPending
	}

	// 3. Chamar gateway para criar cobranÃ§a
	gwResponse, err := uc.gw.CreatePayment(ctx, payment)
	if err != nil {
		// Falha no gateway â†’ marca como falho
		if failErr := payment.Fail(err.Error()); failErr != nil {
			return nil, fmt.Errorf("falha ao marcar pagamento como falho: %w", failErr)
		}
		if saveErr := uc.repo.Update(ctx, payment); saveErr != nil {
			return nil, fmt.Errorf("falha ao salvar pagamento apÃ³s falha: %w", saveErr)
		}
		uc.publishEvents(ctx, payment)
		return port.MapToOutput(payment), fmt.Errorf("gateway rejeitou: %w", err)
	}

	// 4. Aprovar pagamento com ID externo do gateway
	if err := payment.Approve(gwResponse.ExternalID); err != nil {
		return nil, fmt.Errorf("falha ao aprovar pagamento: %w", err)
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
func (uc *ProcessPaymentUseCase) publishEvents(ctx context.Context, payment *domain.Payment) {
	for _, event := range payment.Events() {
		if err := uc.pub.Publish(ctx, event); err != nil {
			// Log e continua â€” nÃ£o falha a operaÃ§Ã£o
			_ = err
		}
	}
}
