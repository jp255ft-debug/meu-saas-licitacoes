package outbound

import (
	"context"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
)

// TransferRepository define o contrato para persistência de transferências.
type TransferRepository interface {
	Save(ctx context.Context, transfer *entity.Transfer) error
	FindByID(ctx context.Context, id string) (*entity.Transfer, error)
	FindByExternalID(ctx context.Context, externalID string) (*entity.Transfer, error)
	Update(ctx context.Context, transfer *entity.Transfer) error
	ListByStatus(ctx context.Context, status entity.TransferStatus, limit, offset int) ([]*entity.Transfer, error)
	CountByStatus(ctx context.Context, status entity.TransferStatus) (int64, error)
}

// CashOutProviderGateway define o contrato para integração com provedores de cash-out.
type CashOutProviderGateway interface {
	CreateTransfer(ctx context.Context, transfer *entity.Transfer) (externalID string, err error)
	GetTransferStatus(ctx context.Context, externalID string) (status string, err error)
	CancelTransfer(ctx context.Context, externalID string) error
}
