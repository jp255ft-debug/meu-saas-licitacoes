package transfer

import (
	"time"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
)

// MapToTransferOutput converte um entity.Transfer para TransferOutput.
func MapToTransferOutput(t *entity.Transfer) *inbound.TransferOutput {
	return &inbound.TransferOutput{
		ID:            t.ID(),
		AmountCents:   t.AmountCents(),
		Status:        string(t.Status()),
		PayeeName:     t.Payee().Name,
		PayeeTaxID:    t.Payee().TaxID,
		PayeePixKey:   t.Payee().PixKey,
		Provider:      string(t.Provider()),
		ExternalID:    t.ExternalID(),
		Description:   t.Description(),
		FailureReason: t.FailureReason(),
		CreatedAt:     t.CreatedAt().Format(time.RFC3339),
		UpdatedAt:     t.UpdatedAt().Format(time.RFC3339),
	}
}