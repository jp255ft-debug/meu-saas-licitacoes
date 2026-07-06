package provider

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// StarkBankMock implementa outbound.CashOutProviderGateway como um mock do provedor Stark Bank.
type StarkBankMock struct {
	pendingStatus map[string]string // externalID -> status
}

// NewStarkBankMock cria uma nova instância do mock Stark Bank.
func NewStarkBankMock() *StarkBankMock {
	return &StarkBankMock{
		pendingStatus: make(map[string]string),
	}
}

// CreateTransfer simula a criação de uma transferência no Stark Bank.
func (m *StarkBankMock) CreateTransfer(ctx context.Context, transfer *entity.Transfer) (externalID string, err error) {
	externalID = fmt.Sprintf("stark_tr_%s", transfer.ID())
	m.pendingStatus[externalID] = "PROCESSING"
	return externalID, nil
}

// GetTransferStatus simula a consulta de status de uma transferência no Stark Bank.
func (m *StarkBankMock) GetTransferStatus(ctx context.Context, externalID string) (status string, err error) {
	s, ok := m.pendingStatus[externalID]
	if !ok {
		return "", fmt.Errorf("transferência não encontrada no Stark Bank: %s", externalID)
	}
	return s, nil
}

// CancelTransfer simula o cancelamento de uma transferência no Stark Bank.
func (m *StarkBankMock) CancelTransfer(ctx context.Context, externalID string) error {
	if _, ok := m.pendingStatus[externalID]; !ok {
		return fmt.Errorf("transferência não encontrada no Stark Bank: %s", externalID)
	}
	m.pendingStatus[externalID] = "CANCELLED"
	return nil
}

// compile-time check
var _ outbound.CashOutProviderGateway = (*StarkBankMock)(nil)