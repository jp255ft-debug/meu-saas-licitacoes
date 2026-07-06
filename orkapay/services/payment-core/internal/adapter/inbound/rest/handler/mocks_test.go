package handler

import (
	"context"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// mockPaymentCreator implementa port.PaymentCreator para testes.
type mockPaymentCreator struct {
	output *port.PaymentOutput
	err    error
}

func (m *mockPaymentCreator) Create(ctx context.Context, input port.CreatePaymentInput) (*port.PaymentOutput, error) {
	return m.output, m.err
}

// mockPaymentGetter implementa port.PaymentGetter para testes.
type mockPaymentGetter struct {
	payment *domain.Payment
	err     error
}

func (m *mockPaymentGetter) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	return m.payment, m.err
}

// mockPaymentCanceller implementa port.PaymentCanceller para testes.
type mockPaymentCanceller struct {
	output *port.PaymentOutput
	err    error
}

func (m *mockPaymentCanceller) Cancel(ctx context.Context, id string, reason string) (*port.PaymentOutput, error) {
	return m.output, m.err
}

// mockPaymentWebhookHandler implementa port.PaymentWebhookHandler para testes.
type mockPaymentWebhookHandler struct {
	err error
}

func (m *mockPaymentWebhookHandler) HandleWebhook(ctx context.Context, event port.WebhookEvent) error {
	return m.err
}
