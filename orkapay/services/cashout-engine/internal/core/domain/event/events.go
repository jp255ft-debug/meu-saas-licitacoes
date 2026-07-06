package event

import "time"

// CashOutEvent é a interface base para eventos de cash-out.
type CashOutEvent interface {
	EventName() string
	OccurredAt() time.Time
	TransferID() string
}

// BaseCashOutEvent implementa a interface CashOutEvent.
type BaseCashOutEvent struct {
	name       string
	timestamp  time.Time
	transferID string
}

func NewBaseCashOutEvent(name, transferID string) BaseCashOutEvent {
	return BaseCashOutEvent{
		name:       name,
		timestamp:  time.Now().UTC(),
		transferID: transferID,
	}
}

func (e BaseCashOutEvent) EventName() string     { return e.name }
func (e BaseCashOutEvent) OccurredAt() time.Time  { return e.timestamp }
func (e BaseCashOutEvent) TransferID() string     { return e.transferID }

// Eventos de cash-out
type (
	TransferCreated struct {
		BaseCashOutEvent
		AmountCents   int64
		PayeeName     string
		PayeeTaxID    string
		PayeePixKey   string
		Provider      string
		Description   string
	}

	TransferApproved struct {
		BaseCashOutEvent
		ExternalID string
	}

	TransferSettled struct {
		BaseCashOutEvent
		SettledAt string
	}

	TransferFailed struct {
		BaseCashOutEvent
		FailureReason string
	}

	TransferCancelled struct {
		BaseCashOutEvent
		Reason string
	}
)
