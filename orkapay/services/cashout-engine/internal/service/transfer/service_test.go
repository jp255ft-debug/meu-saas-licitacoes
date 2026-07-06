package transfer

import (
	"context"
	"errors"
	"testing"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockTransferRepo struct {
	saveFn            func(ctx context.Context, t *entity.Transfer) error
	findByIDFn        func(ctx context.Context, id string) (*entity.Transfer, error)
	findByExternalFn  func(ctx context.Context, externalID string) (*entity.Transfer, error)
	updateFn          func(ctx context.Context, t *entity.Transfer) error
	listByStatusFn    func(ctx context.Context, status entity.TransferStatus, limit, offset int) ([]*entity.Transfer, error)
	listByPayeeTaxFn  func(ctx context.Context, taxID string, limit, offset int) ([]*entity.Transfer, error)
	listExpiredFn     func(ctx context.Context, referenceTime interface{}, limit int) ([]*entity.Transfer, error)
	countByStatusFn   func(ctx context.Context, status entity.TransferStatus) (int64, error)
}

func (m *mockTransferRepo) Save(ctx context.Context, t *entity.Transfer) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, t)
	}
	return nil
}

func (m *mockTransferRepo) FindByID(ctx context.Context, id string) (*entity.Transfer, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTransferRepo) FindByExternalID(ctx context.Context, externalID string) (*entity.Transfer, error) {
	if m.findByExternalFn != nil {
		return m.findByExternalFn(ctx, externalID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTransferRepo) Update(ctx context.Context, t *entity.Transfer) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, t)
	}
	return nil
}

func (m *mockTransferRepo) ListByStatus(ctx context.Context, status entity.TransferStatus, limit, offset int) ([]*entity.Transfer, error) {
	if m.listByStatusFn != nil {
		return m.listByStatusFn(ctx, status, limit, offset)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTransferRepo) ListByPayeeTaxID(ctx context.Context, taxID string, limit, offset int) ([]*entity.Transfer, error) {
	if m.listByPayeeTaxFn != nil {
		return m.listByPayeeTaxFn(ctx, taxID, limit, offset)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTransferRepo) ListExpired(ctx context.Context, referenceTime interface{}, limit int) ([]*entity.Transfer, error) {
	if m.listExpiredFn != nil {
		return m.listExpiredFn(ctx, referenceTime, limit)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTransferRepo) CountByStatus(ctx context.Context, status entity.TransferStatus) (int64, error) {
	if m.countByStatusFn != nil {
		return m.countByStatusFn(ctx, status)
	}
	return 0, errors.New("not implemented")
}

var _ outbound.TransferRepository = (*mockTransferRepo)(nil)

type mockTransferProvider struct {
	createTransferFn    func(ctx context.Context, t *entity.Transfer) (externalID string, err error)
	getTransferStatusFn func(ctx context.Context, externalID string) (status string, err error)
	cancelTransferFn    func(ctx context.Context, externalID string) error
}

func (m *mockTransferProvider) CreateTransfer(ctx context.Context, t *entity.Transfer) (externalID string, err error) {
	if m.createTransferFn != nil {
		return m.createTransferFn(ctx, t)
	}
	return "", errors.New("not implemented")
}

func (m *mockTransferProvider) GetTransferStatus(ctx context.Context, externalID string) (status string, err error) {
	if m.getTransferStatusFn != nil {
		return m.getTransferStatusFn(ctx, externalID)
	}
	return "", errors.New("not implemented")
}

func (m *mockTransferProvider) CancelTransfer(ctx context.Context, externalID string) error {
	if m.cancelTransferFn != nil {
		return m.cancelTransferFn(ctx, externalID)
	}
	return errors.New("not implemented")
}

var _ outbound.CashOutProviderGateway = (*mockTransferProvider)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestTransfer(t *testing.T) *entity.Transfer {
	t.Helper()
	payee := entity.TransferPayee{
		Name:    "Jane Doe",
		TaxID:   "98765432101",
		PixKey:  "jane@example.com",
		PixType: "EMAIL",
	}
	tr, err := entity.NewTransfer(
		50000,
		payee,
		entity.CashOutProviderStarkBank,
		"Test transfer",
		map[string]string{"order": "456"},
	)
	assert.NoError(t, err)
	return tr
}

func newTestCreateInput() inbound.CreateTransferInput {
	return inbound.CreateTransferInput{
		AmountCents:  50000,
		PayeeName:    "Jane Doe",
		PayeeTaxID:   "98765432101",
		PayeePixKey:  "jane@example.com",
		PayeePixType: "EMAIL",
		Provider:     string(entity.CashOutProviderStarkBank),
		Description:  "Test transfer",
		Metadata:     map[string]string{"order": "456"},
	}
}

// ---------------------------------------------------------------------------
// Tests: CreateTransferUseCase
// ---------------------------------------------------------------------------

func TestCreateTransferUseCase_Success(t *testing.T) {
	repo := &mockTransferRepo{
		saveFn: func(ctx context.Context, tr *entity.Transfer) error {
			return nil
		},
		updateFn: func(ctx context.Context, tr *entity.Transfer) error {
			assert.Equal(t, entity.TransferStatusApproved, tr.Status())
			assert.NotEmpty(t, tr.ExternalID())
			return nil
		},
	}
	provider := &mockTransferProvider{
		createTransferFn: func(ctx context.Context, tr *entity.Transfer) (string, error) {
			return "ext_stark_123", nil
		},
	}

	uc := NewCreateTransferUseCase(repo, provider)
	output, err := uc.Execute(context.Background(), newTestCreateInput())

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "ext_stark_123", output.ExternalID)
	assert.Equal(t, string(entity.TransferStatusApproved), output.Status)
}

func TestCreateTransferUseCase_InvalidAmount(t *testing.T) {
	repo := &mockTransferRepo{}
	provider := &mockTransferProvider{}

	uc := NewCreateTransferUseCase(repo, provider)
	input := newTestCreateInput()
	input.AmountCents = -1

	_, err := uc.Execute(context.Background(), input)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestCreateTransferUseCase_SaveError(t *testing.T) {
	repo := &mockTransferRepo{
		saveFn: func(ctx context.Context, tr *entity.Transfer) error {
			return errors.New("erro ao salvar")
		},
	}
	provider := &mockTransferProvider{}

	uc := NewCreateTransferUseCase(repo, provider)
	_, err := uc.Execute(context.Background(), newTestCreateInput())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao salvar transferência")
}

func TestCreateTransferUseCase_ProviderError(t *testing.T) {
	repo := &mockTransferRepo{
		saveFn: func(ctx context.Context, tr *entity.Transfer) error {
			return nil
		},
		updateFn: func(ctx context.Context, tr *entity.Transfer) error {
			return nil
		},
	}
	provider := &mockTransferProvider{
		createTransferFn: func(ctx context.Context, tr *entity.Transfer) (string, error) {
			return "", errors.New("provedor indisponível")
		},
	}

	uc := NewCreateTransferUseCase(repo, provider)
	_, err := uc.Execute(context.Background(), newTestCreateInput())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao criar transferência no provedor")
}

// ---------------------------------------------------------------------------
// Tests: ApproveTransferUseCase
// ---------------------------------------------------------------------------

func TestApproveTransferUseCase_Success(t *testing.T) {
	tr := newTestTransfer(t)

	repo := &mockTransferRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transfer, error) {
			return tr, nil
		},
		updateFn: func(ctx context.Context, tr *entity.Transfer) error {
			assert.Equal(t, entity.TransferStatusApproved, tr.Status())
			return nil
		},
	}

	uc := NewApproveTransferUseCase(repo)
	output, err := uc.Execute(context.Background(), tr.ID(), "ext_approve")

	assert.NoError(t, err)
	assert.Equal(t, "ext_approve", output.ExternalID)
}

func TestApproveTransferUseCase_NotFound(t *testing.T) {
	repo := &mockTransferRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transfer, error) {
			return nil, domain.ErrTransferNotFound
		},
	}

	uc := NewApproveTransferUseCase(repo)
	_, err := uc.Execute(context.Background(), "non-existent", "ext")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTransferNotFound)
}

// ---------------------------------------------------------------------------
// Tests: SettleTransferUseCase
// ---------------------------------------------------------------------------

func TestSettleTransferUseCase_Success(t *testing.T) {
	tr := newTestTransfer(t)
	err := tr.Approve("ext_123")
	assert.NoError(t, err)

	repo := &mockTransferRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transfer, error) {
			return tr, nil
		},
		updateFn: func(ctx context.Context, upTr *entity.Transfer) error {
			assert.Equal(t, entity.TransferStatusSettled, upTr.Status())
			return nil
		},
	}

	uc := NewSettleTransferUseCase(repo)
	output, err := uc.Execute(context.Background(), tr.ID())

	assert.NoError(t, err)
	assert.Equal(t, string(entity.TransferStatusSettled), output.Status)
}

// ---------------------------------------------------------------------------
// Tests: FailTransferUseCase
// ---------------------------------------------------------------------------

func TestFailTransferUseCase_Success(t *testing.T) {
	tr := newTestTransfer(t)

	repo := &mockTransferRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transfer, error) {
			return tr, nil
		},
		updateFn: func(ctx context.Context, upTr *entity.Transfer) error {
			assert.Equal(t, entity.TransferStatusFailed, upTr.Status())
			return nil
		},
	}

	uc := NewFailTransferUseCase(repo)
	output, err := uc.Execute(context.Background(), tr.ID(), "insufficient balance")

	assert.NoError(t, err)
	assert.Equal(t, string(entity.TransferStatusFailed), output.Status)
	assert.Equal(t, "insufficient balance", output.FailureReason)
}

// ---------------------------------------------------------------------------
// Tests: CancelTransferUseCase
// ---------------------------------------------------------------------------

func TestCancelTransferUseCase_Success(t *testing.T) {
	tr := newTestTransfer(t)
	err := tr.Approve("ext_123")
	assert.NoError(t, err)

	repo := &mockTransferRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transfer, error) {
			return tr, nil
		},
		updateFn: func(ctx context.Context, upTr *entity.Transfer) error {
			assert.Equal(t, entity.TransferStatusCancelled, upTr.Status())
			return nil
		},
	}
	provider := &mockTransferProvider{
		cancelTransferFn: func(ctx context.Context, externalID string) error {
			return nil
		},
	}

	uc := NewCancelTransferUseCase(repo, provider)
	output, err := uc.Execute(context.Background(), tr.ID(), "user request")

	assert.NoError(t, err)
	assert.Equal(t, string(entity.TransferStatusCancelled), output.Status)
}

func TestCancelTransferUseCase_ProviderError(t *testing.T) {
	tr := newTestTransfer(t)
	err := tr.Approve("ext_123")
	assert.NoError(t, err)

	repo := &mockTransferRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transfer, error) {
			return tr, nil
		},
		updateFn: func(ctx context.Context, upTr *entity.Transfer) error {
			return nil
		},
	}
	provider := &mockTransferProvider{
		cancelTransferFn: func(ctx context.Context, externalID string) error {
			return errors.New("provedor offline")
		},
	}

	uc := NewCancelTransferUseCase(repo, provider)
	_, err = uc.Execute(context.Background(), tr.ID(), "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao cancelar transferência no provedor")
}

// ---------------------------------------------------------------------------
// Tests: GetTransferByIDUseCase
// ---------------------------------------------------------------------------

func TestGetTransferByIDUseCase_Success(t *testing.T) {
	tr := newTestTransfer(t)
	repo := &mockTransferRepo{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transfer, error) {
			return tr, nil
		},
	}

	uc := NewGetTransferByIDUseCase(repo)
	result, err := uc.Execute(context.Background(), tr.ID())

	assert.NoError(t, err)
	assert.Equal(t, tr.ID(), result.ID())
}

// ---------------------------------------------------------------------------
// Tests: ListTransfersByStatusUseCase
// ---------------------------------------------------------------------------

func TestListTransfersByStatusUseCase_Success(t *testing.T) {
	tr := newTestTransfer(t)
	repo := &mockTransferRepo{
		listByStatusFn: func(ctx context.Context, status entity.TransferStatus, limit, offset int) ([]*entity.Transfer, error) {
			return []*entity.Transfer{tr}, nil
		},
	}

	uc := NewListTransfersByStatusUseCase(repo)
	results, err := uc.Execute(context.Background(), string(entity.TransferStatusPending), 10, 0)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestListTransfersByStatusUseCase_InvalidStatus(t *testing.T) {
	repo := &mockTransferRepo{}

	uc := NewListTransfersByStatusUseCase(repo)
	_, err := uc.Execute(context.Background(), "", 10, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status inválido")
}

// ---------------------------------------------------------------------------
// Tests: ServiceImpl
// ---------------------------------------------------------------------------

func TestTransferServiceImpl_ImplementsInterface(t *testing.T) {
	repo := &mockTransferRepo{}
	provider := &mockTransferProvider{}
	svc := NewService(repo, provider)

	assert.Implements(t, (*inbound.TransferService)(nil), svc)
}

// ---------------------------------------------------------------------------
// Tests: Output mapping
// ---------------------------------------------------------------------------

func TestMapToTransferOutput(t *testing.T) {
	tr := newTestTransfer(t)
	output := MapToTransferOutput(tr)

	assert.Equal(t, tr.ID(), output.ID)
	assert.Equal(t, tr.AmountCents(), output.AmountCents)
	assert.Equal(t, string(tr.Status()), output.Status)
	assert.Equal(t, tr.Payee().Name, output.PayeeName)
	assert.Equal(t, tr.Payee().TaxID, output.PayeeTaxID)
	assert.NotEmpty(t, output.CreatedAt)
	assert.NotEmpty(t, output.UpdatedAt)
}