package inbound

import (
	"context"
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
)

// CreateInvoiceInput é o input para criação de fatura.
type CreateInvoiceInput struct {
	AmountCents int64
	DueDate     time.Time
	PayerName   string
	PayerTaxID  string
	PayerEmail  string
	PayerPhone  string
	Provider    string
	Metadata    map[string]string
}

// InvoiceOutput é a saída padronizada para operações com fatura.
type InvoiceOutput struct {
	ID              string
	AmountCents     int64
	Status          string
	DueDate         string
	PayerName       string
	PayerTaxID      string
	Provider        string
	ExternalID      string
	PixQRCode       string
	PixKey          string
	BoletoBarcode   string
	BoletoDigitable string
	CreatedAt       string
	UpdatedAt       string
}

// InvoiceService define os casos de uso do domínio de cash-in.
type InvoiceService interface {
	Create(ctx context.Context, input CreateInvoiceInput) (*InvoiceOutput, error)
	Issue(ctx context.Context, id string) (*InvoiceOutput, error)
	GetByID(ctx context.Context, id string) (*entity.Invoice, error)
	Pay(ctx context.Context, id string, paidAmountCents int64, paymentMethod string) (*InvoiceOutput, error)
	Cancel(ctx context.Context, id, reason string) (*InvoiceOutput, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entity.Invoice, error)
}
