package postgres

import (
	"github.com/jp255ft-debug/ledger-engine/internal/core/domain/entity"
)

// accountRow representa uma linha da tabela ledger_accounts.
type accountRow struct {
	ID         string
	MerchantID string
	Currency   string
	Status     string
	CreatedAt  string
	UpdatedAt  string
}

func (r *accountRow) toDomain() *entity.Account {
	return &entity.Account{
		ID:         r.ID,
		MerchantID: r.MerchantID,
		Currency:   r.Currency,
		Status:     entity.AccountStatus(r.Status),
	}
}

// entryRow representa uma linha da tabela ledger_entries.
type entryRow struct {
	ID             string
	AccountID      string
	TransactionID  string
	EntryType      string
	Amount         float64
	OperationType  string
	IdempotencyKey string
	CreatedAt      string
}

func (r *entryRow) toDomain() *entity.LedgerEntry {
	return &entity.LedgerEntry{
		ID:             r.ID,
		AccountID:      r.AccountID,
		TransactionID:  r.TransactionID,
		EntryType:      entity.EntryType(r.EntryType),
		Amount:         r.Amount,
		OperationType:  r.OperationType,
		IdempotencyKey: r.IdempotencyKey,
	}
}