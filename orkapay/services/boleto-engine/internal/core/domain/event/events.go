package event

import "time"

// BoletoEvent é a interface base para eventos de boleto.
type BoletoEvent interface {
	EventName() string
	OccurredAt() time.Time
	BoletoID() string
}

// BaseBoletoEvent implementa a interface BoletoEvent.
type BaseBoletoEvent struct {
	name      string
	timestamp time.Time
	boletoID  string
}

func NewBaseBoletoEvent(name, boletoID string) BaseBoletoEvent {
	return BaseBoletoEvent{
		name:      name,
		timestamp: time.Now().UTC(),
		boletoID:  boletoID,
	}
}

func (e BaseBoletoEvent) EventName() string  { return e.name }
func (e BaseBoletoEvent) OccurredAt() time.Time { return e.timestamp }
func (e BaseBoletoEvent) BoletoID() string    { return e.boletoID }

// Eventos de boleto
type (
	BoletoIssued struct {
		BaseBoletoEvent
		BoletoNumber string
		Barcode      string
		Digitable    string
		AmountCents  int64
		DueDate      string
		Provider     string
	}

	BoletoPaid struct {
		BaseBoletoEvent
		PaidAmountCents int64
		PaidAt          string
		PaymentID       string
	}

	BoletoCancelled struct {
		BaseBoletoEvent
		Reason string
	}

	BoletoExpired struct {
		BaseBoletoEvent
		ExpiredAt string
	}

	BoletoRefunded struct {
		BaseBoletoEvent
		RefundAmountCents int64
		RefundReason      string
	}
)
