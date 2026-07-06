package entity

import (
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/event"
)

// InvoiceStatus representa o status de uma fatura/fatura de cash-in.
type InvoiceStatus string

const (
	InvoiceStatusPending   InvoiceStatus = "PENDING"
	InvoiceStatusIssued    InvoiceStatus = "ISSUED"
	InvoiceStatusPaid      InvoiceStatus = "PAID"
	InvoiceStatusCancelled InvoiceStatus = "CANCELLED"
	InvoiceStatusExpired   InvoiceStatus = "EXPIRED"
	InvoiceStatusRefunded  InvoiceStatus = "REFUNDED"
)

// CashInProvider representa os provedores de cash-in suportados.
type CashInProvider string

const (
	CashInProviderAsaas    CashInProvider = "ASAAS"
	CashInProviderStarkBank CashInProvider = "STARK_BANK"
	CashInProviderCelcoin  CashInProvider = "CELCOIN"
	CashInProviderFitbank  CashInProvider = "FITBANK"
)

// InvoicePayer representa o pagador da fatura.
type InvoicePayer struct {
	Name         string
	TaxID        string
	Email        string
	Phone        string
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	ZipCode      string
}

// Invoice é o aggregate raiz do domínio de cash-in.
type Invoice struct {
	id              string
	amountCents     int64
	status          InvoiceStatus
	dueDate         time.Time
	payer           InvoicePayer
	provider        CashInProvider
	externalID      string
	pixQRCode       string
	pixKey          string
	boletoBarcode   string
	boletoDigitable string
	metadata        map[string]string
	paidAmountCents int64
	paidAt          *time.Time
	paymentMethod   string
	cancelledAt     *time.Time
	failureReason   string
	createdAt       time.Time
	updatedAt       time.Time
	version         int
	events          []event.CashInEvent
}

// NewInvoice cria uma nova fatura com validações.
func NewInvoice(
	amountCents int64,
	dueDate time.Time,
	payer InvoicePayer,
	provider CashInProvider,
	metadata map[string]string,
) (*Invoice, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if dueDate.Before(time.Now().UTC().Truncate(24 * time.Hour)) {
		return nil, domain.ErrInvoiceExpired
	}
	if payer.Name == "" || payer.TaxID == "" {
		return nil, domain.ErrInvalidPayer
	}
	if provider == "" {
		return nil, domain.ErrInvalidProvider
	}

	now := time.Now().UTC()
	inv := &Invoice{
		id:          generateID(),
		status:      InvoiceStatusPending,
		amountCents: amountCents,
		dueDate:     dueDate,
		payer:       payer,
		provider:    provider,
		metadata:    copyMetadata(metadata),
		createdAt:   now,
		updatedAt:   now,
		version:     1,
	}

	inv.events = append(inv.events, event.InvoiceCreated{
		BaseCashInEvent: event.NewBaseCashInEvent("invoice.created", inv.id),
		AmountCents:     amountCents,
		PayerName:       payer.Name,
		PayerTaxID:      payer.TaxID,
		Provider:        string(provider),
		DueDate:         dueDate.Format("2006-01-02"),
	})

	return inv, nil
}

// Getters
func (inv *Invoice) ID() string              { return inv.id }
func (inv *Invoice) AmountCents() int64      { return inv.amountCents }
func (inv *Invoice) Status() InvoiceStatus   { return inv.status }
func (inv *Invoice) DueDate() time.Time      { return inv.dueDate }
func (inv *Invoice) Payer() InvoicePayer     { return inv.payer }
func (inv *Invoice) Provider() CashInProvider { return inv.provider }
func (inv *Invoice) ExternalID() string      { return inv.externalID }
func (inv *Invoice) PixQRCode() string       { return inv.pixQRCode }
func (inv *Invoice) PixKey() string          { return inv.pixKey }
func (inv *Invoice) BoletoBarcode() string   { return inv.boletoBarcode }
func (inv *Invoice) BoletoDigitable() string { return inv.boletoDigitable }
func (inv *Invoice) Metadata() map[string]string {
	return copyMetadata(inv.metadata)
}
func (inv *Invoice) PaidAmountCents() int64  { return inv.paidAmountCents }
func (inv *Invoice) PaidAt() *time.Time      { return inv.paidAt }
func (inv *Invoice) PaymentMethod() string   { return inv.paymentMethod }
func (inv *Invoice) FailureReason() string   { return inv.failureReason }
func (inv *Invoice) CreatedAt() time.Time    { return inv.createdAt }
func (inv *Invoice) UpdatedAt() time.Time    { return inv.updatedAt }
func (inv *Invoice) Version() int            { return inv.version }
func (inv *Invoice) Events() []event.CashInEvent {
	result := make([]event.CashInEvent, len(inv.events))
	copy(result, inv.events)
	return result
}

// Issue emite a fatura junto ao provedor.
func (inv *Invoice) Issue(externalID, pixQRCode, pixKey string) error {
	if inv.status != InvoiceStatusPending {
		return domain.ErrInvoiceCannotBeCancelled
	}
	if externalID == "" {
		return domain.ErrInvalidProvider
	}

	inv.externalID = externalID
	inv.pixQRCode = pixQRCode
	inv.pixKey = pixKey
	inv.status = InvoiceStatusIssued
	inv.version++
	inv.updatedAt = time.Now().UTC()

	return nil
}

// Pay registra o pagamento da fatura.
func (inv *Invoice) Pay(paidAmountCents int64, paymentMethod string) error {
	if inv.status == InvoiceStatusPaid {
		return domain.ErrInvoiceAlreadyPaid
	}
	if inv.status != InvoiceStatusIssued {
		return domain.ErrInvoiceCannotBeCancelled
	}
	if paidAmountCents <= 0 {
		return domain.ErrInvalidAmount
	}

	now := time.Now().UTC()
	inv.status = InvoiceStatusPaid
	inv.paidAmountCents = paidAmountCents
	inv.paidAt = &now
	inv.paymentMethod = paymentMethod
	inv.version++
	inv.updatedAt = now

	inv.events = append(inv.events, event.InvoicePaid{
		BaseCashInEvent: event.NewBaseCashInEvent("invoice.paid", inv.id),
		PaidAmountCents: paidAmountCents,
		PaidAt:          now.Format(time.RFC3339),
		PaymentMethod:   paymentMethod,
		ExternalID:      inv.externalID,
	})

	return nil
}

// Cancel cancela a fatura.
func (inv *Invoice) Cancel(reason string) error {
	if inv.status == InvoiceStatusPaid {
		return domain.ErrInvoiceAlreadyPaid
	}
	if inv.status == InvoiceStatusCancelled {
		return domain.ErrInvoiceCannotBeCancelled
	}

	now := time.Now().UTC()
	inv.status = InvoiceStatusCancelled
	inv.cancelledAt = &now
	inv.failureReason = reason
	inv.version++
	inv.updatedAt = now

	inv.events = append(inv.events, event.InvoiceCancelled{
		BaseCashInEvent: event.NewBaseCashInEvent("invoice.cancelled", inv.id),
		Reason:          reason,
	})

	return nil
}

// MarkExpired marca a fatura como expirada.
func (inv *Invoice) MarkExpired() error {
	if inv.status != InvoiceStatusPending && inv.status != InvoiceStatusIssued {
		return domain.ErrInvoiceCannotBeCancelled
	}

	now := time.Now().UTC()
	inv.status = InvoiceStatusExpired
	inv.version++
	inv.updatedAt = now

	inv.events = append(inv.events, event.InvoiceExpired{
		BaseCashInEvent: event.NewBaseCashInEvent("invoice.expired", inv.id),
		ExpiredAt:       now.Format(time.RFC3339),
	})

	return nil
}

// IsExpired verifica se a fatura está vencida.
func (inv *Invoice) IsExpired() bool {
	return time.Now().UTC().After(inv.dueDate)
}

// generateID gera um ID único para a fatura.
func generateID() string {
	return "cin-" + time.Now().Format("20060102150405") + "-" + randomSuffix(8)
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
