package inbound

import "context"

// ReconciliationResult define o resultado do matching.
type ReconciliationResult struct {
	Status   string
	Variance float64
	Message  string
}

// ReconciliationService define as operações do motor de reconciliação.
type ReconciliationService interface {
	// ReconcilePayment tenta fazer o matching entre expectativa e extrato bancário.
	ReconcilePayment(ctx context.Context, expectationID, providerID string, expectedAmount, paidAmount float64) (*ReconciliationResult, error)
}