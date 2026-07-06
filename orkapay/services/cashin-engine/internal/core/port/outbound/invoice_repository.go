package outbound

import (
	"context"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
)

// InvoiceRepository define o contrato para persistência de faturas.
type InvoiceRepository interface {
	Save(ctx context.Context, invoice *entity.Invoice) error
	FindByID(ctx context.Context, id string) (*entity.Invoice, error)
	FindByExternalID(ctx context.Context, externalID string) (*entity.Invoice, error)
	Update(ctx context.Context, invoice *entity.Invoice) error
	ListByStatus(ctx context.Context, status entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error)
	ListExpired(ctx context.Context, referenceTime interface{}, limit int) ([]*entity.Invoice, error)
	CountByStatus(ctx context.Context, status entity.InvoiceStatus) (int64, error)
}

// CashInProviderGateway define o contrato para integração com provedores de cash-in.
type CashInProviderGateway interface {
	CreateInvoice(ctx context.Context, invoice *entity.Invoice) (externalID, pixQRCode, pixKey string, err error)
	GetInvoiceStatus(ctx context.Context, externalID string) (status string, err error)
	CancelInvoice(ctx context.Context, externalID string) error
}
