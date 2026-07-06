package entity

import (
	"time"

	"github.com/google/uuid"
)

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "ACTIVE"
	AccountStatusBlocked AccountStatus = "BLOCKED"
	AccountStatusClosed  AccountStatus = "CLOSED"
)

// Account representa uma conta virtual no livro-razão.
type Account struct {
	ID         string
	MerchantID string
	Currency   string
	Status     AccountStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewAccount cria uma nova conta para um merchant.
func NewAccount(merchantID, currency string) *Account {
	now := time.Now()
	return &Account{
		ID:         uuid.New().String(),
		MerchantID: merchantID,
		Currency:   currency,
		Status:     AccountStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}