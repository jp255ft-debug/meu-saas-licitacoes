package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashout-engine/internal/service/transfer"
)

// TransferHandler expõe os casos de uso de transferência via HTTP REST.
type TransferHandler struct {
	svc *transfer.ServiceImpl
}

// NewTransferHandler cria uma nova instância do handler.
func NewTransferHandler(svc *transfer.ServiceImpl) *TransferHandler {
	return &TransferHandler{svc: svc}
}

// --- Request DTOs ---

// CreateTransferRequest representa o body da requisição POST /transfers.
type CreateTransferRequest struct {
	AmountCents   int64             `json:"amount_cents"`
	PayeeName     string            `json:"payee_name"`
	PayeeTaxID    string            `json:"payee_tax_id"`
	PayeePixKey   string            `json:"payee_pix_key"`
	PayeePixType  string            `json:"payee_pix_type"`
	PayeeBankCode string            `json:"payee_bank_code"`
	PayeeBranch   string            `json:"payee_branch"`
	PayeeAccount  string            `json:"payee_account"`
	PayeeAcctType string            `json:"payee_account_type"`
	Provider      string            `json:"provider"`
	Description   string            `json:"description"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// ActionRequest representa o body para operações que exigem motivo.
type ActionRequest struct {
	Reason string `json:"reason"`
}

// --- Handlers ---

// CreateTransfer godoc
// POST /api/v1/transfers
func (h *TransferHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	input := inbound.CreateTransferInput{
		AmountCents:   req.AmountCents,
		PayeeName:     req.PayeeName,
		PayeeTaxID:    req.PayeeTaxID,
		PayeePixKey:   req.PayeePixKey,
		PayeePixType:  req.PayeePixType,
		PayeeBankCode: req.PayeeBankCode,
		PayeeBranch:   req.PayeeBranch,
		PayeeAccount:  req.PayeeAccount,
		PayeeAcctType: req.PayeeAcctType,
		Provider:      req.Provider,
		Description:   req.Description,
		Metadata:      req.Metadata,
	}

	output, err := h.svc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) || errors.Is(err, domain.ErrInvalidPayee) || errors.Is(err, domain.ErrInvalidPixKey) {
			respondError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "creation_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, output)
}

// GetTransfer godoc
// GET /api/v1/transfers/{id}
func (h *TransferHandler) GetTransfer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da transferência é obrigatório")
		return
	}

	transferEntity, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTransferNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "transferência não encontrada")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, transferEntity)
}

// ListTransfers godoc
// GET /api/v1/transfers?status=PENDING&limit=10&offset=0
func (h *TransferHandler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		respondError(w, http.StatusBadRequest, "missing_status", "parâmetro status é obrigatório")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	transfers, err := h.svc.ListByStatus(r.Context(), status, limit, offset)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "list_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, transfers)
}

// ApproveTransfer godoc
// POST /api/v1/transfers/{id}/approve
func (h *TransferHandler) ApproveTransfer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da transferência é obrigatório")
		return
	}

	var req struct {
		ExternalID string `json:"external_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	output, err := h.svc.Approve(r.Context(), id, req.ExternalID)
	if err != nil {
		if errors.Is(err, domain.ErrTransferNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "transferência não encontrada")
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "approve_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, output)
}

// SettleTransfer godoc
// POST /api/v1/transfers/{id}/settle
func (h *TransferHandler) SettleTransfer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da transferência é obrigatório")
		return
	}

	output, err := h.svc.Settle(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTransferNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "transferência não encontrada")
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "settle_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, output)
}

// FailTransfer godoc
// POST /api/v1/transfers/{id}/fail
func (h *TransferHandler) FailTransfer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da transferência é obrigatório")
		return
	}

	var req ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	output, err := h.svc.Fail(r.Context(), id, req.Reason)
	if err != nil {
		if errors.Is(err, domain.ErrTransferNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "transferência não encontrada")
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "fail_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, output)
}

// CancelTransfer godoc
// POST /api/v1/transfers/{id}/cancel
func (h *TransferHandler) CancelTransfer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da transferência é obrigatório")
		return
	}

	var req ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	output, err := h.svc.Cancel(r.Context(), id, req.Reason)
	if err != nil {
		if errors.Is(err, domain.ErrTransferNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "transferência não encontrada")
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "cancel_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, output)
}

// --- Response Helpers ---

// respondJSON envia uma resposta JSON.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// respondError envia uma resposta de erro padronizada.
func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}