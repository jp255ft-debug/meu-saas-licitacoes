package provider

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// AsaasGateway implementa outbound.CashInProviderGateway usando o cliente HTTP real do Asaas.
type AsaasGateway struct {
	client *AsaasClient
}

// NewAsaasGateway cria uma nova instância do gateway real Asaas.
func NewAsaasGateway(baseURL, apiKey, walletID string) *AsaasGateway {
	return &AsaasGateway{
		client: NewAsaasClient(baseURL, apiKey, walletID),
	}
}

// CreateInvoice cria uma fatura no Asaas com Split de Pagamento.
func (g *AsaasGateway) CreateInvoice(ctx context.Context, invoice *entity.Invoice) (externalID, pixQRCode, pixKey string, err error) {
	req := &CreateInvoiceRequest{
		BillingType: "PIX",
		DueDate:     invoice.DueDate().Format("2006-01-02"),
		Value:       float64(invoice.AmountCents()) / 100.0,
		Description: "",
		ExternalID:  invoice.ID(),
	}

	slog.InfoContext(ctx, "creating invoice on Asaas", "amount", req.Value, "due", req.DueDate)

	resp, err := g.client.CreateInvoice(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create invoice on Asaas", "error", err)
		return "", "", "", err
	}

	return resp.ID, resp.PixQRCode, resp.PixKey, nil
}

// GetInvoiceStatus consulta o status de uma fatura no Asaas.
func (g *AsaasGateway) GetInvoiceStatus(ctx context.Context, externalID string) (string, error) {
	resp, err := g.client.GetInvoice(ctx, externalID)
	if err != nil {
		return "", fmt.Errorf("failed to get invoice status: %w", err)
	}
	return resp.Status, nil
}

// CancelInvoice cancela uma fatura no Asaas.
func (g *AsaasGateway) CancelInvoice(ctx context.Context, externalID string) error {
	return g.client.CancelInvoice(ctx, externalID)
}

// RefundInvoice estorna uma fatura no Asaas.
func (g *AsaasGateway) RefundInvoice(ctx context.Context, externalID string) error {
	return fmt.Errorf("refund not implemented via API, use Asaas dashboard: %s", externalID)
}

// AsaasMock implementa outbound.CashInProviderGateway como mock (usado quando não configurado).
type AsaasMock struct {
	pendingStatus map[string]string
}

func NewAsaasMock() *AsaasMock {
	return &AsaasMock{pendingStatus: make(map[string]string)}
}

func (m *AsaasMock) CreateInvoice(ctx context.Context, invoice *entity.Invoice) (externalID, pixQRCode, pixKey string, err error) {
	externalID = fmt.Sprintf("asaas_inv_%s", invoice.ID())
	pixQRCode = fmt.Sprintf("pixqrcode:%s", invoice.ID())
	pixKey = fmt.Sprintf("pixkey:%s", invoice.ID())
	m.pendingStatus[externalID] = "PENDING"
	return externalID, pixQRCode, pixKey, nil
}

func (m *AsaasMock) GetInvoiceStatus(ctx context.Context, externalID string) (string, error) {
	s, ok := m.pendingStatus[externalID]
	if !ok {
		return "", fmt.Errorf("fatura não encontrada no Asaas: %s", externalID)
	}
	return s, nil
}

func (m *AsaasMock) CancelInvoice(ctx context.Context, externalID string) error {
	if _, ok := m.pendingStatus[externalID]; !ok {
		return fmt.Errorf("fatura não encontrada no Asaas: %s", externalID)
	}
	m.pendingStatus[externalID] = "CANCELLED"
	return nil
}

func (m *AsaasMock) RefundInvoice(ctx context.Context, externalID string) error {
	if _, ok := m.pendingStatus[externalID]; !ok {
		return fmt.Errorf("fatura não encontrada no Asaas: %s", externalID)
	}
	m.pendingStatus[externalID] = "REFUNDED"
	return nil
}

// compile-time checks
var _ outbound.CashInProviderGateway = (*AsaasGateway)(nil)
var _ outbound.CashInProviderGateway = (*AsaasMock)(nil)
