package domain

import "time"

type Command struct {
	RawMessage  string
	CommandType string // "CREATE_INVOICE", "HELP", "BALANCE"
	Amount      float64
	PayerName   string
	PayerTaxID  string
	Description string
	ParsedAt    time.Time
}