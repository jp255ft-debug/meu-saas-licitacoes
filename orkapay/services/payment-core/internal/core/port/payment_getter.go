package port

import (
	"context"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
)

// PaymentGetter define a busca de pagamentos por ID.
type PaymentGetter interface {
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
}
