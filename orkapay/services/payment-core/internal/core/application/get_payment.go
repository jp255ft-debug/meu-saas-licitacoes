package application

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// GetPaymentUseCase orquestra a busca de um pagamento por ID.
type GetPaymentUseCase struct {
	repo port.PaymentRepository
}

// NewGetPaymentUseCase cria uma nova instância do caso de uso.
func NewGetPaymentUseCase(repo port.PaymentRepository) *GetPaymentUseCase {
	return &GetPaymentUseCase{
		repo: repo,
	}
}

// GetByID busca um pagamento pelo ID interno.
func (uc *GetPaymentUseCase) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	payment, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar pagamento: %w", err)
	}
	return payment, nil
}
