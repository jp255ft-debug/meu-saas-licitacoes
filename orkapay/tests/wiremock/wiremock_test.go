//go:build wiremock

package wiremock_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constantes
// ---------------------------------------------------------------------------

const (
	wiremockURL = "http://localhost:8082"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type invoicePayload struct {
	AmountCents int64  `json:"amount_cents"`
	DueDate     string `json:"due_date"`
	PayerName   string `json:"payer_name"`
	PayerTaxID  string `json:"payer_tax_id"`
}

type transferPayload struct {
	AmountCents int64  `json:"amount_cents"`
	PayeeName   string `json:"payee_name"`
	PayeeTaxID  string `json:"payee_tax_id"`
	Description string `json:"description"`
}

func postInvoice(t *testing.T, payload interface{}, timeout time.Duration) (*http.Response, error) {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", wiremockURL+"/api/v3/invoices", bytes.NewReader(body))
	req.Header.Set("access_token", "test-token")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	return client.Do(req)
}

func postTransfer(t *testing.T, payload interface{}, timeout time.Duration) (*http.Response, error) {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", wiremockURL+"/api/v1/transfers", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	return client.Do(req)
}

// ---------------------------------------------------------------------------
// Tests: Asaas - Sucesso
// ---------------------------------------------------------------------------

func TestAsaas_CreateInvoice_Success(t *testing.T) {
	payload := invoicePayload{
		AmountCents: 10000,
		DueDate:     "2026-07-10",
		PayerName:   "João Silva",
		PayerTaxID:  "12345678901",
	}

	resp, err := postInvoice(t, payload, 3*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "inv_wiremock_12345", result["id"])
	assert.Equal(t, "PENDING", result["status"])
}

// ---------------------------------------------------------------------------
// Tests: Asaas - Erro 500
// ---------------------------------------------------------------------------

func TestAsaas_CreateInvoice_InternalError(t *testing.T) {
	payload := invoicePayload{
		AmountCents: 10000,
		DueDate:     "2026-07-10",
		PayerName:   "João Silva",
		PayerTaxID:  "12345678901",
	}

	resp, err := postInvoice(t, payload, 3*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Internal Server Error", result["error"])
	assert.Equal(t, "INTERNAL_ERROR", result["code"])
}

// ---------------------------------------------------------------------------
// Tests: Asaas - Timeout
// ---------------------------------------------------------------------------

func TestAsaas_CreateInvoice_Timeout(t *testing.T) {
	payload := invoicePayload{
		AmountCents: 10000,
		DueDate:     "2026-07-10",
		PayerName:   "João Silva",
		PayerTaxID:  "12345678901",
	}

	// Timeout de 3s, stub responde em 5s → deve falhar
	_, err := postInvoice(t, payload, 3*time.Second)
	assert.Error(t, err, "esperava erro de timeout")
	assert.Contains(t, err.Error(), "timeout")
}

// ---------------------------------------------------------------------------
// Tests: Asaas - Rate Limit (429)
// ---------------------------------------------------------------------------

func TestAsaas_CreateInvoice_RateLimit(t *testing.T) {
	payload := invoicePayload{
		AmountCents: 10000,
		DueDate:     "2026-07-10",
		PayerName:   "João Silva",
		PayerTaxID:  "12345678901",
	}

	resp, err := postInvoice(t, payload, 3*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, "30", resp.Header.Get("Retry-After"))

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Too Many Requests", result["error"])
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", result["code"])
}

// ---------------------------------------------------------------------------
// Tests: Stark Bank - Sucesso
// ---------------------------------------------------------------------------

func TestStarkBank_CreateTransfer_Success(t *testing.T) {
	payload := transferPayload{
		AmountCents: 50000,
		PayeeName:   "Jane Doe",
		PayeeTaxID:  "98765432101",
		Description: "Test transfer",
	}

	resp, err := postTransfer(t, payload, 3*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "trf_wiremock_12345", result["id"])
	assert.Equal(t, "PROCESSING", result["status"])
}

// ---------------------------------------------------------------------------
// Tests: Stark Bank - Erro 500
// ---------------------------------------------------------------------------

func TestStarkBank_CreateTransfer_InternalError(t *testing.T) {
	payload := transferPayload{
		AmountCents: 50000,
		PayeeName:   "Jane Doe",
		PayeeTaxID:  "98765432101",
		Description: "Test transfer",
	}

	resp, err := postTransfer(t, payload, 3*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Internal Server Error", result["error"])
	assert.Equal(t, "internal_error", result["code"])
}

// ---------------------------------------------------------------------------
// Tests: Stark Bank - Timeout
// ---------------------------------------------------------------------------

func TestStarkBank_CreateTransfer_Timeout(t *testing.T) {
	payload := transferPayload{
		AmountCents: 50000,
		PayeeName:   "Jane Doe",
		PayeeTaxID:  "98765432101",
		Description: "Test transfer",
	}

	// Timeout de 3s, stub responde em 5s → deve falhar
	_, err := postTransfer(t, payload, 3*time.Second)
	assert.Error(t, err, "esperava erro de timeout")
	assert.Contains(t, err.Error(), "timeout")
}