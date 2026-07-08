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

// getOrCreateCustomer busca um customer pelo CPF/CNPJ ou cria um novo no Asaas.
func (g *AsaasGateway) getOrCreateCustomer(ctx context.Context, name, taxID string) (*CustomerResponse, error) {
	// Tenta encontrar customer existente
	customer, err := g.client.FindCustomerByTaxID(ctx, taxID)
	if err != nil {
		return nil, fmt.Errorf("failed to find customer: %w", err)
	}
	if customer != nil {
		slog.InfoContext(ctx, "found existing customer", "id", customer.ID, "name", name)
		return customer, nil
	}

	// Cria novo customer
	req := &CreateCustomerRequest{
		Name:    name,
		CpfCnpj: taxID,
	}

	customer, err = g.client.CreateCustomer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	slog.InfoContext(ctx, "created customer on Asaas", "id", customer.ID, "name", name)
	return customer, nil
}

// CreateInvoice cria uma fatura no Asaas com Split de Pagamento.
func (g *AsaasGateway) CreateInvoice(ctx context.Context, invoice *entity.Invoice) (externalID, pixQRCode, pixKey string, err error) {
	// 1. Buscar ou criar customer no Asaas
	payer := invoice.Payer()
	customer, err := g.getOrCreateCustomer(ctx, payer.Name, payer.TaxID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get/customer: %w", err)
	}

	// 2. Montar request da invoice
	req := &CreateInvoiceRequest{
		Customer:    customer.ID,
		BillingType: "PIX",
		DueDate:     invoice.DueDate().Format("2006-01-02"),
		Value:       float64(invoice.AmountCents()) / 100.0,
		Description: "",
		ExternalID:  invoice.ID(),
	}

	slog.InfoContext(ctx, "creating invoice on Asaas",
		"customer", customer.ID,
		"amount", req.Value,
		"due", req.DueDate,
	)

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