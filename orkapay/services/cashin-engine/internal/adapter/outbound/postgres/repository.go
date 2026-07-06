package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// InvoiceRepository implementa outbound.InvoiceRepository usando PostgreSQL + pgx.
type InvoiceRepository struct {
	pool *pgxpool.Pool
}

// NewInvoiceRepository cria uma nova instância do repositório de faturas.
func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{pool: pool}
}

// Save persiste uma nova fatura.
func (r *InvoiceRepository) Save(ctx context.Context, invoice *entity.Invoice) error {
	db := fromDomain(invoice)

	query := `
		INSERT INTO invoices (
			id, external_id, invoice_number, status,
			amount_cents, due_date,
			payer_name, payer_tax_id, payer_email, payer_phone,
			payer_address_line1, payer_address_line2, payer_city, payer_state, payer_zip_code,
			provider, pix_qr_code, pix_key,
			boleto_barcode, boleto_digitable,
			metadata,
			paid_amount_cents, paid_at, payment_method,
			cancelled_at, failure_reason,
			version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29
		)
	`

	_, err := r.pool.Exec(ctx, query,
		db.ID, db.ExternalID, db.InvoiceNumber, db.Status,
		db.AmountCents, db.DueDate,
		db.PayerName, db.PayerTaxID, db.PayerEmail, db.PayerPhone,
		db.PayerAddressLine1, db.PayerAddressLine2, db.PayerCity, db.PayerState, db.PayerZipCode,
		db.Provider, db.PixQRCode, db.PixKey,
		db.BoletoBarcode, db.BoletoDigitable,
		db.Metadata,
		db.PaidAmountCents, db.PaidAt, db.PaymentMethod,
		db.CancelledAt, db.FailureReason,
		db.Version, db.CreatedAt, db.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("falha ao inserir fatura: %w", err)
	}

	return nil
}

// FindByID busca uma fatura pelo ID interno.
func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (*entity.Invoice, error) {
	query := `SELECT * FROM invoices WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	db, err := scanInvoice(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("falha ao buscar fatura: %w", err)
	}

	return db.toDomain()
}

// FindByExternalID busca uma fatura pelo ID externo (provedor).
func (r *InvoiceRepository) FindByExternalID(ctx context.Context, externalID string) (*entity.Invoice, error) {
	query := `SELECT * FROM invoices WHERE external_id = $1`
	row := r.pool.QueryRow(ctx, query, externalID)

	db, err := scanInvoice(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("falha ao buscar fatura por externalID: %w", err)
	}

	return db.toDomain()
}

// Update atualiza o estado de uma fatura existente com optimistic locking.
func (r *InvoiceRepository) Update(ctx context.Context, invoice *entity.Invoice) error {
	db := fromDomain(invoice)
	prevVersion := db.Version - 1

	query := `
		UPDATE invoices SET
			status = $1, external_id = $2, pix_qr_code = $3, pix_key = $4,
			boleto_barcode = $5, boleto_digitable = $6,
			paid_amount_cents = $7, paid_at = $8, payment_method = $9,
			cancelled_at = $10, failure_reason = $11,
			version = $12, updated_at = $13
		WHERE id = $14 AND version = $15
	`

	cmdTag, err := r.pool.Exec(ctx, query,
		db.Status, db.ExternalID, db.PixQRCode, db.PixKey,
		db.BoletoBarcode, db.BoletoDigitable,
		db.PaidAmountCents, db.PaidAt, db.PaymentMethod,
		db.CancelledAt, db.FailureReason,
		db.Version, db.UpdatedAt,
		db.ID, prevVersion,
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar fatura: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("fatura já foi modificada por outra operação: %w", domain.ErrInvoiceCannotBeCancelled)
	}

	return nil
}

// ListByStatus lista faturas por status com paginação.
func (r *InvoiceRepository) ListByStatus(ctx context.Context, status entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error) {
	query := `SELECT * FROM invoices WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faturas: %w", err)
	}
	defer rows.Close()

	return r.scanInvoices(rows)
}

// ListByPayerTaxID lista faturas de um pagador específico.
func (r *InvoiceRepository) ListByPayerTaxID(ctx context.Context, taxID string, limit, offset int) ([]*entity.Invoice, error) {
	query := `SELECT * FROM invoices WHERE payer_tax_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, taxID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faturas por pagador: %w", err)
	}
	defer rows.Close()

	return r.scanInvoices(rows)
}

// ListExpired lista faturas pendentes ou emitidas que já venceram.
func (r *InvoiceRepository) ListExpired(ctx context.Context, referenceTime interface{}, limit int) ([]*entity.Invoice, error) {
	beforeTime, ok := referenceTime.(time.Time)
	if !ok {
		return nil, fmt.Errorf("referenceTime deve ser time.Time, recebeu %T", referenceTime)
	}
	query := `SELECT * FROM invoices WHERE status IN ('PENDING', 'ISSUED') AND due_date < $1 ORDER BY due_date ASC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, beforeTime, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faturas expiradas: %w", err)
	}
	defer rows.Close()

	return r.scanInvoices(rows)
}

// CountByStatus retorna a contagem de faturas por status.
func (r *InvoiceRepository) CountByStatus(ctx context.Context, status entity.InvoiceStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM invoices WHERE status = $1`
	var count int64
	if err := r.pool.QueryRow(ctx, query, string(status)).Scan(&count); err != nil {
		return 0, fmt.Errorf("falha ao contar faturas: %w", err)
	}
	return count, nil
}

// scanInvoices converte um pgx.Rows em slice de *entity.Invoice.
func (r *InvoiceRepository) scanInvoices(rows pgx.Rows) ([]*entity.Invoice, error) {
	var invoices []*entity.Invoice
	for rows.Next() {
		db, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		inv, err := db.toDomain()
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}

// Compile-time check: InvoiceRepository implementa outbound.InvoiceRepository
var _ outbound.InvoiceRepository = (*InvoiceRepository)(nil)