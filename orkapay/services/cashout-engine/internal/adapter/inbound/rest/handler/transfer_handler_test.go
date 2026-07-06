package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Tests: CreateTransfer
// ---------------------------------------------------------------------------

func TestTransferHandler_CreateTransfer_InvalidJSON(t *testing.T) {
	h := NewTransferHandler(nil)
	body := `{invalid`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transfers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateTransfer(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"invalid_request"`)
}

// ---------------------------------------------------------------------------
// Tests: GetTransfer
// ---------------------------------------------------------------------------

func TestTransferHandler_GetTransfer_NotFound(t *testing.T) {
	h := NewTransferHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transfers/cout-123", nil)
	req.SetPathValue("id", "cout-123")
	rec := httptest.NewRecorder()

	h.GetTransfer(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"not_found"`)
}

func TestTransferHandler_GetTransfer_MissingID(t *testing.T) {
	h := NewTransferHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transfers/", nil)
	req.SetPathValue("id", "")
	rec := httptest.NewRecorder()

	h.GetTransfer(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"missing_id"`)
}

// ---------------------------------------------------------------------------
// Tests: ListTransfers
// ---------------------------------------------------------------------------

func TestTransferHandler_ListTransfers_MissingStatus(t *testing.T) {
	h := NewTransferHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transfers", nil)
	rec := httptest.NewRecorder()

	h.ListTransfers(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"missing_status"`)
}

// ---------------------------------------------------------------------------
// Tests: CancelTransfer
// ---------------------------------------------------------------------------

func TestTransferHandler_CancelTransfer_MissingID(t *testing.T) {
	h := NewTransferHandler(nil)

	body := `{"reason":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transfers//cancel", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "")
	rec := httptest.NewRecorder()

	h.CancelTransfer(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"missing_id"`)
}

func TestTransferHandler_CancelTransfer_InvalidJSON(t *testing.T) {
	h := NewTransferHandler(nil)

	body := `{invalid`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transfers/cout-123/cancel", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "cout-123")
	rec := httptest.NewRecorder()

	h.CancelTransfer(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"invalid_request"`)
}

// ---------------------------------------------------------------------------
// Tests: HealthHandler
// ---------------------------------------------------------------------------

func TestTransferHealthHandler_Health(t *testing.T) {
	h := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"healthy"`)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestTransferHealthHandler_Ready(t *testing.T) {
	h := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ready"`)
}

// ---------------------------------------------------------------------------
// Tests: respond helpers
// ---------------------------------------------------------------------------

func TestTransferRespondJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	respondJSON(rec, http.StatusOK, map[string]string{"key": "value"})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"key":"value"`)
}

func TestTransferRespondError(t *testing.T) {
	rec := httptest.NewRecorder()
	respondError(rec, http.StatusBadRequest, "test_code", "test message")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"test_code"`)
}

// avoid unused import
var _ = json.Marshal
var _ = domain.ErrTransferNotFound
var _ = entity.TransferStatusPending