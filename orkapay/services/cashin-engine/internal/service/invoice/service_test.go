package invoice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockInvoiceRepo struct {
	saveFn            func(ctx context.Context, inv *entity.Invoice) error
	findByIDFn        func(ctx context.Context, id string) (*entity.Invoice, error)
	findByExternalFn  func(ctx context.Context, externalID string) (*entity.Invoice, error)
	updateFn          func(ctx context.Context, inv *entity.Invoice) error
	listByStatusFn    func(ctx context.Context, status entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error)
	listByPayerTaxFn  func(ctx context.Context, taxID string, limit, offset int) ([]*entity.Invoice, error)
	listExpiredFn     func(ctx context.Context, referenceTime interface{}, limit int) ([]*entity.Invoice, error)
	countByStatusFn   func(ctx context.Context, status entity.InvoiceStatus) (int64, error)
}

func (m *mockInvoiceRepo) Save(ctx context.Context, inv *entity.Invoice) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, inv)
	}
	return nil
}

func (m *mockInvoiceRepo) FindByID(ctx context.Context, id string) (*entity.Invoice, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockInvoiceRepo) FindByExternalID(ctx context.Context, externalID string) (*entity.Invoice, error) {
	if m.findByExternalFn != nil {
		return m.findByExternalFn(ctx, externalID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockInvoiceRepo) Update(ctx context.Context, inv *entity.Invoice) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, inv)
	}
	return nil
}

func (m *mockInvoiceRepo) ListByStatus(ctx context.Context, status entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error) {
	if m.listByStatusFn != nil {
		return m.listByStatusFn(ctx, status, limit, offset)
	}
	return nil, errors.New("not implemented")
}

func (m *mockInvoiceRepo) ListByPayerTaxID(ctx context.Context, taxID string, limit, offset int) ([]*entity.Invoice, error) {
	if m.listByPayerTaxFn != nil {
		return m.listByPayerTaxFn(ctx, taxID, limit, offset)
	}
	return nil, errors.New("not implemented")
}

func (m *mockInvoiceRepo) ListExpired(ctx context.Context, referenceTime interface{}, limit int) ([]*entity.Invoice, error) {
	if m.listExpiredFn != nil {
		return m.listExpiredFn(ctx, referenceTime, limit)
	}
	return nil, errors.New("not implemented")
}

func (m *mockInvoiceRepo) CountByStatus(ctx context.Context, status entity.InvoiceStatus) (int64, error) {
	if m.countByStatusFn != nil {
		return m.countByStatusFn(ctx, status)
	}
	return 0, errors.New("not implemented")
}

var _ outbound.InvoiceRepository = (*mockInvoiceRepo)(nil)

type mockInvoiceProvider struct {
	createInvoiceFn    func(ctx context.Context, inv *entity.Invoice) (externalID, pixQRCode, pixKey string, err error)
	getInvoiceStatusFn func(ctx context.Context, externalID string) (status string, err error)
	cancelInvoiceFn    func(ctx context.Context, externalID string) error
	refundInvoiceFn    func(ctx context.Context, externalID string) error
}

func (m *mockInvoiceProvider) CreateInvoice(ctx context.Context, inv *entity.Invoice) (externalID, pixQRCode, pixKey string, err error) {
	if m.createInvoiceFn != nil {
		return m.createInvoiceFn(ctx, inv)
	}
	return "", "", "", errors.New("not implemented")
}

func (m *mockInvoiceProvider) GetInvoiceStatus(ctx context.Context, externalID string) (status string, err error) {
	if m.getInvoiceStatusFn != nil {
		return m.getInvoiceStatusFn(ctx, externalID)
	}
	return "", errors.New("not implemented")
}

func (m *mockInvoiceProvider) CancelInvoice(ctx context.Context, externalID string) error {
	if m.cancelInvoiceFn != nil {
		return m.cancelInvoiceFn(ctx, externalID)
	}
	return errors.New("not implemented")
}

func (m *mockInvoiceProvider) RefundInvoice(ctx context.Context, externalID string) error {
	if m.refundInvoiceFn != nil {
		return m.refundInvoiceFn(ctx, externalID)
	}
	return errors.New("not implemented")
}

var _ outbound.CashInProviderGateway = (*mockInvoiceProvider)(nil)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newTestInvoice(t *testing.T) *entity.Invoice {
	t.Helper()
	payer := entity.InvoicePayer{
		Name:  "John Doe",
		TaxID: "12345678901",
		Email: "john@example.com",
	}
	inv, err := entity.NewInvoice(
		10000,
		time.Now().Add(72*time.Hour),
		payer,
		entity.CashInProviderAsaas,
		map[string]string{"order": "123"},
	)
	assert.NoError(t, err)
	return inv
}

func newTestCreateInput() inbound.CreateInvoiceInput {
	return inbound.CreateInvoiceInput{
		AmountCents: 10000,
		DueDate:     time.Now().Add(72 * time.Hour),
		PayerName:   "John Doe",
		PayerTaxID:  "12345678901",
		PayerEmail:  "john@example.com",
		Provider:    string(entity.CashInProviderAsaas),
		Metadata:    map[string]string{"order": "123"},
	}
}

// ---------------------------------------------------------------------------
// Tests: CreateInvoiceUseCase
// ---------------------------------------------------------------------------

func TestCreateInvoiceUseCase_Success(t *testing.T) {
	repo := &mockInvoiceRepo{
		saveFn: func(ctx context.Context, inv *entity.Invoice) error {
			return nil
		},
		updateFn: func(ctx context.Context, inv *entity.Invoice) error {
			assert.Equal(t, entity.InvoiceStatusIssued, inv.Status())
			assert.NotEmpty(t, inv.ExternalID())
			return nil
		},
	}
	provider := &mockInvoiceProvider{
		createInvoiceFn: func(ctx context.Context, inv *entity.Invoice) (string, string, string, error) {
			return "ext_123", "pixqrcode:abc", "pixkey:abc", nil
		},
	}

	uc := NewCreateInvoiceUseCase(repo, provider)
	output, err := uc.Execute(context.Background(), newTestCreateInput())

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "ext_123", output.ExternalID)
	assert.Equal(t, "pixqrcode:abc", output.PixQRCode)
	assert.Equal(t, "pixkey:abc", output.PixKey)
	assert.Equal(t, string(entity.InvoiceStatusIssued), output.Status)
}

func TestCreateInvoiceUseCase_InvalidInput(t *testing.T) {
	repo := &mockInvoiceRepo{}
	provider := &mockInvoiceProvider{}

	uc := NewCreateInvoiceUseCase(repo, provider)
	input := newTestCreateInput()
	input.AmountCents = 0 // inválido

	_, err := uc.Execute(context.Background(), input)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestCreateInvoiceUseCase_SaveError(t *testing.T) {
	repo := &mockInvoiceRepo{
		saveFn: func(ctx context.Context, inv *entity.Invoice) error {
			return errors.New("erro ao salvar")
		},
	}
	provider := &mockInvoiceProvider{}

	uc := NewCreateInvoiceUseCase(repo, provider)
	_, err := uc.Execute(context.Background(), newTestCreateInput())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao salvar fatura")
}

func TestCreateInvoiceUseCase_ProviderError(t *testing.T) {
	repo := &mockInvoiceRepo{
		saveFn: func(ctx context.Context, inv *entity.Invoice) error {
			return nil
		},
		updateFn: func(ctx context.Context, inv *entity.Invoice) error {
			return nil
		},
	}
	provider := &mockInvoiceProvider{
		createInvoiceFn: func(ctx context.Context, inv *entity.Invoice) (string, string, string, error) {
			return "", "", "", errors.New("provedor indisponível")
		},
	}

	uc := NewCreateInvoiceUseCase(repo, provider)
	_, err := uc.Execute(context.Background(), newTestCreateInput())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao criar fatura no provedor")
}

func TestCreateInvoiceUseCase_OptimisticLockError(t *testing.T) {
	repo := &mockInvoiceRepo{
		saveFn: func(ctx context.Context, inv *entity.Invoice) error {
			return nil
		},
		updateFn: func(ctx context.Context, inv *entity.Invoice) error {
			return errors.New("fatura já foi modificada por outra operação")
		},
	}
	provider := &mockInvoiceProvider{
		createInvoiceFn: func(ctx context.Context, inv *entity.Invoice) (string, string, string, error) {
			return "ext_123", "pixqrcode:abc", "pixkey:abc", nil
		},
	}

	uc := NewCreateInvoiceUseCase(repo, provider)
	_, err := uc.Execute(context.Background(), newTestCreateInput())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao atualizar fatura com dados do provedor")
}

// ---------------------------------------------------------------------------
// Tests: IssueInvoiceUseCase
// ---------------------------------------------------------------------------

func TestIssueInvoiceUseCase_Success(t *testing.T) {
	inv := newTestInvoice(t)

	repo := &mockInvoiceRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Invoice, error) {
			return inv, nil
		},
		updateFn: func(ctx context.Context, i *entity.Invoice) error {
			assert.Equal(t, entity.InvoiceStatusIssued, i.Status())
			return nil
		},
	}
	provider := &mockInvoiceProvider{
		createInvoiceFn: func(ctx context.Context, i *entity.Invoice) (string, string, string, error) {
			return "ext_issue", "qrcode:issue", "pixkey:issue", nil
		},
	}

	uc := NewIssueInvoiceUseCase(repo, provider)
	output, err := uc.Execute(context.Background(), inv.ID())

	assert.NoError(t, err)
	assert.Equal(t, "ext_issue", output.ExternalID)
}

func TestIssueInvoiceUseCase_NotFound(t *testing.T) {
	repo := &mockInvoiceRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Invoice, error) {
			return nil, domain.ErrInvoiceNotFound
		},
	}
	provider := &mockInvoiceProvider{}

	uc := NewIssueInvoiceUseCase(repo, provider)
	_, err := uc.Execute(context.Background(), "non-existent")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvoiceNotFound)
}

// ---------------------------------------------------------------------------
// Tests: PayInvoiceUseCase
// ---------------------------------------------------------------------------

func TestPayInvoiceUseCase_Success(t *testing.T) {
	inv := newTestInvoice(t)
	err := inv.Issue("ext_123", "qrcode", "pixkey")
	assert.NoError(t, err)

	repo := &mockInvoiceRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Invoice, error) {
			return inv, nil
		},
		updateFn: func(ctx context.Context, i *entity.Invoice) error {
			assert.Equal(t, entity.InvoiceStatusPaid, i.Status())
			return nil
		},
	}

	uc := NewPayInvoiceUseCase(repo)
	output, err := uc.Execute(context.Background(), inv.ID(), 10000, "PIX")

	assert.NoError(t, err)
	assert.Equal(t, string(entity.InvoiceStatusPaid), output.Status)
}

// ---------------------------------------------------------------------------
// Tests: CancelInvoiceUseCase
// ---------------------------------------------------------------------------

func TestCancelInvoiceUseCase_Success(t *testing.T) {
	inv := newTestInvoice(t)
	err := inv.Issue("ext_123", "qrcode", "pixkey")
	assert.NoError(t, err)

	repo := &mockInvoiceRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Invoice, error) {
			return inv, nil
		},
		updateFn: func(ctx context.Context, i *entity.Invoice) error {
			assert.Equal(t, entity.InvoiceStatusCancelled, i.Status())
			return nil
		},
	}
	provider := &mockInvoiceProvider{
		cancelInvoiceFn: func(ctx context.Context, externalID string) error {
			return nil
		},
	}

	uc := NewCancelInvoiceUseCase(repo, provider)
	output, err := uc.Execute(context.Background(), inv.ID(), "customer request")

	assert.NoError(t, err)
	assert.Equal(t, string(entity.InvoiceStatusCancelled), output.Status)
}

func TestCancelInvoiceUseCase_ProviderError(t *testing.T) {
	inv := newTestInvoice(t)
	err := inv.Issue("ext_123", "qrcode", "pixkey")
	assert.NoError(t, err)

	repo := &mockInvoiceRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Invoice, error) {
			return inv, nil
		},
		updateFn: func(ctx context.Context, i *entity.Invoice) error {
			return nil
		},
	}
	provider := &mockInvoiceProvider{
		cancelInvoiceFn: func(ctx context.Context, externalID string) error {
			return errors.New("provedor offline")
		},
	}

	uc := NewCancelInvoiceUseCase(repo, provider)
	_, err = uc.Execute(context.Background(), inv.ID(), "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao cancelar fatura no provedor")
}

// ---------------------------------------------------------------------------
// Tests: GetInvoiceByIDUseCase
// ---------------------------------------------------------------------------

func TestGetInvoiceByIDUseCase_Success(t *testing.T) {
	inv := newTestInvoice(t)
	repo := &mockInvoiceRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Invoice, error) {
			return inv, nil
		},
	}

	uc := NewGetInvoiceByIDUseCase(repo)
	result, err := uc.Execute(context.Background(), inv.ID())

	assert.NoError(t, err)
	assert.Equal(t, inv.ID(), result.ID())
}

// ---------------------------------------------------------------------------
// Tests: ListInvoicesByStatusUseCase
// ---------------------------------------------------------------------------

func TestListInvoicesByStatusUseCase_Success(t *testing.T) {
	inv := newTestInvoice(t)
	repo := &mockInvoiceRepo{
		listByStatusFn: func(ctx context.Context, status entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error) {
			return []*entity.Invoice{inv}, nil
		},
	}

	uc := NewListInvoicesByStatusUseCase(repo)
	results, err := uc.Execute(context.Background(), string(entity.InvoiceStatusPending), 10, 0)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestListInvoicesByStatusUseCase_InvalidStatus(t *testing.T) {
	repo := &mockInvoiceRepo{}

	uc := NewListInvoicesByStatusUseCase(repo)
	_, err := uc.Execute(context.Background(), "", 10, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status inválido")
}

// ---------------------------------------------------------------------------
// Tests: ServiceImpl
// ---------------------------------------------------------------------------

func TestServiceImpl_ImplementsInterface(t *testing.T) {
	repo := &mockInvoiceRepo{}
	provider := &mockInvoiceProvider{}
	svc := NewService(repo, provider)

	assert.Implements(t, (*inbound.InvoiceService)(nil), svc)
}

// ---------------------------------------------------------------------------
// Tests: Output mapping
// ---------------------------------------------------------------------------

func TestMapToInvoiceOutput(t *testing.T) {
	inv := newTestInvoice(t)
	output := MapToInvoiceOutput(inv)

	assert.Equal(t, inv.ID(), output.ID)
	assert.Equal(t, inv.AmountCents(), output.AmountCents)
	assert.Equal(t, string(inv.Status()), output.Status)
	assert.Equal(t, inv.Payer().Name, output.PayerName)
	assert.Equal(t, inv.Payer().TaxID, output.PayerTaxID)
	assert.NotEmpty(t, output.CreatedAt)
	assert.NotEmpty(t, output.UpdatedAt)
}