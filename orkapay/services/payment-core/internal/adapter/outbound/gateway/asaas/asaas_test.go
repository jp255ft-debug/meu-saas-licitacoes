package asaas_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/adapter/outbound/gateway/asaas"
	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Testes Unitarios: Webhook ---

func TestValidateSignature_EmptySignature(t *testing.T) {
	err := asaas.ValidateSignature([]byte(`{"test":"data"}`), "", "key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ausente")
}

func TestValidateSignature_EmptyAPIKey(t *testing.T) {
	err := asaas.ValidateSignature([]byte(`{"test":"data"}`), "sig", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nao configurada")
}

func TestParseWebhook_Valid(t *testing.T) {
	body := `{"event":"PAYMENT_RECEIVED","payment":{"id":"pay-123","status":"RECEIVED","value":1000}}`
	event, err := asaas.ParseWebhook(strings.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, "PAYMENT_RECEIVED", event.Event)
	assert.Equal(t, "pay-123", event.Payment.ID)
	assert.Equal(t, "RECEIVED", event.Payment.Status)
	assert.Equal(t, int64(1000), event.Payment.Value)
}

func TestParseWebhook_EmptyEvent(t *testing.T) {
	body := `{"event":"","payment":{"id":"pay-123","status":"RECEIVED","value":1000}}`
	_, err := asaas.ParseWebhook(strings.NewReader(body))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "evento vazio")
}

func TestParseWebhook_EmptyPaymentID(t *testing.T) {
	body := `{"event":"PAYMENT_RECEIVED","payment":{"id":"","status":"RECEIVED","value":1000}}`
	_, err := asaas.ParseWebhook(strings.NewReader(body))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment ID vazio")
}

func TestParseWebhook_InvalidJSON(t *testing.T) {
	_, err := asaas.ParseWebhook(strings.NewReader(`{invalid`))
	assert.Error(t, err)
}

// --- Testes Unitarios: PaymentGateway ---

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, port.PaymentGateway) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := asaas.NewClient(server.URL, "test-api-key", 5*time.Second)
	gateway := asaas.NewPaymentGateway(client)
	return server, gateway
}

func TestPaymentGateway_CreatePayment_Success(t *testing.T) {
	server, gw := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v3/payments", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("access_token"))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"pay-abc-123","status":"PENDING","value":10000}`))
	})
	defer server.Close()

	payment := createTestPayment(t)
	externalID, status, err := gw.CreatePayment(context.Background(), payment)
	require.NoError(t, err)
	assert.Equal(t, "pay-abc-123", externalID)
	assert.Equal(t, "PENDING", status)
}

func TestPaymentGateway_CreatePayment_Unauthorized(t *testing.T) {
	server, gw := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	defer server.Close()

	payment := createTestPayment(t)
	_, _, err := gw.CreatePayment(context.Background(), payment)
	assert.ErrorIs(t, err, domain.ErrMissingGatewayConfig)
}

func TestPaymentGateway_GetPaymentStatus_Success(t *testing.T) {
	server, gw := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v3/payments/pay-abc-123")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"pay-abc-123","status":"RECEIVED","value":10000}`))
	})
	defer server.Close()

	status, err := gw.GetPaymentStatus(context.Background(), "pay-abc-123")
	require.NoError(t, err)
	assert.Equal(t, "APPROVED", status)
}

func TestPaymentGateway_GetPaymentStatus_NotFound(t *testing.T) {
	server, gw := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	_, err := gw.GetPaymentStatus(context.Background(), "non-existent")
	assert.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestPaymentGateway_CancelPayment_Success(t *testing.T) {
	server, gw := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := gw.CancelPayment(context.Background(), "pay-abc-123")
	assert.NoError(t, err)
}

func TestPaymentGateway_HealthCheck_Success(t *testing.T) {
	server, gw := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := gw.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestPaymentGateway_RefundPayment_Success(t *testing.T) {
	server, gw := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/refund")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := gw.RefundPayment(context.Background(), "pay-abc-123")
	assert.NoError(t, err)
}

func TestPaymentGateway_GatewayRejection(t *testing.T) {
	server, gw := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"errors":[{"code":"invalid_value","description":"Valor invalido"}]}`))
	})
	defer server.Close()

	payment := createTestPayment(t)
	_, _, err := gw.CreatePayment(context.Background(), payment)
	assert.ErrorIs(t, err, domain.ErrGatewayRejection)
	assert.Contains(t, err.Error(), "Valor invalido")
}

// --- Helpers ---

func createTestPayment(t *testing.T) *domain.Payment {
	t.Helper()
	amount, err := domain.NewMoney(10000)
	require.NoError(t, err)
	payer := domain.Payer{
		Name:   "Joao Silva",
		TaxID:  "52998224725",
		Email:  "joao@email.com",
		PixKey: domain.MustNewPixKey("joao@email.com"),
	}
	payee := domain.Payee{
		Name:   "OrkaPay LTDA",
		TaxID:  "11222333000181",
		PixKey: domain.MustNewPixKey("11222333000181"),
	}
	payment, err := domain.NewPayment(
		domain.PaymentTypePix,
		domain.PaymentMethodPixQRCode,
		amount,
		payer,
		payee,
		"Teste integracao Asaas",
		map[string]string{"test": "true"},
	)
	require.NoError(t, err)
	return payment
}