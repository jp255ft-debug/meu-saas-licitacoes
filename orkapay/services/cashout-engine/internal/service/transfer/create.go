package transfer

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// CreateTransferUseCase orquestra a criação de uma nova transferência.
type CreateTransferUseCase struct {
	repo     outbound.TransferRepository
	provider outbound.CashOutProviderGateway
}

// NewCreateTransferUseCase cria uma nova instância do caso de uso.
func NewCreateTransferUseCase(repo outbound.TransferRepository, provider outbound.CashOutProviderGateway) *CreateTransferUseCase {
	return &CreateTransferUseCase{repo: repo, provider: provider}
}

// Execute executa o fluxo de criação de transferência.
func (uc *CreateTransferUseCase) Execute(ctx context.Context, input inbound.CreateTransferInput) (*inbound.TransferOutput, error) {
	// 1. Construir entidade do domínio
	payee := entity.TransferPayee{
		Name:   input.PayeeName,
		TaxID:  input.PayeeTaxID,
		PixKey: input.PayeePixKey,
		PixType: input.PayeePixType,
	}
	if input.PayeeBankCode != "" {
		payee.Bank = &entity.BankAccount{
			BankCode:    input.PayeeBankCode,
			Branch:      input.PayeeBranch,
			Account:     input.PayeeAccount,
			AccountType: input.PayeeAcctType,
		}
	}

	provider := entity.CashOutProvider(input.Provider)
	if provider == "" {
		provider = entity.CashOutProviderStarkBank
	}

	transfer, err := entity.NewTransfer(
		input.AmountCents,
		payee,
		provider,
		input.Description,
		input.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar transferência: %w", err)
	}

	// 2. Persistir com status PENDING
	if err := uc.repo.Save(ctx, transfer); err != nil {
		return nil, fmt.Errorf("falha ao salvar transferência: %w", err)
	}

	// 3. Criar no provedor externo
	externalID, err := uc.provider.CreateTransfer(ctx, transfer)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar transferência no provedor: %w", err)
	}

	// 4. Aprovar transferência com ID externo
	if err := transfer.Approve(externalID); err != nil {
		return nil, fmt.Errorf("falha ao aprovar transferência: %w", err)
	}

	// 5. Atualizar com optimistic locking
	if err := uc.repo.Update(ctx, transfer); err != nil {
		return nil, fmt.Errorf("falha ao atualizar transferência com dados do provedor: %w", err)
	}

	return MapToTransferOutput(transfer), nil
}