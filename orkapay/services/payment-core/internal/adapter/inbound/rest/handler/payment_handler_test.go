package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
	"github.com/stretchr/testify/assert"
	
)

func TestPaymentHandler_CreatePayment(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func() *mockPaymentCreator
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"payment_type":"PIX","method":"PIX_QRCODE","amount_cents":10000,"payer_name":"John","payer_tax_id":"12345678901","payer_email":"john@example.com","payer_pix_key":"john@example.com","payee_name":"Jane","payee_tax_id":"98765432101","payee_pix_key":"jane@example.com","description":"Test"}`,
			setup: func() *mockPaymentCreator {
				return &mockPaymentCreator{
					output: &port.PaymentOutput{
					ID:     "pay-123",
					Status: "PENDING",
					Amount: "R$ 100,00",
				},
			}
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"pay-123","status":"PENDING","amount":"R$ 100,00"`,
		},
		{
			name:       "invalid json",
			body:       `{invalid`,
			setup:      func() *mockPaymentCreator { return &mockPaymentCreator{} },
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request"`,
		},
		{
			name: "use case error",
			body: `{"payment_type":"PIX","method":"PIX_QRCODE","amount_cents":10000,"payer_name":"John","payer_tax_id":"12345678901","payer_email":"john@example.com","payer_pix_key":"john@example.com","payee_name":"Jane","payee_tax_id":"98765432101","payee_pix_key":"jane@example.com","description":"Test"}`,
			setup: func() *mockPaymentCreator {
				return &mockPaymentCreator{err: errors.New("creation failed")}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"error":{"code":"creation_failed"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := tt.setup()
			h := NewPaymentHandler(creator, nil, nil, nil, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.CreatePayment(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantBody)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		})
	}
}

func TestPaymentHandler_GetPayment(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setup      func() *mockPaymentGetter
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			id:   "pay-123",
			setup: func() *mockPaymentGetter {
				amount, _ := domain.NewMoney(10000)
				payerPixKey := domain.MustNewPixKey("john@example.com")
				payeePixKey := domain.MustNewPixKey("jane@example.com")
				payment, _ := domain.NewPayment(
					domain.PaymentTypePix,
					domain.PaymentMethodPixQRCode,
					amount,
					domain.Payer{Name: "John", TaxID: "12345678901", Email: "john@example.com", PixKey: payerPixKey},
					domain.Payee{Name: "Jane", TaxID: "98765432101", PixKey: payeePixKey},
					"Test",
					nil,
				)
				return &mockPaymentGetter{payment: payment}
			},
			wantStatus: http.StatusOK,
                wantBody:   `"Status":"PENDING"`,
		},
		{
			name:       "missing id",
			id:         "",
			setup:      func() *mockPaymentGetter { return &mockPaymentGetter{} },
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"missing_id"`,
		},
		{
			name: "not found",
			id:   "pay-999",
			setup: func() *mockPaymentGetter {
				return &mockPaymentGetter{err: domain.ErrPaymentNotFound}
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":{"code":"not_found"`,
		},
		{
			name: "internal error",
			id:   "pay-123",
			setup: func() *mockPaymentGetter {
				return &mockPaymentGetter{err: errors.New("db error")}
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":{"code":"internal_error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getter := tt.setup()
			h := NewPaymentHandler(nil, nil, nil, getter, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rec := httptest.NewRecorder()

			h.GetPayment(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantBody)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		})
	}
}

func TestPaymentHandler_CancelPayment(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		setup      func() *mockPaymentCanceller
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			id:   "pay-123",
			body: `{"reason":"requested by user"}`,
			setup: func() *mockPaymentCanceller {
				return &mockPaymentCanceller{
					output: &port.PaymentOutput{
						ID:     "pay-123",
						Status: "CANCELLED",
					},
				}
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"id":"pay-123","status":"CANCELLED"`,
		},
		{
			name:       "missing id",
			id:         "",
			body:       `{"reason":"test"}`,
			setup:      func() *mockPaymentCanceller { return &mockPaymentCanceller{} },
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"missing_id"`,
		},
		{
			name:       "invalid json",
			id:         "pay-123",
			body:       `{invalid`,
			setup:      func() *mockPaymentCanceller { return &mockPaymentCanceller{} },
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request"`,
		},
		{
			name: "use case error",
			id:   "pay-123",
			body: `{"reason":"test"}`,
			setup: func() *mockPaymentCanceller {
				return &mockPaymentCanceller{err: errors.New("cancel failed")}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"error":{"code":"cancel_failed"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canceller := tt.setup()
			h := NewPaymentHandler(nil, nil, canceller, nil, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/"+tt.id+"/cancel", bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.id)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.CancelPayment(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantBody)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		})
	}
}

func TestPaymentHandler_WebhookAsaas(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func() *mockPaymentWebhookHandler
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"id":"evt-123","type":"PAYMENT_RECEIVED","external_id":"ext-123","status":"RECEIVED","amount":10000,"timestamp":"2026-07-01T00:00:00Z","signature":"sig"}`,
			setup: func() *mockPaymentWebhookHandler {
				return &mockPaymentWebhookHandler{}
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"processed"}`,
		},
		{
			name:       "invalid json",
			body:       `{invalid`,
			setup:      func() *mockPaymentWebhookHandler { return &mockPaymentWebhookHandler{} },
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_webhook"`,
		},
		{
			name: "use case error",
			body: `{"id":"evt-123","type":"PAYMENT_RECEIVED","external_id":"ext-123","status":"RECEIVED","amount":10000,"timestamp":"2026-07-01T00:00:00Z","signature":"sig"}`,
			setup: func() *mockPaymentWebhookHandler {
				return &mockPaymentWebhookHandler{err: errors.New("webhook failed")}
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"error":{"code":"webhook_failed"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhook := tt.setup()
			h := NewPaymentHandler(nil, nil, nil, nil, webhook)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/asaas", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.WebhookAsaas(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantBody)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		})
	}
}

func TestHealthHandler_Health(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"healthy"`)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestHealthHandler_Ready(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ready"`)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
