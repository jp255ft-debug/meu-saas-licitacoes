package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// PaymentRepository implementa port.PaymentRepository usando PostgreSQL + pgx.
type PaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPaymentRepository cria uma nova instÃ¢ncia do repositÃ³rio.
func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

// Save persiste um novo pagamento.
func (r *PaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	db := fromDomain(payment)

	query := `
		INSERT INTO payments (
			id, external_id, status, payment_type, method,
			amount_cents, fee_cents, net_amount_cents,
			payer_name, payer_tax_id, payer_email, payer_pix_key, payer_pix_type,
			payee_name, payee_tax_id, payee_pix_key, payee_pix_type,
			payee_bank_code, payee_branch, payee_account, payee_account_type,
			description, expires_at, processed_at, settled_at, failed_at, cancelled_at,
			failure_reason, metadata, created_at, updated_at, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24,
			$25, $26, $27, $28, $29, $30, $31, $32
		)
	`

	_, err := r.pool.Exec(ctx, query,
		db.ID, db.ExternalID, db.Status, db.PaymentType, db.Method,
		db.AmountCents, db.FeeCents, db.NetAmountCents,
		db.PayerName, db.PayerTaxID, db.PayerEmail, db.PayerPixKey, db.PayerPixType,
		db.PayeeName, db.PayeeTaxID, db.PayeePixKey, db.PayeePixType,
		db.PayeeBankCode, db.PayeeBranch, db.PayeeAccount, db.PayeeAccountType,
		db.Description, db.ExpiresAt, db.ProcessedAt, db.SettledAt, db.FailedAt, db.CancelledAt,
		db.FailureReason, db.Metadata, db.CreatedAt, db.UpdatedAt, db.Version,
	)
	if err != nil {
		return fmt.Errorf("falha ao inserir pagamento: %w", err)
	}

	return nil
}

// FindByID busca um pagamento pelo ID interno.
func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `SELECT * FROM payments WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	db, err := scanPayment(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("falha ao buscar pagamento: %w", err)
	}

	return db.toDomain()
}

// FindByExternalID busca um pagamento pelo ID do gateway externo.
func (r *PaymentRepository) FindByExternalID(ctx context.Context, externalID string) (*domain.Payment, error) {
	query := `SELECT * FROM payments WHERE external_id = $1`
	row := r.pool.QueryRow(ctx, query, externalID)

	db, err := scanPayment(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("falha ao buscar pagamento por externalID: %w", err)
	}

	return db.toDomain()
}

// Update atualiza o estado de um pagamento existente com optimistic locking.
func (r *PaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	db := fromDomain(payment)
	prevVersion := db.Version - 1

	query := `
		UPDATE payments SET
			status = $1, external_id = $2, processed_at = $3, settled_at = $4,
			failed_at = $5, cancelled_at = $6, failure_reason = $7,
			updated_at = $8, version = $9
		WHERE id = $10 AND version = $11
	`

	cmdTag, err := r.pool.Exec(ctx, query,
		db.Status, db.ExternalID, db.ProcessedAt, db.SettledAt,
		db.FailedAt, db.CancelledAt, db.FailureReason,
		db.UpdatedAt, db.Version,
		db.ID, prevVersion,
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar pagamento: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrPaymentAlreadyProcessed // optimistic locking conflict
	}

	return nil
}

// ListByStatus lista pagamentos por status com paginaÃ§Ã£o.
func (r *PaymentRepository) ListByStatus(ctx context.Context, status domain.PaymentStatus, limit, offset int) ([]*domain.Payment, error) {
	query := `SELECT * FROM payments WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar pagamentos: %w", err)
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// ListByPayerTaxID lista pagamentos de um pagador especÃ­fico.
func (r *PaymentRepository) ListByPayerTaxID(ctx context.Context, taxID string, limit, offset int) ([]*domain.Payment, error) {
	query := `SELECT * FROM payments WHERE payer_tax_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, taxID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar pagamentos: %w", err)
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// ListExpired lista pagamentos pendentes que jÃ¡ expiraram.
func (r *PaymentRepository) ListExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Payment, error) {
	query := `SELECT * FROM payments WHERE status = 'PENDING' AND expires_at < $1 LIMIT $2`
	rows, err := r.pool.Query(ctx, query, before, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar pagamentos expirados: %w", err)
	}
	defer rows.Close()

	return r.scanPayments(rows)
}

// CountByStatus retorna a contagem de pagamentos por status.
func (r *PaymentRepository) CountByStatus(ctx context.Context, status domain.PaymentStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM payments WHERE status = $1`
	var count int64
	if err := r.pool.QueryRow(ctx, query, string(status)).Scan(&count); err != nil {
		return 0, fmt.Errorf("falha ao contar pagamentos: %w", err)
	}
	return count, nil
}

// scanPayments converte um pgx.Rows em slice de *domain.Payment.
func (r *PaymentRepository) scanPayments(rows pgx.Rows) ([]*domain.Payment, error) {
	var payments []*domain.Payment
	for rows.Next() {
		db, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		p, err := db.toDomain()
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

// Compile-time check: PaymentRepository implementa port.PaymentRepository
var _ port.PaymentRepository = (*PaymentRepository)(nil)
