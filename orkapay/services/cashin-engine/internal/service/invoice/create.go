package invoice

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// CreateInvoiceUseCase orquestra a criação de uma nova fatura.
type CreateInvoiceUseCase struct {
	repo     outbound.InvoiceRepository
	provider outbound.CashInProviderGateway
}

// NewCreateInvoiceUseCase cria uma nova instância do caso de uso.
func NewCreateInvoiceUseCase(repo outbound.InvoiceRepository, provider outbound.CashInProviderGateway) *CreateInvoiceUseCase {
	return &CreateInvoiceUseCase{repo: repo, provider: provider}
}

// Execute executa o fluxo de criação de fatura.
func (uc *CreateInvoiceUseCase) Execute(ctx context.Context, input inbound.CreateInvoiceInput) (*inbound.InvoiceOutput, error) {
	// 1. Construir entidade do domínio
	payer := entity.InvoicePayer{
		Name:    input.PayerName,
		TaxID:   input.PayerTaxID,
		Email:   input.PayerEmail,
		Phone:   input.PayerPhone,
	}

	provider := entity.CashInProvider(input.Provider)
	if provider == "" {
		provider = entity.CashInProviderAsaas
	}

	invoice, err := entity.NewInvoice(
		input.AmountCents,
		input.DueDate,
		payer,
		provider,
		input.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar fatura: %w", err)
	}

	// 2. Persistir com status PENDING
	if err := uc.repo.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("falha ao salvar fatura: %w", err)
	}

	// 3. Criar no provedor externo
	externalID, pixQRCode, pixKey, err := uc.provider.CreateInvoice(ctx, invoice)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar fatura no provedor: %w", err)
	}

	// 4. Emitir fatura com dados do provedor
	if err := invoice.Issue(externalID, pixQRCode, pixKey); err != nil {
		return nil, fmt.Errorf("falha ao emitir fatura: %w", err)
	}

	// 5. Atualizar com optimistic locking
	if err := uc.repo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("falha ao atualizar fatura com dados do provedor: %w", err)
	}

	return MapToInvoiceOutput(invoice), nil
}