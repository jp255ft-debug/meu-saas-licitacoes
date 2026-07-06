package inbound

import (
	"context"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
)

// CreateTransferInput é o input para criação de transferência.
type CreateTransferInput struct {
	AmountCents   int64
	PayeeName     string
	PayeeTaxID    string
	PayeePixKey   string
	PayeePixType  string
	PayeeBankCode string
	PayeeBranch   string
	PayeeAccount  string
	PayeeAcctType string
	Provider      string
	Description   string
	Metadata      map[string]string
}

// TransferOutput é a saída padronizada para operações com transferência.
type TransferOutput struct {
	ID            string
	AmountCents   int64
	Status        string
	PayeeName     string
	PayeeTaxID    string
	PayeePixKey   string
	Provider      string
	ExternalID    string
	Description   string
	FailureReason string
	CreatedAt     string
	UpdatedAt     string
}

// TransferService define os casos de uso do domínio de cash-out.
type TransferService interface {
	Create(ctx context.Context, input CreateTransferInput) (*TransferOutput, error)
	Approve(ctx context.Context, id, externalID string) (*TransferOutput, error)
	GetByID(ctx context.Context, id string) (*entity.Transfer, error)
	Settle(ctx context.Context, id string) (*TransferOutput, error)
	Fail(ctx context.Context, id, reason string) (*TransferOutput, error)
	Cancel(ctx context.Context, id, reason string) (*TransferOutput, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entity.Transfer, error)
}
