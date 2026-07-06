package reconciliation

import (
	"context"
	"fmt"
	"math"

	"github.com/jp255ft-debug/reconciliation/internal/core/port/inbound"
)

const fuzzyMatchTolerance = 0.05 // 5%

// ReconciliationServiceImpl implementa inbound.ReconciliationService.
type ReconciliationServiceImpl struct{}

// NewReconciliationService cria uma nova instância.
func NewReconciliationService() *ReconciliationServiceImpl {
	return &ReconciliationServiceImpl{}
}

func (s *ReconciliationServiceImpl) ReconcilePayment(ctx context.Context, expectationID, providerID string, expectedAmount, paidAmount float64) (*inbound.ReconciliationResult, error) {
	// 1. Exact Match
	if providerID == expectationID && expectedAmount == paidAmount {
		return &inbound.ReconciliationResult{
			Status:   "EXACT_MATCH",
			Variance: 0,
			Message:  "Exact match between expectation and payment",
		}, nil
	}

	// 2. Fuzzy Match (5% tolerance)
	variance := paidAmount - expectedAmount
	maxAllowed := expectedAmount * fuzzyMatchTolerance

	if providerID == expectationID && math.Abs(variance) <= maxAllowed {
		return &inbound.ReconciliationResult{
			Status:   "FUZZY_MATCH",
			Variance: variance,
			Message:  fmt.Sprintf("Fuzzy match with variance of %.2f (within %.0f%% tolerance)", variance, fuzzyMatchTolerance*100),
		}, nil
	}

	// 3. Discrepancy
	if providerID == expectationID {
		return &inbound.ReconciliationResult{
			Status:   "DISCREPANCY",
			Variance: variance,
			Message:  fmt.Sprintf("Amount diverges by %.2f (exceeds %.0f%% tolerance)", variance, fuzzyMatchTolerance*100),
		}, nil
	}

	// 4. Not Found
	return &inbound.ReconciliationResult{
		Status:   "NOT_FOUND",
		Variance: 0,
		Message:  "Provider ID mismatch - no matching expectation found",
	}, nil
}