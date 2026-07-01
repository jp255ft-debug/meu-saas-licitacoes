package port

import (
	"context"

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

// WebhookEvent representa um evento de webhook recebido.
type WebhookEvent struct {
	ID        string
	Type      string
	ExternalID string
	Status    string
	Amount    int64
	Timestamp string
	Signature string
}
