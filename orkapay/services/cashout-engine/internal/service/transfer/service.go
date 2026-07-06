package transfer

import (
	"context"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// ServiceImpl implementa inbound.TransferService orquestrando os use cases.
type ServiceImpl struct {
	createUseCase  *CreateTransferUseCase
	approveUseCase *ApproveTransferUseCase
	settleUseCase  *SettleTransferUseCase
	failUseCase    *FailTransferUseCase
	cancelUseCase  *CancelTransferUseCase
	getByIDUseCase *GetTransferByIDUseCase
	listByStatusUC *ListTransfersByStatusUseCase
}

// NewService cria uma nova instância do serviço de transferência.
func NewService(repo outbound.TransferRepository, provider outbound.CashOutProviderGateway) *ServiceImpl {
	return &ServiceImpl{
		createUseCase:  NewCreateTransferUseCase(repo, provider),
		approveUseCase: NewApproveTransferUseCase(repo),
		settleUseCase:  NewSettleTransferUseCase(repo),
		failUseCase:    NewFailTransferUseCase(repo),
		cancelUseCase:  NewCancelTransferUseCase(repo, provider),
		getByIDUseCase: NewGetTransferByIDUseCase(repo),
		listByStatusUC: NewListTransfersByStatusUseCase(repo),
	}
}

// Create cria uma nova transferência.
func (s *ServiceImpl) Create(ctx context.Context, input inbound.CreateTransferInput) (*inbound.TransferOutput, error) {
	return s.createUseCase.Execute(ctx, input)
}

// Approve aprova uma transferência.
func (s *ServiceImpl) Approve(ctx context.Context, id, externalID string) (*inbound.TransferOutput, error) {
	return s.approveUseCase.Execute(ctx, id, externalID)
}

// Settle liquida uma transferência.
func (s *ServiceImpl) Settle(ctx context.Context, id string) (*inbound.TransferOutput, error) {
	return s.settleUseCase.Execute(ctx, id)
}

// Fail marca uma transferência como falha.
func (s *ServiceImpl) Fail(ctx context.Context, id, reason string) (*inbound.TransferOutput, error) {
	return s.failUseCase.Execute(ctx, id, reason)
}

// Cancel cancela uma transferência.
func (s *ServiceImpl) Cancel(ctx context.Context, id, reason string) (*inbound.TransferOutput, error) {
	return s.cancelUseCase.Execute(ctx, id, reason)
}

// GetByID busca uma transferência pelo ID.
func (s *ServiceImpl) GetByID(ctx context.Context, id string) (*entity.Transfer, error) {
	return s.getByIDUseCase.Execute(ctx, id)
}

// ListByStatus lista transferências por status.
func (s *ServiceImpl) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entity.Transfer, error) {
	return s.listByStatusUC.Execute(ctx, status, limit, offset)
}

// Compile-time check
var _ inbound.TransferService = (*ServiceImpl)(nil)