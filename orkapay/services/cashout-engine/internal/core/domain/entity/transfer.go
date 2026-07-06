package entity

import (
	"time"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/event"
)

// TransferStatus representa o status de uma transferência.
type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "PENDING"
	TransferStatusApproved  TransferStatus = "APPROVED"
	TransferStatusSettled   TransferStatus = "SETTLED"
	TransferStatusFailed    TransferStatus = "FAILED"
	TransferStatusCancelled TransferStatus = "CANCELLED"
)

// CashOutProvider representa os provedores de cash-out suportados.
type CashOutProvider string

const (
	CashOutProviderStarkBank CashOutProvider = "STARK_BANK"
	CashOutProviderAsaas     CashOutProvider = "ASAAS"
	CashOutProviderCelcoin   CashOutProvider = "CELCOIN"
	CashOutProviderFitbank   CashOutProvider = "FITBANK"
)

// BankAccount representa os dados bancários do beneficiário.
type BankAccount struct {
	BankCode    string
	Branch      string
	Account     string
	AccountType string // CHECKING, SAVINGS, etc.
}

// TransferPayee representa o beneficiário da transferência.
type TransferPayee struct {
	Name      string
	TaxID     string
	PixKey    string
	PixType   string // CPF, CNPJ, EMAIL, PHONE, EVP
	Bank      *BankAccount
}

// Transfer é o aggregate raiz do domínio de cash-out.
type Transfer struct {
	id            string
	amountCents   int64
	status        TransferStatus
	payee         TransferPayee
	provider      CashOutProvider
	externalID    string
	description   string
	metadata      map[string]string
	processedAt   *time.Time
	settledAt     *time.Time
	failedAt      *time.Time
	cancelledAt   *time.Time
	failureReason string
	createdAt     time.Time
	updatedAt     time.Time
	version       int
	events        []event.CashOutEvent
}

// NewTransfer cria uma nova transferência com validações.
func NewTransfer(
	amountCents int64,
	payee TransferPayee,
	provider CashOutProvider,
	description string,
	metadata map[string]string,
) (*Transfer, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if payee.Name == "" || payee.TaxID == "" {
		return nil, domain.ErrInvalidPayee
	}
	if payee.PixKey == "" {
		return nil, domain.ErrInvalidPixKey
	}
	if provider == "" {
		return nil, domain.ErrInvalidProvider
	}

	now := time.Now().UTC()
	t := &Transfer{
		id:          generateID(),
		amountCents: amountCents,
		status:      TransferStatusPending,
		payee:       payee,
		provider:    provider,
		description: description,
		metadata:    copyMetadata(metadata),
		createdAt:   now,
		updatedAt:   now,
		version:     1,
	}

	t.events = append(t.events, event.TransferCreated{
		BaseCashOutEvent: event.NewBaseCashOutEvent("transfer.created", t.id),
		AmountCents:      amountCents,
		PayeeName:        payee.Name,
		PayeeTaxID:       payee.TaxID,
		PayeePixKey:      payee.PixKey,
		Provider:         string(provider),
		Description:      description,
	})

	return t, nil
}

// Getters
func (t *Transfer) ID() string              { return t.id }
func (t *Transfer) AmountCents() int64      { return t.amountCents }
func (t *Transfer) Status() TransferStatus  { return t.status }
func (t *Transfer) Payee() TransferPayee    { return t.payee }
func (t *Transfer) Provider() CashOutProvider { return t.provider }
func (t *Transfer) ExternalID() string      { return t.externalID }
func (t *Transfer) Description() string     { return t.description }
func (t *Transfer) Metadata() map[string]string {
	return copyMetadata(t.metadata)
}
func (t *Transfer) ProcessedAt() *time.Time  { return t.processedAt }
func (t *Transfer) SettledAt() *time.Time    { return t.settledAt }
func (t *Transfer) FailedAt() *time.Time     { return t.failedAt }
func (t *Transfer) FailureReason() string    { return t.failureReason }
func (t *Transfer) CreatedAt() time.Time     { return t.createdAt }
func (t *Transfer) UpdatedAt() time.Time     { return t.updatedAt }
func (t *Transfer) Version() int             { return t.version }
func (t *Transfer) Events() []event.CashOutEvent {
	result := make([]event.CashOutEvent, len(t.events))
	copy(result, t.events)
	return result
}

// Approve aprova a transferência junto ao provedor.
func (t *Transfer) Approve(externalID string) error {
	if t.status != TransferStatusPending {
		return domain.ErrTransferCannotBeCancelled
	}
	if externalID == "" {
		return domain.ErrInvalidProvider
	}

	now := time.Now().UTC()
	t.externalID = externalID
	t.status = TransferStatusApproved
	t.processedAt = &now
	t.version++
	t.updatedAt = now

	t.events = append(t.events, event.TransferApproved{
		BaseCashOutEvent: event.NewBaseCashOutEvent("transfer.approved", t.id),
		ExternalID:       externalID,
	})

	return nil
}

// Settle liquida a transferência.
func (t *Transfer) Settle() error {
	if t.status != TransferStatusApproved {
		return domain.ErrTransferCannotBeCancelled
	}

	now := time.Now().UTC()
	t.status = TransferStatusSettled
	t.settledAt = &now
	t.version++
	t.updatedAt = now

	t.events = append(t.events, event.TransferSettled{
		BaseCashOutEvent: event.NewBaseCashOutEvent("transfer.settled", t.id),
		SettledAt:        now.Format(time.RFC3339),
	})

	return nil
}

// Fail marca a transferência como falha.
func (t *Transfer) Fail(reason string) error {
	if t.status == TransferStatusSettled {
		return domain.ErrTransferAlreadySettled
	}
	if t.status == TransferStatusFailed {
		return domain.ErrTransferFailed
	}

	now := time.Now().UTC()
	t.status = TransferStatusFailed
	t.failedAt = &now
	t.failureReason = reason
	t.version++
	t.updatedAt = now

	t.events = append(t.events, event.TransferFailed{
		BaseCashOutEvent: event.NewBaseCashOutEvent("transfer.failed", t.id),
		FailureReason:    reason,
	})

	return nil
}

// Cancel cancela a transferência.
func (t *Transfer) Cancel(reason string) error {
	if t.status == TransferStatusSettled {
		return domain.ErrTransferAlreadySettled
	}
	if t.status == TransferStatusCancelled {
		return domain.ErrTransferCannotBeCancelled
	}

	now := time.Now().UTC()
	t.status = TransferStatusCancelled
	t.cancelledAt = &now
	t.failureReason = reason
	t.version++
	t.updatedAt = now

	t.events = append(t.events, event.TransferCancelled{
		BaseCashOutEvent: event.NewBaseCashOutEvent("transfer.cancelled", t.id),
		Reason:           reason,
	})

	return nil
}

// CanRetry verifica se a transferência pode ser retentada.
func (t *Transfer) CanRetry() bool {
	return t.status == TransferStatusFailed
}

// generateID gera um ID único para a transferência.
func generateID() string {
	return "cout-" + time.Now().Format("20060102150405") + "-" + randomSuffix(8)
}

// randomSuffix gera um sufixo aleatório.
func randomSuffix(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := range result {
		now := time.Now().UnixNano()
		result[i] = letters[(now/int64(i+1))%int64(len(letters))]
	}
	return string(result)
}

// copyMetadata retorna uma cópia do mapa de metadados.
func copyMetadata(m map[string]string) map[string]string {
	if m == nil {
		return make(map[string]string)
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
