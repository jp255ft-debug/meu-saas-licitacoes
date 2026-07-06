package postgres

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
)

// invoiceDB é a representação flat da fatura no banco de dados.
type invoiceDB struct {
	ID                string
	ExternalID        *string
	InvoiceNumber     string
	Status            string
	AmountCents       int64
	DueDate           time.Time
	PayerName         string
	PayerTaxID        string
	PayerEmail        *string
	PayerPhone        *string
	PayerAddressLine1 *string
	PayerAddressLine2 *string
	PayerCity         *string
	PayerState        *string
	PayerZipCode      *string
	Provider          string
	PixQRCode         *string
	PixKey            *string
	BoletoBarcode     *string
	BoletoDigitable   *string
	Metadata          map[string]string
	PaidAmountCents   int64
	PaidAt            *time.Time
	PaymentMethod     *string
	CancelledAt       *time.Time
	FailureReason     *string
	Version           int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// fromDomain converte um entity.Invoice para invoiceDB.
func fromDomain(inv *entity.Invoice) *invoiceDB {
	return &invoiceDB{
		ID:                inv.ID(),
		ExternalID:        strPtr(inv.ExternalID()),
		InvoiceNumber:     inv.ID(), // invoice_number = ID interno
		Status:            string(inv.Status()),
		AmountCents:       inv.AmountCents(),
		DueDate:           inv.DueDate(),
		PayerName:         inv.Payer().Name,
		PayerTaxID:        inv.Payer().TaxID,
		PayerEmail:        strPtr(inv.Payer().Email),
		PayerPhone:        strPtr(inv.Payer().Phone),
		PayerAddressLine1: strPtr(inv.Payer().AddressLine1),
		PayerAddressLine2: strPtr(inv.Payer().AddressLine2),
		PayerCity:         strPtr(inv.Payer().City),
		PayerState:        strPtr(inv.Payer().State),
		PayerZipCode:      strPtr(inv.Payer().ZipCode),
		Provider:          string(inv.Provider()),
		PixQRCode:         strPtr(inv.PixQRCode()),
		PixKey:            strPtr(inv.PixKey()),
		BoletoBarcode:     strPtr(inv.BoletoBarcode()),
		BoletoDigitable:   strPtr(inv.BoletoDigitable()),
		Metadata:          inv.Metadata(),
		PaidAmountCents:   inv.PaidAmountCents(),
		PaidAt:            inv.PaidAt(),
		PaymentMethod:     strPtr(inv.PaymentMethod()),
		CancelledAt:       nil, // entity não expõe CancelledAt como getter
		FailureReason:     strPtr(inv.FailureReason()),
		Version:           inv.Version(),
		CreatedAt:         inv.CreatedAt(),
		UpdatedAt:         inv.UpdatedAt(),
	}
}

// toDomain reconstrói um entity.Invoice a partir do invoiceDB.
func (db *invoiceDB) toDomain() (*entity.Invoice, error) {
	payer := entity.InvoicePayer{
		Name:         db.PayerName,
		TaxID:        db.PayerTaxID,
		Email:        safeStr(db.PayerEmail),
		Phone:        safeStr(db.PayerPhone),
		AddressLine1: safeStr(db.PayerAddressLine1),
		AddressLine2: safeStr(db.PayerAddressLine2),
		City:         safeStr(db.PayerCity),
		State:        safeStr(db.PayerState),
		ZipCode:      safeStr(db.PayerZipCode),
	}

	inv, err := entity.NewInvoice(
		db.AmountCents,
		db.DueDate,
		payer,
		entity.CashInProvider(db.Provider),
		db.Metadata,
	)
	if err != nil {
		return nil, err
	}

	// Restaurar estado após criação (precisamos usar reflection ou setters privados - optamos por
	// reconstruir via métodos do domínio conforme o status)
	// Como os campos são privados e não temos setters, restauramos através dos métodos do aggregate.
	// Estado inicial é PENDING. Se o status for diferente, aplicamos as transições.

	switch entity.InvoiceStatus(db.Status) {
	case entity.InvoiceStatusIssued:
		if err := inv.Issue(safeStr(db.ExternalID), safeStr(db.PixQRCode), safeStr(db.PixKey)); err != nil {
			return nil, err
		}
	case entity.InvoiceStatusPaid:
		if err := inv.Issue(safeStr(db.ExternalID), safeStr(db.PixQRCode), safeStr(db.PixKey)); err != nil {
			return nil, err
		}
		if err := inv.Pay(db.PaidAmountCents, safeStr(db.PaymentMethod)); err != nil {
			return nil, err
		}
	case entity.InvoiceStatusCancelled:
		inv.Cancel(safeStr(db.FailureReason))
	case entity.InvoiceStatusExpired:
		inv.MarkExpired()
	case entity.InvoiceStatusRefunded:
		// Para REFUNDED, partimos do status ISSUED e então marcamos como pago e estornamos
		// Como não temos método específico, usamos Cancel com motivo
		inv.Cancel("refunded")
	}

	return inv, nil
}

// scanInvoice lê uma linha do pgx.Row (QueryRow ou Rows.Next).
func scanInvoice(row pgx.Row) (*invoiceDB, error) {
	var db invoiceDB
	err := row.Scan(
		&db.ID, &db.ExternalID, &db.InvoiceNumber, &db.Status,
		&db.AmountCents, &db.DueDate,
		&db.PayerName, &db.PayerTaxID, &db.PayerEmail, &db.PayerPhone,
		&db.PayerAddressLine1, &db.PayerAddressLine2, &db.PayerCity, &db.PayerState, &db.PayerZipCode,
		&db.Provider, &db.PixQRCode, &db.PixKey,
		&db.BoletoBarcode, &db.BoletoDigitable,
		&db.Metadata,
		&db.PaidAmountCents, &db.PaidAt, &db.PaymentMethod,
		&db.CancelledAt, &db.FailureReason,
		&db.Version, &db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

// strPtr converte string para *string (nil se vazia).
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// safeStr extrai string de *string, retornando vazio se nil.
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

