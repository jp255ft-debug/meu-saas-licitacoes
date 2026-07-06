package bsp

import (
	"context"
	"log/slog"
)

type MockSender struct{}

func NewMockSender() *MockSender {
	return &MockSender{}
}

func (m *MockSender) SendTextMessage(ctx context.Context, toNumber, message string) error {
	slog.InfoContext(ctx, "[MOCK] sending WhatsApp message",
		"to", toNumber,
		"message", message,
	)
	return nil
}

func (m *MockSender) SendPaymentLink(ctx context.Context, toNumber, paymentLink string, amount float64) error {
	msg := "🔗 *Link de pagamento gerado:*\n" + paymentLink
	return m.SendTextMessage(ctx, toNumber, msg)
}