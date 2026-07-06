package provider

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// AsaasMock implementa outbound.CashInProviderGateway como um mock do provedor Asaas.
type AsaasMock struct {
	pendingStatus map[string]string // externalID -> status
}

// NewAsaasMock cria uma nova instância do mock Asaas.
func NewAsaasMock() *AsaasMock {
	return &AsaasMock{
		pendingStatus: make(map[string]string),
	}
}

// CreateInvoice simula a criação de uma fatura no Asaas.
func (m *AsaasMock) CreateInvoice(ctx context.Context, invoice *entity.Invoice) (externalID, pixQRCode, pixKey string, err error) {
	externalID = fmt.Sprintf("asaas_inv_%s", invoice.ID())
	pixQRCode = fmt.Sprintf("pixqrcode:%s", invoice.ID())
	pixKey = fmt.Sprintf("pixkey:%s", invoice.ID())

	m.pendingStatus[externalID] = "PENDING"

	return externalID, pixQRCode, pixKey, nil
}

// GetInvoiceStatus simula a consulta de status de uma fatura no Asaas.
func (m *AsaasMock) GetInvoiceStatus(ctx context.Context, externalID string) (status string, err error) {
	s, ok := m.pendingStatus[externalID]
	if !ok {
		return "", fmt.Errorf("fatura não encontrada no Asaas: %s", externalID)
	}
	return s, nil
}

// CancelInvoice simula o cancelamento de uma fatura no Asaas.
func (m *AsaasMock) CancelInvoice(ctx context.Context, externalID string) error {
	if _, ok := m.pendingStatus[externalID]; !ok {
		return fmt.Errorf("fatura não encontrada no Asaas: %s", externalID)
	}
	m.pendingStatus[externalID] = "CANCELLED"
	return nil
}

// RefundInvoice simula o estorno de uma fatura no Asaas.
func (m *AsaasMock) RefundInvoice(ctx context.Context, externalID string) error {
	if _, ok := m.pendingStatus[externalID]; !ok {
		return fmt.Errorf("fatura não encontrada no Asaas: %s", externalID)
	}
	m.pendingStatus[externalID] = "REFUNDED"
	return nil
}

// compile-time check
var _ outbound.CashInProviderGateway = (*AsaasMock)(nil)