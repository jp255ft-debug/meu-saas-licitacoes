package application

import (
	"context"
	"errors"
	"testing"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookPaymentUseCase_HandleWebhook_Received(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: "ext-123",
		Status:     "RECEIVED",
	}

	err := uc.HandleWebhook(context.Background(), event)

	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusSettled, payment.Status())
	assert.True(t, repo.updated)
	assert.True(t, len(pub.publishedEvents) >= 1)
}

func TestWebhookPaymentUseCase_HandleWebhook_Confirmed(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: "ext-123",
		Status:     "CONFIRMED",
	}

	err := uc.HandleWebhook(context.Background(), event)

	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusSettled, payment.Status())
	assert.True(t, repo.updated)
}

func TestWebhookPaymentUseCase_HandleWebhook_Overdue(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: "ext-123",
		Status:     "OVERDUE",
	}

	err := uc.HandleWebhook(context.Background(), event)

	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusFailed, payment.Status())
	assert.True(t, repo.updated)
	assert.True(t, len(pub.publishedEvents) >= 1)
}

func TestWebhookPaymentUseCase_HandleWebhook_Cancelled(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: payment.ExternalID(),
		Status:     "CANCELLED",
	}

	err := uc.HandleWebhook(context.Background(), event)

	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusCancelled, payment.Status())
	assert.True(t, repo.updated)
}

func TestWebhookPaymentUseCase_HandleWebhook_Refunded(t *testing.T) {
	// REFUNDED webhook: payment was settled, then refunded by gateway
	// Cancel() only works on PENDING, so we simulate a different flow
	// For now, test that webhook handles unknown/future statuses gracefully
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	payment.Settle()
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: "ext-123",
		Status:     "REFUNDED",
	}

	// This will fail because Cancel doesn't work on SETTLED
	// The webhook handler should handle this gracefully
	err := uc.HandleWebhook(context.Background(), event)

	// Currently the domain doesn't support refund after settlement
	// So we expect an error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelar")
}

func TestWebhookPaymentUseCase_HandleWebhook_UnknownStatus(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: "ext-123",
		Status:     "UNKNOWN_STATUS",
	}

	err := uc.HandleWebhook(context.Background(), event)

	require.NoError(t, err)
	assert.False(t, repo.updated)
	assert.Len(t, pub.publishedEvents, 0)
}

func TestWebhookPaymentUseCase_HandleWebhook_NotFound(t *testing.T) {
	repo := &mockPaymentRepository{findErr: errors.New("not found")}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: "ext-123",
		Status:     "RECEIVED",
	}

	err := uc.HandleWebhook(context.Background(), event)

	require.Error(t, err)
	assert.Error(t, err)
}

func TestWebhookPaymentUseCase_HandleWebhook_SettleFails(t *testing.T) {
	payment := newTestPayment(t)
	// Payment is pending (not approved), so Settle will fail
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: payment.ExternalID(),
		Status:     "RECEIVED",
	}

	err := uc.HandleWebhook(context.Background(), event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "liquidar")
}

func TestWebhookPaymentUseCase_HandleWebhook_UpdateFails(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	repo := &mockPaymentRepository{
		payment:   payment,
		updateErr: errors.New("update error"),
	}
	pub := &mockEventPublisher{}
	uc := NewWebhookPaymentUseCase(repo, pub)

	event := port.WebhookEvent{
		ExternalID: "ext-123",
		Status:     "RECEIVED",
	}

	err := uc.HandleWebhook(context.Background(), event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "atualizar")
}
