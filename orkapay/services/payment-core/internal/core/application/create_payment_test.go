package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/core/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePaymentUseCase_Create_Success(t *testing.T) {
	repo := &mockPaymentRepository{}
	pub := &mockEventPublisher{}
	uc := NewCreatePaymentUseCase(repo, pub)

	input := newTestCreateInput()
	output, err := uc.Create(context.Background(), input)

	require.NoError(t, err)
	assert.NotEmpty(t, output.ID)
	assert.Equal(t, "PENDING", output.Status)
	assert.Equal(t, "R$ 100,00", output.Amount)
	assert.True(t, repo.saved)
	assert.True(t, len(pub.publishedEvents) >= 1)
}

func TestCreatePaymentUseCase_Create_InvalidAmount(t *testing.T) {
	repo := &mockPaymentRepository{}
	pub := &mockEventPublisher{}
	uc := NewCreatePaymentUseCase(repo, pub)

	input := newTestCreateInput()
	input.AmountCents = -1

	output, err := uc.Create(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "valor")
	assert.False(t, repo.saved)
}

func TestCreatePaymentUseCase_Create_InvalidPayerPixKey(t *testing.T) {
	repo := &mockPaymentRepository{}
	pub := &mockEventPublisher{}
	uc := NewCreatePaymentUseCase(repo, pub)

	input := newTestCreateInput()
	input.PayerPixKey = "invalid-key-!!!"

	output, err := uc.Create(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "chave Pix")
	assert.False(t, repo.saved)
}

func TestCreatePaymentUseCase_Create_InvalidPayeePixKey(t *testing.T) {
	repo := &mockPaymentRepository{}
	pub := &mockEventPublisher{}
	uc := NewCreatePaymentUseCase(repo, pub)

	input := newTestCreateInput()
	input.PayeePixKey = "invalid-key-!!!"

	output, err := uc.Create(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "chave Pix")
	assert.False(t, repo.saved)
}

func TestCreatePaymentUseCase_Create_SaveFails(t *testing.T) {
	repo := &mockPaymentRepository{saveErr: errors.New("db error")}
	pub := &mockEventPublisher{}
	uc := NewCreatePaymentUseCase(repo, pub)

	input := newTestCreateInput()
	output, err := uc.Create(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "salvar")
	assert.True(t, repo.saved)
}

func TestCreatePaymentUseCase_Create_PublishFails(t *testing.T) {
	repo := &mockPaymentRepository{}
	pub := &mockEventPublisher{publishErr: errors.New("kafka error")}
	uc := NewCreatePaymentUseCase(repo, pub)

	input := newTestCreateInput()
	output, err := uc.Create(context.Background(), input)

	// Publish failure should not fail the operation
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, repo.saved)
	assert.Len(t, pub.publishedEvents, 0)
}

func TestCreatePaymentUseCase_MapToOutput(t *testing.T) {
	payment := newTestPayment(t)
	output := MapToOutput(payment)

	assert.Equal(t, payment.ID(), output.ID)
	assert.Equal(t, string(payment.Status()), output.Status)
	assert.Equal(t, payment.Amount().String(), output.Amount)
	assert.Equal(t, payment.Fee().String(), output.Fee)
	assert.Equal(t, payment.Payer().Name, output.PayerName)
	assert.Equal(t, payment.Payee().Name, output.PayeeName)
}

func TestCreatePaymentUseCase_FormatTime(t *testing.T) {
	now := time.Now()
	assert.Equal(t, now.Format(time.RFC3339), port.FormatTime(&now))
	assert.Equal(t, "", port.FormatTime(nil))
}
