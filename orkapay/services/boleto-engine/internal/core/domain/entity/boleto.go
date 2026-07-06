package entity

import (
	"time"

	"github.com/jp255ft-debug/boleto-engine/internal/core/domain"
	"github.com/jp255ft-debug/boleto-engine/internal/core/domain/event"
)

// BoletoStatus representa o status de um boleto.
type BoletoStatus string

const (
	BoletoStatusPending   BoletoStatus = "PENDING"
	BoletoStatusIssued    BoletoStatus = "ISSUED"
	BoletoStatusPaid      BoletoStatus = "PAID"
	BoletoStatusCancelled BoletoStatus = "CANCELLED"
	BoletoStatusExpired   BoletoStatus = "EXPIRED"
	BoletoStatusRefunded  BoletoStatus = "REFUNDED"
)

// BoletoProvider representa os provedores de boleto suportados.
type BoletoProvider string

const (
	ProviderAsaas    BoletoProvider = "ASAAS"
	ProviderStarkBank BoletoProvider = "STARK_BANK"
	ProviderCelcoin  BoletoProvider = "CELCOIN"
	ProviderFitbank  BoletoProvider = "FITBANK"
)

// BoletoPayer representa o pagador do boleto.
type BoletoPayer struct {
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

// BoletoFee representa as taxas/configurações financeiras do boleto.
type BoletoFee struct {
	FineRate       float64 // Percentual de multa (ex: 2.0 = 2%)
	InterestRate   float64 // Percentual de juros ao dia (ex: 0.033 = 0.033%)
	DiscountDays   int     // Dias para desconto
	DiscountRate   float64 // Percentual de desconto
}

// Boleto é o aggregate raiz do domínio de boletos.
type Boleto struct {
	id              string
	boletoNumber    string
	barcode         string
	digitable       string
	status          BoletoStatus
	amountCents     int64
	fee             BoletoFee
	dueDate         time.Time
	payer           BoletoPayer
	provider        BoletoProvider
	externalID      string
	bankAccountID   string
	metadata        map[string]string
	paidAmountCents int64
	paidAt          *time.Time
	cancelledAt     *time.Time
	failureReason   string
	createdAt       time.Time
	updatedAt       time.Time
	version         int
	events          []event.BoletoEvent
}

// NewBoleto cria um novo boleto com validações.
func NewBoleto(
	boletoNumber string,
	amountCents int64,
	dueDate time.Time,
	payer BoletoPayer,
	provider BoletoProvider,
	metadata map[string]string,
) (*Boleto, error) {
	if boletoNumber == "" {
		return nil, domain.ErrInvalidBoletoNumber
	}
	if amountCents <= 0 {
		return nil, domain.ErrInvalidBoletoNumber
	}
	if dueDate.Before(time.Now().UTC().Truncate(24 * time.Hour)) {
		return nil, domain.ErrInvalidDueDate
	}
	if payer.Name == "" || payer.TaxID == "" {
		return nil, domain.ErrInvalidPayerInfo
	}
	if provider == "" {
		return nil, domain.ErrInvalidProvider
	}

	now := time.Now().UTC()
	b := &Boleto{
		id:           generateID(),
		boletoNumber: boletoNumber,
		status:       BoletoStatusPending,
		amountCents:  amountCents,
		dueDate:      dueDate,
		payer:        payer,
		provider:     provider,
		metadata:     copyMetadata(metadata),
		createdAt:    now,
		updatedAt:    now,
		version:      1,
	}

	return b, nil
}

// Getters
func (b *Boleto) ID() string              { return b.id }
func (b *Boleto) BoletoNumber() string    { return b.boletoNumber }
func (b *Boleto) Barcode() string         { return b.barcode }
func (b *Boleto) Digitable() string       { return b.digitable }
func (b *Boleto) Status() BoletoStatus    { return b.status }
func (b *Boleto) AmountCents() int64      { return b.amountCents }
func (b *Boleto) Fee() BoletoFee          { return b.fee }
func (b *Boleto) DueDate() time.Time      { return b.dueDate }
func (b *Boleto) Payer() BoletoPayer      { return b.payer }
func (b *Boleto) Provider() BoletoProvider { return b.provider }
func (b *Boleto) ExternalID() string      { return b.externalID }
func (b *Boleto) Metadata() map[string]string {
	return copyMetadata(b.metadata)
}
func (b *Boleto) PaidAmountCents() int64  { return b.paidAmountCents }
func (b *Boleto) PaidAt() *time.Time      { return b.paidAt }
func (b *Boleto) FailureReason() string   { return b.failureReason }
func (b *Boleto) CreatedAt() time.Time    { return b.createdAt }
func (b *Boleto) UpdatedAt() time.Time    { return b.updatedAt }
func (b *Boleto) Version() int            { return b.version }
func (b *Boleto) Events() []event.BoletoEvent {
	result := make([]event.BoletoEvent, len(b.events))
	copy(result, b.events)
	return result
}

// Issue emite o boleto junto ao provedor, gerando código de barras e linha digitável.
func (b *Boleto) Issue(barcode, digitable, externalID string) error {
	if b.status != BoletoStatusPending {
		return domain.ErrBoletoCannotBeCancelled
	}
	if barcode == "" || digitable == "" {
		return domain.ErrInvalidBarcode
	}

	b.barcode = barcode
	b.digitable = digitable
	b.externalID = externalID
	b.status = BoletoStatusIssued
	b.version++
	b.updatedAt = time.Now().UTC()

	b.events = append(b.events, event.BoletoIssued{
		BaseBoletoEvent: event.NewBaseBoletoEvent("boleto.issued", b.id),
		BoletoNumber:    b.boletoNumber,
		Barcode:         barcode,
		Digitable:       digitable,
		AmountCents:     b.amountCents,
		DueDate:         b.dueDate.Format("2006-01-02"),
		Provider:        string(b.provider),
	})

	return nil
}

// Pay registra o pagamento do boleto.
func (b *Boleto) Pay(paidAmountCents int64, paymentID string) error {
	if b.status == BoletoStatusPaid {
		return domain.ErrBoletoAlreadyPaid
	}
	if b.status != BoletoStatusIssued {
		return domain.ErrBoletoCannotBeCancelled
	}
	if paidAmountCents <= 0 {
		return domain.ErrInvalidBoletoNumber
	}

	now := time.Now().UTC()
	b.status = BoletoStatusPaid
	b.paidAmountCents = paidAmountCents
	b.paidAt = &now
	b.version++
	b.updatedAt = now

	b.events = append(b.events, event.BoletoPaid{
		BaseBoletoEvent: event.NewBaseBoletoEvent("boleto.paid", b.id),
		PaidAmountCents: paidAmountCents,
		PaidAt:          now.Format(time.RFC3339),
		PaymentID:       paymentID,
	})

	return nil
}

// Cancel cancela o boleto.
func (b *Boleto) Cancel(reason string) error {
	if b.status == BoletoStatusPaid {
		return domain.ErrBoletoAlreadyPaid
	}
	if b.status == BoletoStatusCancelled {
		return domain.ErrBoletoCannotBeCancelled
	}

	now := time.Now().UTC()
	b.status = BoletoStatusCancelled
	b.cancelledAt = &now
	b.failureReason = reason
	b.version++
	b.updatedAt = now

	b.events = append(b.events, event.BoletoCancelled{
		BaseBoletoEvent: event.NewBaseBoletoEvent("boleto.cancelled", b.id),
		Reason:          reason,
	})

	return nil
}

// MarkExpired marca o boleto como expirado.
func (b *Boleto) MarkExpired() error {
	if b.status != BoletoStatusPending && b.status != BoletoStatusIssued {
		return domain.ErrBoletoCannotBeCancelled
	}

	now := time.Now().UTC()
	b.status = BoletoStatusExpired
	b.version++
	b.updatedAt = now

	b.events = append(b.events, event.BoletoExpired{
		BaseBoletoEvent: event.NewBaseBoletoEvent("boleto.expired", b.id),
		ExpiredAt:       now.Format(time.RFC3339),
	})

	return nil
}

// IsExpired verifica se o boleto está vencido.
func (b *Boleto) IsExpired() bool {
	return time.Now().UTC().After(b.dueDate)
}

// CanBeCancelled verifica se o boleto pode ser cancelado.
func (b *Boleto) CanBeCancelled() bool {
	return b.status == BoletoStatusPending || b.status == BoletoStatusIssued
}

// generateID gera um ID único para o boleto.
func generateID() string {
	return "bol-" + time.Now().Format("20060102150405") + "-" + randomSuffix(8)
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
