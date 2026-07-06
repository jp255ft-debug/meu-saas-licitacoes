package inbound

import (
	"context"
	"time"

	"github.com/jp255ft-debug/boleto-engine/internal/core/domain/entity"
)

// CreateBoletoInput é o input para criação de boleto.
type CreateBoletoInput struct {
	BoletoNumber string
	AmountCents  int64
	DueDate      time.Time
	PayerName    string
	PayerTaxID   string
	PayerEmail   string
	PayerPhone   string
	PayerAddress string
	Provider     string
	Metadata     map[string]string
}

// BoletoOutput é a saída padronizada para operações com boleto.
type BoletoOutput struct {
	ID           string
	BoletoNumber string
	Barcode      string
	Digitable    string
	Status       string
	AmountCents  int64
	DueDate      string
	PayerName    string
	PayerTaxID   string
	Provider     string
	ExternalID   string
	CreatedAt    string
	UpdatedAt    string
}

// BoletoService define os casos de uso do domínio de boletos.
type BoletoService interface {
	Create(ctx context.Context, input CreateBoletoInput) (*BoletoOutput, error)
	Issue(ctx context.Context, id string) (*BoletoOutput, error)
	GetByID(ctx context.Context, id string) (*entity.Boleto, error)
	Pay(ctx context.Context, id string, paidAmountCents int64, paymentID string) (*BoletoOutput, error)
	Cancel(ctx context.Context, id, reason string) (*BoletoOutput, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entity.Boleto, error)
}
