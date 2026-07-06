package application

import (
	"context"
	"errors"
	"testing"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancelPaymentUseCase_Cancel_Success(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{}
	uc := NewCancelPaymentUseCase(repo, pub, gw)

	output, err := uc.Cancel(context.Background(), payment.ID(), "requested by user")

	require.NoError(t, err)
	assert.Equal(t, payment.ID(), output.ID)
	assert.Equal(t, "CANCELLED", output.Status)
	assert.True(t, repo.updated)
	assert.Len(t, pub.publishedEvents, 2)
}

func TestCancelPaymentUseCase_Cancel_WithExternalID(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{}
	uc := NewCancelPaymentUseCase(repo, pub, gw)

	output, err := uc.Cancel(context.Background(), payment.ID(), "requested by user")

	require.NoError(t, err)
	assert.Equal(t, "CANCELLED", output.Status)
	assert.True(t, repo.updated)
}

func TestCancelPaymentUseCase_Cancel_NotFound(t *testing.T) {
	repo := &mockPaymentRepository{findErr: errors.New("not found")}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{}
	uc := NewCancelPaymentUseCase(repo, pub, gw)

	output, err := uc.Cancel(context.Background(), "non-existent", "reason")

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Error(t, err)
}

func TestCancelPaymentUseCase_Cancel_NotPending(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	payment.Settle()
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{}
	uc := NewCancelPaymentUseCase(repo, pub, gw)

	output, err := uc.Cancel(context.Background(), payment.ID(), "reason")

	require.Error(t, err)
	assert.Nil(t, output)
	assert.ErrorIs(t, err, domain.ErrPaymentCannotBeCancelled)
}

func TestCancelPaymentUseCase_Cancel_GatewayFails(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{cancelErr: errors.New("gateway timeout")}
	uc := NewCancelPaymentUseCase(repo, pub, gw)

	output, err := uc.Cancel(context.Background(), payment.ID(), "reason")

	// Gateway failure should not stop the cancellation
	require.NoError(t, err)
	assert.Equal(t, "CANCELLED", output.Status)
	assert.True(t, repo.updated)
}

func TestCancelPaymentUseCase_Cancel_DomainFails(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	payment.Settle()
	// After settle, payment is not pending, so Cancel should fail
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{}
	uc := NewCancelPaymentUseCase(repo, pub, gw)

	output, err := uc.Cancel(context.Background(), payment.ID(), "reason")

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Error(t, err)
}

func TestCancelPaymentUseCase_Cancel_UpdateFails(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{
		payment:   payment,
		updateErr: errors.New("update error"),
	}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{}
	uc := NewCancelPaymentUseCase(repo, pub, gw)

	output, err := uc.Cancel(context.Background(), payment.ID(), "reason")

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "atualizar")
}
