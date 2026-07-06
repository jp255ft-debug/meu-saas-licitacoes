package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// ExpireInvoiceUseCase orquestra a expiração de uma fatura vencida.
type ExpireInvoiceUseCase struct {
	repo outbound.InvoiceRepository
}

// NewExpireInvoiceUseCase cria uma nova instância do caso de uso.
func NewExpireInvoiceUseCase(repo outbound.InvoiceRepository) *ExpireInvoiceUseCase {
	return &ExpireInvoiceUseCase{repo: repo}
}

// Execute marca uma fatura como expirada.
func (uc *ExpireInvoiceUseCase) Execute(ctx context.Context, id string) (*inbound.InvoiceOutput, error) {
	invoice, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar fatura: %w", err)
	}

	if invoice.Status() != entity.InvoiceStatusPending && invoice.Status() != entity.InvoiceStatusIssued {
		return nil, domain.ErrInvoiceCannotBeCancelled
	}

	// Atualiza diretamente via repositório
	if err := uc.repo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("falha ao atualizar fatura expirada: %w", err)
	}

	return MapToInvoiceOutput(invoice), nil
}

// ProcessExpiredInvoicesUseCase processa em lote todas as faturas expiradas.
type ProcessExpiredInvoicesUseCase struct {
	repo outbound.InvoiceRepository
}

func NewProcessExpiredInvoicesUseCase(repo outbound.InvoiceRepository) *ProcessExpiredInvoicesUseCase {
	return &ProcessExpiredInvoicesUseCase{repo: repo}
}

func (uc *ProcessExpiredInvoicesUseCase) Execute(ctx context.Context) (int, error) {
	expired, err := uc.repo.ListExpired(ctx, interface{}(time.Now().UTC()), 100)
	if err != nil {
		return 0, fmt.Errorf("falha ao listar faturas expiradas: %w", err)
	}
	count := 0
	for _, inv := range expired {
		if err := uc.repo.Update(ctx, inv); err != nil {
			continue
		}
		count++
	}
	return count, nil
}