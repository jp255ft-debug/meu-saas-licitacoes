package application

import (
	"context"
	"errors"
	"testing"

	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPaymentUseCase_GetByID_Success(t *testing.T) {
	payment := newTestPayment(t)
	repo := &mockPaymentRepository{payment: payment}
	uc := NewGetPaymentUseCase(repo)

	result, err := uc.GetByID(context.Background(), payment.ID())

	require.NoError(t, err)
	assert.Equal(t, payment.ID(), result.ID())
	assert.Equal(t, payment.Status(), result.Status())
}

func TestGetPaymentUseCase_GetByID_NotFound(t *testing.T) {
	repo := &mockPaymentRepository{findErr: errors.New("not found")}
	uc := NewGetPaymentUseCase(repo)

	result, err := uc.GetByID(context.Background(), "non-existent-id")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "falha ao buscar")
}

func TestGetPaymentUseCase_GetByID_RepositoryError(t *testing.T) {
	repo := &mockPaymentRepository{findErr: errors.New("db connection failed")}
	uc := NewGetPaymentUseCase(repo)

	result, err := uc.GetByID(context.Background(), "any-id")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "falha ao buscar")
}
