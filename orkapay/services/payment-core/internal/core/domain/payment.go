package domain

import (
	"fmt"
	"time"
)

// PaymentStatus representa o status de um pagamento.
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"
	PaymentStatusProcessing PaymentStatus = "PROCESSING"
	PaymentStatusApproved   PaymentStatus = "APPROVED"
	PaymentStatusSettled    PaymentStatus = "SETTLED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusCancelled  PaymentStatus = "CANCELLED"
)

// PaymentType representa o tipo de pagamento.
type PaymentType string

const (
	PaymentTypePix      PaymentType = "PIX"
	PaymentTypePixAuto  PaymentType = "PIX_AUTO"
	PaymentTypeBoleto   PaymentType = "BOLETO"
	PaymentTypeTransfer PaymentType = "TRANSFER"
)

// PaymentMethod representa o método específico de pagamento.
type PaymentMethod string

const (
	PaymentMethodPixQRCode  PaymentMethod = "PIX_QRCODE"
	PaymentMethodPixCopyPaste PaymentMethod = "PIX_COPY_PASTE"
	PaymentMethodBoletoPrint PaymentMethod = "BOLETO_PRINT"
	PaymentMethodBoletoHybrid PaymentMethod = "BOLETO_HYBRID"
)

// Payer representa o pagador.
type Payer struct {
	Name      string
	TaxID     string // CPF ou CNPJ
	Email     string
	PixKey    PixKey
}

// Payee representa o beneficiário.
type Payee struct {
	Name   string
	TaxID  string
	PixKey PixKey
	BankAccount *BankAccount
}

// BankAccount representa dados bancários.
type BankAccount struct {
	BankCode    string
	Branch      string
	AccountNumber string
	AccountType string // CHECKING, SAVINGS
}

// Payment é o Aggregate Root do domínio de pagamentos.
// Contém toda a lógica de negócio e invariantes.
type Payment struct {
	id            string
	externalID    string // ID do gateway externo
	status        PaymentStatus
	paymentType   PaymentType
	method        PaymentMethod
	amount        Money
	fee           Money // Taxa do orquestrador
	netAmount     Money // amount - fee
	payer         Payer
	payee         Payee
	description   string
	expiresAt     *time.Time
	processedAt   *time.Time
	settledAt     *time.Time
	failedAt      *time.Time
	cancelledAt   *time.Time
	failureReason string
	metadata      map[string]string
	createdAt     time.Time
	updatedAt     time.Time
	version       int // Otimistic locking
	events        []DomainEvent
}

// DomainEvent interface base para eventos.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// Payment getters (read-only acesso externo)

func (p *Payment) ID() string              { return p.id }
func (p *Payment) ExternalID() string      { return p.externalID }
func (p *Payment) Status() PaymentStatus    { return p.status }
func (p *Payment) PaymentType() PaymentType { return p.paymentType }
func (p *Payment) Method() PaymentMethod    { return p.method }
func (p *Payment) Amount() Money            { return p.amount }
func (p *Payment) Fee() Money               { return p.fee }
func (p *Payment) NetAmount() Money          { return p.netAmount }
func (p *Payment) Payer() Payer             { return p.payer }
func (p *Payment) Payee() Payee             { return p.payee }
func (p *Payment) Description() string      { return p.description }
func (p *Payment) ExpiresAt() *time.Time    { return p.expiresAt }
func (p *Payment) ProcessedAt() *time.Time { return p.processedAt }
func (p *Payment) SettledAt() *time.Time    { return p.settledAt }
func (p *Payment) FailedAt() *time.Time     { return p.failedAt }
func (p *Payment) CancelledAt() *time.Time  { return p.cancelledAt }
func (p *Payment) FailureReason() string     { return p.failureReason }
func (p *Payment) Metadata() map[string]string {
	// Retorna cópia para imutabilidade
	m := make(map[string]string, len(p.metadata))
	for k, v := range p.metadata {
		m[k] = v
	}
	return m
}
func (p *Payment) CreatedAt() time.Time { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time { return p.updatedAt }
func (p *Payment) Version() int         { return p.version }

// NewPayment cria um novo pagamento (factory).
// Aplica invariantes e regras de negócio na criação.
func NewPayment(
	paymentType PaymentType,
	method PaymentMethod,
	amount Money,
	payer Payer,
	payee Payee,
	description string,
	metadata map[string]string,
) (*Payment, error) {
	// Invariantes de criação
	if !amount.IsPositive() {
		return nil, ErrNegativeAmount
	}
	if payer.TaxID == "" {
		return nil, ErrInvalidPayer
	}
	if payee.TaxID == "" {
		return nil, ErrInvalidPayee
	}
	if description == "" {
		return nil, ErrInvalidPaymentStatus // reutilizando — ou criar erro específico
	}

	id := generatePaymentID()
	now := time.Now().UTC()

	// Calcula taxa do orquestrador (ex: 0,25% para Pix)
	fee := calculateFee(amount, paymentType)
	netAmount := amount.Sub(fee)

	payment := &Payment{
		id:          id,
		status:      PaymentStatusPending,
		paymentType: paymentType,
		method:      method,
		amount:      amount,
		fee:         fee,
		netAmount:   netAmount,
		payer:       payer,
		payee:       payee,
		description: description,
		metadata:    copyMetadata(metadata),
		createdAt:   now,
		updatedAt:   now,
		version:     1,
		events:      make([]DomainEvent, 0),
	}

	// Define expiração baseada no tipo
	payment.setExpiration()

	// Emite evento de criação
	payment.emit(PaymentCreatedEvent{
		PaymentID:   payment.id,
		Amount:      payment.amount.Cents(),
		PaymentType: string(payment.paymentType),
		Method:      string(payment.method),
		PayerTaxID:  payment.payer.TaxID,
		PayeeTaxID:  payment.payee.TaxID,
		Timestamp:   now,
	})

	return payment, nil
}

// Approve aprova o pagamento (mudança de status).
func (p *Payment) Approve(externalID string) error {
	if p.status == PaymentStatusApproved || p.status == PaymentStatusSettled {
		return ErrPaymentAlreadyProcessed
	}
	if p.status != PaymentStatusPending && p.status != PaymentStatusProcessing {
		return ErrInvalidPaymentStatus
	}

	now := time.Now().UTC()
	p.status = PaymentStatusApproved
	p.externalID = externalID
	p.processedAt = &now
	p.updatedAt = now
	p.version++

	p.emit(PaymentApprovedEvent{
		PaymentID:  p.id,
		ExternalID: externalID,
		Amount:     p.amount.Cents(),
		Timestamp:  now,
	})

	return nil
}

// Settle liquida o pagamento (dinheiro na conta).
func (p *Payment) Settle() error {
	if p.status != PaymentStatusApproved {
		return ErrInvalidPaymentStatus
	}
	if p.status == PaymentStatusSettled {
		return ErrPaymentAlreadySettled
	}

	now := time.Now().UTC()
	p.status = PaymentStatusSettled
	p.settledAt = &now
	p.updatedAt = now
	p.version++

	p.emit(PaymentSettledEvent{
		PaymentID:  p.id,
		Amount:     p.amount.Cents(),
		NetAmount:  p.netAmount.Cents(),
		Fee:        p.fee.Cents(),
		Timestamp:  now,
	})

	return nil
}

// Fail marca o pagamento como falho.
func (p *Payment) Fail(reason string) error {
	if p.status == PaymentStatusSettled {
		return ErrPaymentAlreadySettled
	}
	if p.status == PaymentStatusFailed {
		return ErrPaymentAlreadyFailed
	}
	if p.status == PaymentStatusCancelled {
		return ErrPaymentAlreadyCancelled
	}

	now := time.Now().UTC()
	p.status = PaymentStatusFailed
	p.failureReason = reason
	p.failedAt = &now
	p.updatedAt = now
	p.version++

	p.emit(PaymentFailedEvent{
		PaymentID:  p.id,
		Reason:     reason,
		Amount:     p.amount.Cents(),
		Timestamp:  now,
	})

	return nil
}

// Cancel cancela um pagamento pendente.
func (p *Payment) Cancel(reason string) error {
	if p.status != PaymentStatusPending {
		return ErrPaymentCannotBeCancelled
	}

	now := time.Now().UTC()
	p.status = PaymentStatusCancelled
	p.cancelledAt = &now
	p.updatedAt = now
	p.version++

	p.emit(PaymentCancelledEvent{
		PaymentID:  p.id,
		Reason:     reason,
		Timestamp:  now,
	})

	return nil
}

// IsExpired verifica se o pagamento expirou.
func (p *Payment) IsExpired() bool {
	if p.expiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*p.expiresAt)
}

// CanRetry verifica se o pagamento pode ser retentado.
func (p *Payment) CanRetry() bool {
	return p.status == PaymentStatusFailed || (p.status == PaymentStatusPending && p.IsExpired())
}

// Events retorna os eventos não consumidos e os limpa.
func (p *Payment) Events() []DomainEvent {
	evts := make([]DomainEvent, len(p.events))
	copy(evts, p.events)
	p.events = p.events[:0] // limpa
	return evts
}

// emit adiciona um evento ao aggregate.
func (p *Payment) emit(event DomainEvent) {
	p.events = append(p.events, event)
}

// setExpiration define o tempo de expiração baseado no tipo.
func (p *Payment) setExpiration() {
	now := time.Now().UTC()
	var expiry time.Time

	switch p.paymentType {
	case PaymentTypePix:
		expiry = now.Add(24 * time.Hour) // Pix expira em 24h
	case PaymentTypePixAuto:
		expiry = now.Add(48 * time.Hour) // Pix Automático em 48h
	case PaymentTypeBoleto:
		expiry = now.Add(7 * 24 * time.Hour) // Boleto em 7 dias
	default:
		expiry = now.Add(24 * time.Hour)
	}

	p.expiresAt = &expiry
}

// calculateFee calcula a taxa do orquestrador baseada no tipo.
// A taxa é expressa em basis points (bps):
// - Pix: 25 bps = 0,25%
// - Boleto: 50 bps = 0,50%
func calculateFee(amount Money, paymentType PaymentType) Money {
	var bps int64
	switch paymentType {
	case PaymentTypePix:
		bps = 25 // 0,25%
	case PaymentTypePixAuto:
		bps = 25 // 0,25%
	case PaymentTypeBoleto:
		bps = 50 // 0,50%
	default:
		bps = 25 // 0,25%
	}
	// (cents * bps) / 10000 = valor da taxa em centavos
	return Money{cents: (amount.cents * bps) / 10000}
}

// copyMetadata cria uma cópia do mapa de metadata.
func copyMetadata(m map[string]string) map[string]string {
	if m == nil {
		return make(map[string]string)
	}
	c := make(map[string]string, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

// generatePaymentID gera um ID único para o pagamento.
// Implementação simples sem dependências externas.
// Em produção, pode ser substituído por UUID via adaptador.
func generatePaymentID() string {
	return fmt.Sprintf("PAY-%d", time.Now().UTC().UnixNano())
}

// PaymentCreatedEvent — emitido quando um pagamento é criado.
type PaymentCreatedEvent struct {
	PaymentID   string
	Amount      int64
	PaymentType string
	Method      string
	PayerTaxID  string
	PayeeTaxID  string
	Timestamp   time.Time
}

func (e PaymentCreatedEvent) EventName() string    { return "payment.created" }
func (e PaymentCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

// PaymentApprovedEvent — emitido quando o gateway aprova.
type PaymentApprovedEvent struct {
	PaymentID  string
	ExternalID string
	Amount     int64
	Timestamp  time.Time
}

func (e PaymentApprovedEvent) EventName() string    { return "payment.approved" }
func (e PaymentApprovedEvent) OccurredAt() time.Time { return e.Timestamp }

// PaymentSettledEvent — emitido quando o dinheiro cai na conta.
type PaymentSettledEvent struct {
	PaymentID  string
	Amount     int64
	NetAmount  int64
	Fee        int64
	Timestamp  time.Time
}

func (e PaymentSettledEvent) EventName() string    { return "payment.settled" }
func (e PaymentSettledEvent) OccurredAt() time.Time { return e.Timestamp }

// PaymentFailedEvent — emitido quando o pagamento falha.
type PaymentFailedEvent struct {
	PaymentID  string
	Reason     string
	Amount     int64
	Timestamp  time.Time
}

func (e PaymentFailedEvent) EventName() string    { return "payment.failed" }
func (e PaymentFailedEvent) OccurredAt() time.Time { return e.Timestamp }

// PaymentCancelledEvent — emitido quando o pagamento é cancelado.
type PaymentCancelledEvent struct {
	PaymentID  string
	Reason     string
	Timestamp  time.Time
}

func (e PaymentCancelledEvent) EventName() string    { return "payment.cancelled" }
func (e PaymentCancelledEvent) OccurredAt() time.Time { return e.Timestamp }
