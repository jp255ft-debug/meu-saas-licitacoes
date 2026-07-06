package domain

import "time"

// ReconciliationStatus representa o resultado do matching.
type ReconciliationStatus string

const (
	StatusExactMatch   ReconciliationStatus = "EXACT_MATCH"
	StatusFuzzyMatch   ReconciliationStatus = "FUZZY_MATCH"
	StatusDiscrepancy  ReconciliationStatus = "DISCREPANCY"
	StatusNotFound     ReconciliationStatus = "NOT_FOUND"
)

// PaymentExpectation representa o que deveria ter sido pago.
type PaymentExpectation struct {
	ID            string
	Amount        float64
	DueDate       time.Time
	ProviderID    string
	MerchantID    string
}

// BankStatementLine representa o que foi recebido pelo BaaS.
type BankStatementLine struct {
	ProviderID  string
	AmountPaid  float64
	PaidAt      time.Time
}

// ReconciliationResult representa o resultado do matching.
type ReconciliationResult struct {
	Status   ReconciliationStatus
	Variance float64
	Message  string
}