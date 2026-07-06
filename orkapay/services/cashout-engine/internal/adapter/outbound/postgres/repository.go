package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// TransferRepository implementa outbound.TransferRepository usando PostgreSQL + pgx.
type TransferRepository struct {
	pool *pgxpool.Pool
}

// NewTransferRepository cria uma nova instância do repositório de transferências.
func NewTransferRepository(pool *pgxpool.Pool) *TransferRepository {
	return &TransferRepository{pool: pool}
}

// Save persiste uma nova transferência.
func (r *TransferRepository) Save(ctx context.Context, transfer *entity.Transfer) error {
	db := fromDomain(transfer)

	query := `
		INSERT INTO transfers (
			id, external_id, status,
			amount_cents,
			payee_name, payee_tax_id, payee_pix_key, payee_pix_type,
			payee_bank_code, payee_branch, payee_account, payee_account_type,
			provider, description,
			metadata,
			processed_at, settled_at, failed_at, cancelled_at,
			failure_reason,
			version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20,
			$21, $22, $23
		)
	`

	_, err := r.pool.Exec(ctx, query,
		db.ID, db.ExternalID, db.Status,
		db.AmountCents,
		db.PayeeName, db.PayeeTaxID, db.PayeePixKey, db.PayeePixType,
		db.PayeeBankCode, db.PayeeBranch, db.PayeeAccount, db.PayeeAccountType,
		db.Provider, db.Description,
		db.Metadata,
		db.ProcessedAt, db.SettledAt, db.FailedAt, db.CancelledAt,
		db.FailureReason,
		db.Version, db.CreatedAt, db.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("falha ao inserir transferência: %w", err)
	}

	return nil
}

// FindByID busca uma transferência pelo ID interno.
func (r *TransferRepository) FindByID(ctx context.Context, id string) (*entity.Transfer, error) {
	query := `SELECT * FROM transfers WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	db, err := scanTransfer(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrTransferNotFound
		}
		return nil, fmt.Errorf("falha ao buscar transferência: %w", err)
	}

	return db.toDomain()
}

// FindByExternalID busca uma transferência pelo ID externo (provedor).
func (r *TransferRepository) FindByExternalID(ctx context.Context, externalID string) (*entity.Transfer, error) {
	query := `SELECT * FROM transfers WHERE external_id = $1`
	row := r.pool.QueryRow(ctx, query, externalID)

	db, err := scanTransfer(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrTransferNotFound
		}
		return nil, fmt.Errorf("falha ao buscar transferência por externalID: %w", err)
	}

	return db.toDomain()
}

// Update atualiza o estado de uma transferência existente com optimistic locking.
func (r *TransferRepository) Update(ctx context.Context, transfer *entity.Transfer) error {
	db := fromDomain(transfer)
	prevVersion := db.Version - 1

	query := `
		UPDATE transfers SET
			status = $1, external_id = $2,
			processed_at = $3, settled_at = $4, failed_at = $5,
			cancelled_at = $6, failure_reason = $7,
			version = $8, updated_at = $9
		WHERE id = $10 AND version = $11
	`

	cmdTag, err := r.pool.Exec(ctx, query,
		db.Status, db.ExternalID,
		db.ProcessedAt, db.SettledAt, db.FailedAt,
		db.CancelledAt, db.FailureReason,
		db.Version, db.UpdatedAt,
		db.ID, prevVersion,
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar transferência: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("transferência já foi modificada por outra operação: %w", domain.ErrTransferCannotBeCancelled)
	}

	return nil
}

// ListByStatus lista transferências por status com paginação.
func (r *TransferRepository) ListByStatus(ctx context.Context, status entity.TransferStatus, limit, offset int) ([]*entity.Transfer, error) {
	query := `SELECT * FROM transfers WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar transferências: %w", err)
	}
	defer rows.Close()

	return r.scanTransfers(rows)
}

// CountByStatus retorna a contagem de transferências por status.
func (r *TransferRepository) CountByStatus(ctx context.Context, status entity.TransferStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM transfers WHERE status = $1`
	var count int64
	if err := r.pool.QueryRow(ctx, query, string(status)).Scan(&count); err != nil {
		return 0, fmt.Errorf("falha ao contar transferências: %w", err)
	}
	return count, nil
}

// ListByPayeeTaxID lista transferências de um beneficiário específico.
func (r *TransferRepository) ListByPayeeTaxID(ctx context.Context, taxID string, limit, offset int) ([]*entity.Transfer, error) {
	query := `SELECT * FROM transfers WHERE payee_tax_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, taxID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar transferências por beneficiário: %w", err)
	}
	defer rows.Close()

	return r.scanTransfers(rows)
}

// ListExpired lista transferências pendentes que não foram processadas dentro do prazo.
func (r *TransferRepository) ListExpired(ctx context.Context, referenceTime interface{}, limit int) ([]*entity.Transfer, error) {
	beforeTime, ok := referenceTime.(time.Time)
	if !ok {
		return nil, fmt.Errorf("referenceTime deve ser time.Time, recebeu %T", referenceTime)
	}
	query := `SELECT * FROM transfers WHERE status = 'PENDING' AND created_at < $1 ORDER BY created_at ASC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, beforeTime, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar transferências expiradas: %w", err)
	}
	defer rows.Close()

	return r.scanTransfers(rows)
}

// scanTransfers converte um pgx.Rows em slice de *entity.Transfer.
func (r *TransferRepository) scanTransfers(rows pgx.Rows) ([]*entity.Transfer, error) {
	var transfers []*entity.Transfer
	for rows.Next() {
		db, err := scanTransfer(rows)
		if err != nil {
			return nil, err
		}
		t, err := db.toDomain()
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}
	return transfers, rows.Err()
}

// Compile-time check: TransferRepository implementa outbound.TransferRepository
var _ outbound.TransferRepository = (*TransferRepository)(nil)