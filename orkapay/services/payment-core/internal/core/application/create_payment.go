package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// CreatePaymentUseCase orquestra a criaÃ§Ã£o de um novo pagamento.
// Fluxo: validar input â†’ criar Payment (domain) â†’ persistir â†’ publicar evento.
type CreatePaymentUseCase struct {
	repo port.PaymentRepository
	pub  port.EventPublisher
}

// NewCreatePaymentUseCase cria uma nova instÃ¢ncia do caso de uso.
func NewCreatePaymentUseCase(repo port.PaymentRepository, pub port.EventPublisher) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{
		repo: repo,
		pub:  pub,
	}
}

// Create executa o fluxo de criaÃ§Ã£o de pagamento.
func (uc *CreatePaymentUseCase) Create(ctx context.Context, input port.CreatePaymentInput) (*port.PaymentOutput, error) {
	// 1. Criar value objects do domÃ­nio
	amount, err := domain.NewMoney(input.AmountCents)
	if err != nil {
		return nil, fmt.Errorf("valor invÃ¡lido: %w", err)
	}

	payerPixKey, err := domain.NewPixKey(input.PayerPixKey)
	if err != nil {
		return nil, fmt.Errorf("chave Pix do pagador invÃ¡lida: %w", err)
	}

	payeePixKey, err := domain.NewPixKey(input.PayeePixKey)
	if err != nil {
		return nil, fmt.Errorf("chave Pix do beneficiÃ¡rio invÃ¡lida: %w", err)
	}

	// 2. Criar entidades do domÃ­nio
	payer := domain.Payer{
		Name:   input.PayerName,
		TaxID:  input.PayerTaxID,
		Email:  input.PayerEmail,
		PixKey: payerPixKey,
	}

	payee := domain.Payee{
		Name:   input.PayeeName,
		TaxID:  input.PayeeTaxID,
		PixKey: payeePixKey,
	}

	paymentType := domain.PaymentType(input.PaymentType)
	method := domain.PaymentMethod(input.Method)

	// 3. Criar aggregate (aplica regras de negÃ³cio e emite eventos)
	payment, err := domain.NewPayment(
		paymentType,
		method,
		amount,
		payer,
		payee,
		input.Description,
		input.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar pagamento: %w", err)
	}

	// 4. Persistir no repositÃ³rio
	if err := uc.repo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("falha ao salvar pagamento: %w", err)
	}

	// 5. Publicar eventos de domÃ­nio
	for _, event := range payment.Events() {
		if err := uc.pub.Publish(ctx, event); err != nil {
			// Log e continua â€” nÃ£o falha a operaÃ§Ã£o por causa do evento
			// Em produÃ§Ã£o, usar outbox pattern ou retry
			_ = err
		}
	}

	return port.MapToOutput(payment), nil
}

// MapToOutput converte um domain.Payment para PaymentOutput.
func MapToOutput(p *domain.Payment) *port.PaymentOutput {
	return &port.PaymentOutput{
		ID:          p.ID(),
		Status:      string(p.Status()),
		Amount:      p.Amount().String(),
		Fee:         p.Fee().String(),
		NetAmount:   p.NetAmount().String(),
		PaymentType: string(p.PaymentType()),
		Method:      string(p.Method()),
		PayerName:   p.Payer().Name,
		PayerTaxID:  p.Payer().TaxID,
		PayeeName:   p.Payee().Name,
		PayeeTaxID:  p.Payee().TaxID,
		Description: p.Description(),
		ExpiresAt:   FormatTime(p.ExpiresAt()),
		CreatedAt:   p.CreatedAt().Format(time.RFC3339),
	}
}

// FormatTime formata um *time.Time para string RFC3339 ou vazio.
func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
