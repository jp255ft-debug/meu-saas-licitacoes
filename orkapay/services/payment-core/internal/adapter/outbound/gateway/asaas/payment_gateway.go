package asaas

import (
	"context"
	"fmt"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// PaymentGateway implementa port.PaymentGateway usando a API do Asaas.
type PaymentGateway struct {
	client *Client
}

// NewPaymentGateway cria uma nova instancia do gateway Asaas.
func NewPaymentGateway(client *Client) port.PaymentGateway {
	return &PaymentGateway{client: client}
}

// mapStatus converte status Asaas para status de dominio.
func mapStatus(asaasStatus string) string {
	switch asaasStatus {
	case "PENDING":
		return string(domain.PaymentStatusPending)
	case "RECEIVED":
		return string(domain.PaymentStatusApproved)
	case "CONFIRMED", "RECEIVED_IN_CASH":
		return string(domain.PaymentStatusSettled)
	case "OVERDUE":
		return string(domain.PaymentStatusFailed)
	case "CANCELLED":
		return string(domain.PaymentStatusCancelled)
	case "REFUNDED":
		return "REFUNDED"
	default:
		return string(domain.PaymentStatusPending)
	}
}

// CreatePayment cria um pagamento no Asaas.
func (g *PaymentGateway) CreatePayment(ctx context.Context, payment *domain.Payment) (externalID string, status string, err error) {
	dueDate := time.Now().UTC().Add(72 * time.Hour).Format("2006-01-02")
	if payment.ExpiresAt() != nil {
		dueDate = payment.ExpiresAt().Format("2006-01-02")
	}

	req := CreatePaymentRequest{
		Customer:          payment.Payee().TaxID,
		BillingType:       "PIX",
		Value:             payment.Amount().Cents(),
		DueDate:           dueDate,
		Description:       fmt.Sprintf("Pagamento %s - %s", payment.ID(), payment.Description()),
		ExternalReference: payment.ID(),
	}

	var resp CreatePaymentResponse
	if err := g.client.doRequest(ctx, "POST", "/v3/payments", &req, &resp); err != nil {
		return "", "", err
	}

	return resp.ID, mapStatus(resp.Status), nil
}

// GetPaymentStatus consulta o status de um pagamento no Asaas.
func (g *PaymentGateway) GetPaymentStatus(ctx context.Context, externalID string) (status string, err error) {
	var resp GetPaymentResponse
	if err := g.client.doRequest(ctx, "GET", fmt.Sprintf("/v3/payments/%s", externalID), nil, &resp); err != nil {
		return "", err
	}
	return mapStatus(resp.Status), nil
}

// CancelPayment cancela um pagamento no Asaas.
func (g *PaymentGateway) CancelPayment(ctx context.Context, externalID string) error {
	return g.client.doRequest(ctx, "DELETE", fmt.Sprintf("/v3/payments/%s", externalID), nil, nil)
}

// RefundPayment estorna um pagamento no Asaas.
func (g *PaymentGateway) RefundPayment(ctx context.Context, externalID string) error {
	return g.client.doRequest(ctx, "POST", fmt.Sprintf("/v3/payments/%s/refund", externalID), nil, nil)
}

// SetWebhookStatus mantem o status do webhook no gateway.
func (g *PaymentGateway) SetWebhookStatus(ctx context.Context, externalID string) error {
	_, err := g.GetPaymentStatus(ctx, externalID)
	return err
}

// HealthCheck verifica a saude da API do Asaas.
func (g *PaymentGateway) HealthCheck(ctx context.Context) error {
	return g.client.doRequest(ctx, "GET", "/v3/health", nil, nil)
}