package domain

import (
	"testing"
	"time"
)

func TestNewPayment_Valid(t *testing.T) {
	amount := MustNewMoney(10000) // R$ 100,00
	payer := Payer{
		Name:   "João Silva",
		TaxID:  "52998224725",
		Email:  "joao@email.com",
		PixKey: MustNewPixKey("joao@email.com"),
	}
	payee := Payee{
		Name:   "OrkaPay LTDA",
		TaxID:  "11222333000181",
		PixKey: MustNewPixKey("11222333000181"),
	}

	payment, err := NewPayment(
		PaymentTypePix,
		PaymentMethodPixQRCode,
		amount,
		payer,
		payee,
		"Pagamento teste",
		map[string]string{"order_id": "123"},
	)
	if err != nil {
		t.Fatalf("esperado sem erro, got: %v", err)
	}
	if payment.ID() == "" {
		t.Error("esperado ID gerado")
	}
	if payment.Status() != PaymentStatusPending {
		t.Errorf("esperado status PENDING, got: %s", payment.Status())
	}
	if !payment.Amount().Equals(amount) {
		t.Error("esperado amount igual ao informado")
	}
	if payment.Fee().IsZero() {
		t.Error("esperado fee calculado")
	}
	if payment.ExpiresAt() == nil {
		t.Error("esperado expiresAt definido")
	}

	// Eventos
	evts := payment.Events()
	if len(evts) != 1 {
		t.Errorf("esperado 1 evento, got: %d", len(evts))
	}
	if evts[0].EventName() != "payment.created" {
		t.Errorf("esperado evento payment.created, got: %s", evts[0].EventName())
	}
}

func TestNewPayment_InvalidAmount(t *testing.T) {
	_, err := NewPayment(
		PaymentTypePix,
		PaymentMethodPixQRCode,
		Zero(),
		Payer{Name: "João", TaxID: "123"},
		Payee{Name: "OrkaPay", TaxID: "456"},
		"Teste",
		nil,
	)
	if err != ErrNegativeAmount {
		t.Fatalf("esperado ErrNegativeAmount, got: %v", err)
	}
}

func TestNewPayment_InvalidPayer(t *testing.T) {
	_, err := NewPayment(
		PaymentTypePix,
		PaymentMethodPixQRCode,
		MustNewMoney(1000),
		Payer{Name: "João", TaxID: ""},
		Payee{Name: "OrkaPay", TaxID: "456"},
		"Teste",
		nil,
	)
	if err != ErrInvalidPayer {
		t.Fatalf("esperado ErrInvalidPayer, got: %v", err)
	}
}

func TestNewPayment_InvalidPayee(t *testing.T) {
	_, err := NewPayment(
		PaymentTypePix,
		PaymentMethodPixQRCode,
		MustNewMoney(1000),
		Payer{Name: "João", TaxID: "123"},
		Payee{Name: "OrkaPay", TaxID: ""},
		"Teste",
		nil,
	)
	if err != ErrInvalidPayee {
		t.Fatalf("esperado ErrInvalidPayee, got: %v", err)
	}
}

func TestPayment_Approve(t *testing.T) {
	payment := createTestPayment(t)

	err := payment.Approve("EXT-12345")
	if err != nil {
		t.Fatalf("esperado sem erro, got: %v", err)
	}
	if payment.Status() != PaymentStatusApproved {
		t.Errorf("esperado APPROVED, got: %s", payment.Status())
	}
	if payment.ExternalID() != "EXT-12345" {
		t.Errorf("esperado externalID EXT-12345, got: %s", payment.ExternalID())
	}
	if payment.ProcessedAt() == nil {
		t.Error("esperado processedAt definido")
	}

	// Evento
	evts := payment.Events()
	found := false
	for _, e := range evts {
		if e.EventName() == "payment.approved" {
			found = true
			break
		}
	}
	if !found {
		t.Error("esperado evento payment.approved")
	}
}

func TestPayment_Approve_AlreadyApproved(t *testing.T) {
	payment := createTestPayment(t)
	payment.Approve("EXT-1")

	err := payment.Approve("EXT-2")
	if err != ErrPaymentAlreadyProcessed {
		t.Fatalf("esperado ErrPaymentAlreadyProcessed, got: %v", err)
	}
}

func TestPayment_Settle(t *testing.T) {
	payment := createTestPayment(t)
	payment.Approve("EXT-123")

	err := payment.Settle()
	if err != nil {
		t.Fatalf("esperado sem erro, got: %v", err)
	}
	if payment.Status() != PaymentStatusSettled {
		t.Errorf("esperado SETTLED, got: %s", payment.Status())
	}
	if payment.SettledAt() == nil {
		t.Error("esperado settledAt definido")
	}
}

func TestPayment_Settle_NotApproved(t *testing.T) {
	payment := createTestPayment(t)

	err := payment.Settle()
	if err != ErrInvalidPaymentStatus {
		t.Fatalf("esperado ErrInvalidPaymentStatus, got: %v", err)
	}
}

func TestPayment_Fail(t *testing.T) {
	payment := createTestPayment(t)

	err := payment.Fail("Saldo insuficiente")
	if err != nil {
		t.Fatalf("esperado sem erro, got: %v", err)
	}
	if payment.Status() != PaymentStatusFailed {
		t.Errorf("esperado FAILED, got: %s", payment.Status())
	}
	if payment.FailureReason() != "Saldo insuficiente" {
		t.Errorf("esperado failure reason, got: %s", payment.FailureReason())
	}
}

func TestPayment_Fail_AlreadySettled(t *testing.T) {
	payment := createTestPayment(t)
	payment.Approve("EXT-1")
	payment.Settle()

	err := payment.Fail("erro")
	if err != ErrPaymentAlreadySettled {
		t.Fatalf("esperado ErrPaymentAlreadySettled, got: %v", err)
	}
}

func TestPayment_Cancel(t *testing.T) {
	payment := createTestPayment(t)

	err := payment.Cancel("Cancelado pelo usuário")
	if err != nil {
		t.Fatalf("esperado sem erro, got: %v", err)
	}
	if payment.Status() != PaymentStatusCancelled {
		t.Errorf("esperado CANCELLED, got: %s", payment.Status())
	}
}

func TestPayment_Cancel_AlreadyApproved(t *testing.T) {
	payment := createTestPayment(t)
	payment.Approve("EXT-1")

	err := payment.Cancel("cancelar")
	if err != ErrPaymentCannotBeCancelled {
		t.Fatalf("esperado ErrPaymentCannotBeCancelled, got: %v", err)
	}
}

func TestPayment_IsExpired(t *testing.T) {
	payment := createTestPayment(t)
	if payment.IsExpired() {
		t.Error("pagamento recém-criado não deve estar expirado")
	}

	// Simula expiração forçada
	past := time.Now().UTC().Add(-1 * time.Hour)
	payment.expiresAt = &past
	if !payment.IsExpired() {
		t.Error("esperado IsExpired = true após forçar expiração")
	}
}

func TestPayment_CanRetry(t *testing.T) {
	payment := createTestPayment(t)
	if payment.CanRetry() {
		t.Error("pagamento pendente não expirado não deve ser retryable")
	}

	payment.Fail("erro")
	if !payment.CanRetry() {
		t.Error("pagamento falho deve ser retryable")
	}
}

func TestPayment_VersionIncrement(t *testing.T) {
	payment := createTestPayment(t)
	v1 := payment.Version()

	payment.Approve("EXT-1")
	if payment.Version() != v1+1 {
		t.Errorf("esperado version %d, got: %d", v1+1, payment.Version())
	}
}

func TestPayment_MetadataImmutability(t *testing.T) {
	payment := createTestPayment(t)
	m := payment.Metadata()
	m["hacked"] = "true"

	m2 := payment.Metadata()
	if _, ok := m2["hacked"]; ok {
		t.Error("metadata deve retornar cópia, não referência")
	}
}

// Helper
func createTestPayment(t *testing.T) *Payment {
	t.Helper()
	payment, err := NewPayment(
		PaymentTypePix,
		PaymentMethodPixQRCode,
		MustNewMoney(10000),
		Payer{
			Name:   "João Silva",
			TaxID:  "52998224725",
			Email:  "joao@email.com",
			PixKey: MustNewPixKey("joao@email.com"),
		},
		Payee{
			Name:   "OrkaPay LTDA",
			TaxID:  "11222333000181",
			PixKey: MustNewPixKey("11222333000181"),
		},
		"Pagamento de teste",
		map[string]string{"order_id": "123"},
	)
	if err != nil {
		t.Fatalf("falha ao criar pagamento de teste: %v", err)
	}
	return payment
}
