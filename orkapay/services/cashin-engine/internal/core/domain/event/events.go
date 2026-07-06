package event

import "time"

// CashInEvent é a interface base para eventos de cash-in.
type CashInEvent interface {
	EventName() string
	OccurredAt() time.Time
	InvoiceID() string
}

// BaseCashInEvent implementa a interface CashInEvent.
type BaseCashInEvent struct {
	name      string
	timestamp time.Time
	invoiceID string
}

func NewBaseCashInEvent(name, invoiceID string) BaseCashInEvent {
	return BaseCashInEvent{
		name:      name,
		timestamp: time.Now().UTC(),
		invoiceID: invoiceID,
	}
}

func (e BaseCashInEvent) EventName() string    { return e.name }
func (e BaseCashInEvent) OccurredAt() time.Time { return e.timestamp }
func (e BaseCashInEvent) InvoiceID() string     { return e.invoiceID }

// Eventos de cash-in
type (
	InvoiceCreated struct {
		BaseCashInEvent
		AmountCents int64
		PayerName   string
		PayerTaxID  string
		Provider    string
		DueDate     string
	}

	InvoicePaid struct {
		BaseCashInEvent
		PaidAmountCents int64
		PaidAt          string
		PaymentMethod   string
		ExternalID      string
	}

	InvoiceCancelled struct {
		BaseCashInEvent
		Reason string
	}

	InvoiceExpired struct {
		BaseCashInEvent
		ExpiredAt string
	}

	InvoiceRefunded struct {
		BaseCashInEvent
		RefundAmountCents int64
		RefundReason      string
	}
)
