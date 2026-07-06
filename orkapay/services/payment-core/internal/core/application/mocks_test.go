package application

import (
	"context"
	"testing"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// mockPaymentRepository implementa port.PaymentRepository para testes.
type mockPaymentRepository struct {
	saveErr   error
	findErr   error
	updateErr error
	payment   *domain.Payment
	payments  []*domain.Payment
	count     int64

	saved   bool
	updated bool
}

func (m *mockPaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	m.saved = true
	return m.saveErr
}

func (m *mockPaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.payment, nil
}

func (m *mockPaymentRepository) FindByExternalID(ctx context.Context, externalID string) (*domain.Payment, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.payment, nil
}

func (m *mockPaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	m.updated = true
	return m.updateErr
}

func (m *mockPaymentRepository) ListByStatus(ctx context.Context, status domain.PaymentStatus, limit, offset int) ([]*domain.Payment, error) {
	return m.payments, nil
}

func (m *mockPaymentRepository) ListByPayerTaxID(ctx context.Context, taxID string, limit, offset int) ([]*domain.Payment, error) {
	return m.payments, nil
}

func (m *mockPaymentRepository) ListExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Payment, error) {
	return m.payments, nil
}

func (m *mockPaymentRepository) CountByStatus(ctx context.Context, status domain.PaymentStatus) (int64, error) {
	return m.count, nil
}

// mockEventPublisher implementa port.EventPublisher para testes.
type mockEventPublisher struct {
	publishErr      error
	publishedEvents []domain.DomainEvent
	closed          bool
}

func (m *mockEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventPublisher) Close() error {
	m.closed = true
	return nil
}

// mockPaymentGateway implementa port.PaymentGateway para testes.
type mockPaymentGateway struct {
	createErr  error
	cancelErr  error
	statusErr  error
	refundErr  error
	healthErr  error
	response   *port.GatewayResponse
	status     *port.GatewayStatus
	gatewayName string
}

func (m *mockPaymentGateway) CreatePayment(ctx context.Context, payment *domain.Payment) (*port.GatewayResponse, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.response, nil
}

func (m *mockPaymentGateway) GetPaymentStatus(ctx context.Context, externalID string) (*port.GatewayStatus, error) {
	if m.statusErr != nil {
		return nil, m.statusErr
	}
	return m.status, nil
}

func (m *mockPaymentGateway) CancelPayment(ctx context.Context, externalID string, reason string) error {
	if m.cancelErr != nil {
		return m.cancelErr
	}
	return nil
}

func (m *mockPaymentGateway) RefundPayment(ctx context.Context, externalID string, amount domain.Money, reason string) (*port.GatewayResponse, error) {
	if m.refundErr != nil {
		return nil, m.refundErr
	}
	return m.response, nil
}

func (m *mockPaymentGateway) HealthCheck(ctx context.Context) error {
	return m.healthErr
}

func (m *mockPaymentGateway) Name() string {
	if m.gatewayName != "" {
		return m.gatewayName
	}
	return "mock"
}

// helper para criar um pagamento válido para testes.
func newTestPayment(t testing.TB) *domain.Payment {
	amount, err := domain.NewMoney(10000)
	if err != nil {
		t.Fatalf("failed to create money: %v", err)
	}

	payerPixKey := domain.MustNewPixKey("test@example.com")
	payeePixKey := domain.MustNewPixKey("payee@example.com")

	payer := domain.Payer{
		Name:   "John Doe",
		TaxID:  "12345678901",
		Email:  "john@example.com",
		PixKey: payerPixKey,
	}

	payee := domain.Payee{
		Name:   "Jane Doe",
		TaxID:  "98765432101",
		PixKey: payeePixKey,
	}

	payment, err := domain.NewPayment(
		domain.PaymentTypePix,
		domain.PaymentMethodPixQRCode,
		amount,
		payer,
		payee,
		"Test payment",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	return payment
}

// helper para criar input válido de criação.
func newTestCreateInput() port.CreatePaymentInput {
	return port.CreatePaymentInput{
		PaymentType: string(domain.PaymentTypePix),
		Method:      string(domain.PaymentMethodPixQRCode),
		AmountCents: 10000,
		PayerName:   "John Doe",
		PayerTaxID:  "12345678901",
		PayerEmail:  "john@example.com",
		PayerPixKey: "test@example.com",
		PayeeName:   "Jane Doe",
		PayeeTaxID:  "98765432101",
		PayeePixKey: "payee@example.com",
		Description: "Test payment",
	}
}
