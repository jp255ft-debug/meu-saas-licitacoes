package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashin-engine/internal/service/invoice"
)

// InvoiceHandler expõe os casos de uso de fatura via HTTP REST.
type InvoiceHandler struct {
	svc *invoice.ServiceImpl
}

// NewInvoiceHandler cria uma nova instância do handler.
func NewInvoiceHandler(svc *invoice.ServiceImpl) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

// --- Request DTOs ---

// CreateInvoiceRequest representa o body da requisição POST /invoices.
type CreateInvoiceRequest struct {
	AmountCents int64             `json:"amount_cents"`
	DueDate     string            `json:"due_date"`
	PayerName   string            `json:"payer_name"`
	PayerTaxID  string            `json:"payer_tax_id"`
	PayerEmail  string            `json:"payer_email"`
	PayerPhone  string            `json:"payer_phone"`
	Provider    string            `json:"provider"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// --- Handlers ---

// CreateInvoice godoc
// POST /api/v1/invoices
func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_due_date", "data de vencimento inválida, use YYYY-MM-DD")
		return
	}

	input := inbound.CreateInvoiceInput{
		AmountCents: req.AmountCents,
		DueDate:     dueDate,
		PayerName:   req.PayerName,
		PayerTaxID:  req.PayerTaxID,
		PayerEmail:  req.PayerEmail,
		PayerPhone:  req.PayerPhone,
		Provider:    req.Provider,
		Metadata:    req.Metadata,
	}

	output, err := h.svc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) || errors.Is(err, domain.ErrInvalidPayer) || errors.Is(err, domain.ErrInvalidProvider) {
			respondError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "creation_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, output)
}

// GetInvoice godoc
// GET /api/v1/invoices/{id}
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da fatura é obrigatório")
		return
	}

	invoiceEntity, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "fatura não encontrada")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, invoiceEntity)
}

// ListInvoices godoc
// GET /api/v1/invoices?status=PENDING&limit=10&offset=0
func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
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

	invoices, err := h.svc.ListByStatus(r.Context(), status, limit, offset)
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, "list_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, invoices)
}

// IssueInvoice godoc
// POST /api/v1/invoices/{id}/issue
func (h *InvoiceHandler) IssueInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da fatura é obrigatório")
		return
	}

	output, err := h.svc.Issue(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "fatura não encontrada")
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "issue_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, output)
}

// PayInvoiceRequest representa o body da requisição de pagamento.
type PayInvoiceRequest struct {
	PaidAmountCents int64  `json:"paid_amount_cents"`
	PaymentMethod   string `json:"payment_method"`
}

// PayInvoice godoc
// POST /api/v1/invoices/{id}/pay
func (h *InvoiceHandler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da fatura é obrigatório")
		return
	}

	var req PayInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	output, err := h.svc.Pay(r.Context(), id, req.PaidAmountCents, req.PaymentMethod)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "fatura não encontrada")
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "pay_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, output)
}

// CancelInvoiceRequest representa o body da requisição de cancelamento.
type CancelInvoiceRequest struct {
	Reason string `json:"reason"`
}

// CancelInvoice godoc
// POST /api/v1/invoices/{id}/cancel
func (h *InvoiceHandler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da fatura é obrigatório")
		return
	}

	var req CancelInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	output, err := h.svc.Cancel(r.Context(), id, req.Reason)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "fatura não encontrada")
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "cancel_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, output)
}

// ExpireInvoice godoc
// POST /api/v1/invoices/{id}/expire
func (h *InvoiceHandler) ExpireInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "id da fatura é obrigatório")
		return
	}

	output, err := h.svc.Expire(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "fatura não encontrada")
			return
		}
		if errors.Is(err, domain.ErrInvoiceCannotBeCancelled) {
			respondError(w, http.StatusConflict, "invalid_status", "fatura não pode ser expirada no status atual")
			return
		}
		respondError(w, http.StatusUnprocessableEntity, "expire_failed", err.Error())
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