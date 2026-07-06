package port

import (
	"context"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
)

// PaymentCreator define a criaÃ§Ã£o de pagamentos.
type PaymentCreator interface {
	Create(ctx context.Context, input CreatePaymentInput) (*PaymentOutput, error)
}

// PaymentProcessor define o processamento de pagamentos.
type PaymentProcessor interface {
	Process(ctx context.Context, paymentID string) (*PaymentOutput, error)
}

// PaymentCanceller define o cancelamento de pagamentos.
type PaymentCanceller interface {
	Cancel(ctx context.Context, paymentID string, reason string) (*PaymentOutput, error)
}

// PaymentWebhookHandler define o processamento de webhooks.
type PaymentWebhookHandler interface {
	HandleWebhook(ctx context.Context, event WebhookEvent) error
}

// CreatePaymentInput representa os dados de entrada para criar um pagamento.
type CreatePaymentInput struct {
	PaymentType string
	Method      string
	AmountCents int64
	PayerName   string
	PayerTaxID  string
	PayerEmail  string
	PayerPixKey string
	PayeeName   string
	PayeeTaxID  string
	PayeePixKey string
	Description string
	Metadata    map[string]string
}

// PaymentOutput representa a resposta de uma operaÃ§Ã£o com pagamento.
type PaymentOutput struct {
	ID          string
	Status      string
	Amount      string
	Fee         string
	NetAmount   string
	PaymentType string
	Method      string
	PayerName   string
	PayerTaxID  string
	PayeeName   string
	PayeeTaxID  string
	Description string
	ExpiresAt   string
	CreatedAt   string
}

// MapToOutput converte um domain.Payment para PaymentOutput.
func MapToOutput(p *domain.Payment) *PaymentOutput {
	return &PaymentOutput{
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

// WebhookEvent representa um evento de webhook recebido.
type WebhookEvent struct {
	ID         string
	Type       string
	ExternalID string
	Status     string
	Amount     int64
	Timestamp  string
	Signature  string
}
