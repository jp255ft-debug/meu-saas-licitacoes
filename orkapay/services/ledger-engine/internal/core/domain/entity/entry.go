package entity

import (
	"time"

	"github.com/google/uuid"
)

type EntryType string

const (
	EntryTypeCredit EntryType = "CREDIT"
	EntryTypeDebit  EntryType = "DEBIT"
)

// LedgerEntry representa uma entrada no livro-razão (partidas dobradas).
type LedgerEntry struct {
	ID              string
	AccountID       string
	TransactionID   string
	EntryType       EntryType
	Amount          float64
	OperationType   string
	IdempotencyKey  string
	CreatedAt       time.Time
}

// NewCreditEntry cria uma entrada de crédito.
func NewCreditEntry(accountID, transactionID, operationType, idempotencyKey string, amount float64) *LedgerEntry {
	return &LedgerEntry{
		ID:             uuid.New().String(),
		AccountID:      accountID,
		TransactionID:  transactionID,
		EntryType:      EntryTypeCredit,
		Amount:         amount,
		OperationType:  operationType,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}
}

// NewDebitEntry cria uma entrada de débito.
func NewDebitEntry(accountID, transactionID, operationType, idempotencyKey string, amount float64) *LedgerEntry {
	return &LedgerEntry{
		ID:             uuid.New().String(),
		AccountID:      accountID,
		TransactionID:  transactionID,
		EntryType:      EntryTypeDebit,
		Amount:         amount,
		OperationType:  operationType,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}
}