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

func TestProcessPaymentUseCase_Process_Success(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{
		response: &port.GatewayResponse{ExternalID: "ext-123"},
	}
	uc := NewProcessPaymentUseCase(repo, pub, gw)

	output, err := uc.Process(context.Background(), payment.ID())

	require.NoError(t, err)
	assert.Equal(t, payment.ID(), output.ID)
	assert.Equal(t, "APPROVED", output.Status)
	assert.True(t, repo.updated)
	assert.Len(t, pub.publishedEvents, 2)
}

func TestProcessPaymentUseCase_Process_NotFound(t *testing.T) {
	repo := &mockPaymentRepository{findErr: errors.New("not found")}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{}
	uc := NewProcessPaymentUseCase(repo, pub, gw)

	output, err := uc.Process(context.Background(), "non-existent-id")

	require.Error(t, err)
	assert.Nil(t, output)
}

func TestProcessPaymentUseCase_Process_NotPending(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123")
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{}
	uc := NewProcessPaymentUseCase(repo, pub, gw)

	output, err := uc.Process(context.Background(), payment.ID())

	require.Error(t, err)
	assert.Nil(t, output)
	assert.ErrorIs(t, err, domain.ErrPaymentNotPending)
}

func TestProcessPaymentUseCase_Process_GatewayFails(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{createErr: errors.New("gateway timeout")}
	uc := NewProcessPaymentUseCase(repo, pub, gw)

	output, err := uc.Process(context.Background(), payment.ID())

	require.Error(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "FAILED", output.Status)
	assert.True(t, repo.updated)
	assert.Len(t, pub.publishedEvents, 2)
	assert.Contains(t, err.Error(), "gateway rejeitou")
}

func TestProcessPaymentUseCase_Process_GatewayFails_UpdateFails(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{
		payment:   payment,
		updateErr: errors.New("update error"),
	}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{createErr: errors.New("gateway timeout")}
	uc := NewProcessPaymentUseCase(repo, pub, gw)

	output, err := uc.Process(context.Background(), payment.ID())

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "salvar")
}

func TestProcessPaymentUseCase_Process_ApproveFails(t *testing.T) {
	payment := newTestPayment(t)
	payment.Approve("ext-123") // already approved
	payment.Cancel("test")
	repo := &mockPaymentRepository{payment: payment}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{
		response: &port.GatewayResponse{ExternalID: "ext-456"},
	}
	uc := NewProcessPaymentUseCase(repo, pub, gw)

	output, err := uc.Process(context.Background(), payment.ID())

	require.Error(t, err)
	assert.Nil(t, output)
}

func TestProcessPaymentUseCase_Process_UpdateFails(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{
		payment:   payment,
		updateErr: errors.New("update error"),
	}
	pub := &mockEventPublisher{}
	gw := &mockPaymentGateway{
		response: &port.GatewayResponse{ExternalID: "ext-123"},
	}
	uc := NewProcessPaymentUseCase(repo, pub, gw)

	output, err := uc.Process(context.Background(), payment.ID())

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "atualizar")
}
