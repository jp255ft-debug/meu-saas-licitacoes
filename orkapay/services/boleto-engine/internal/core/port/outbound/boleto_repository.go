package outbound

import (
	"context"

	"github.com/jp255ft-debug/boleto-engine/internal/core/domain/entity"
)

// BoletoRepository define o contrato para persistência de boletos.
type BoletoRepository interface {
	Save(ctx context.Context, boleto *entity.Boleto) error
	FindByID(ctx context.Context, id string) (*entity.Boleto, error)
	FindByBoletoNumber(ctx context.Context, boletoNumber string) (*entity.Boleto, error)
	FindByExternalID(ctx context.Context, externalID string) (*entity.Boleto, error)
	Update(ctx context.Context, boleto *entity.Boleto) error
	ListByStatus(ctx context.Context, status entity.BoletoStatus, limit, offset int) ([]*entity.Boleto, error)
	ListExpired(ctx context.Context, referenceTime interface{}, limit int) ([]*entity.Boleto, error)
	CountByStatus(ctx context.Context, status entity.BoletoStatus) (int64, error)
}

// BoletoProviderGateway define o contrato para integração com provedores de boleto.
type BoletoProviderGateway interface {
	IssueBoleto(ctx context.Context, boleto *entity.Boleto) (barcode, digitable, externalID string, err error)
	GetBoletoStatus(ctx context.Context, externalID string) (status string, err error)
	CancelBoleto(ctx context.Context, externalID string) error
}
