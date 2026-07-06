package invoice

import (
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
)

// MapToInvoiceOutput converte um entity.Invoice para InvoiceOutput.
func MapToInvoiceOutput(inv *entity.Invoice) *inbound.InvoiceOutput {
	return &inbound.InvoiceOutput{
		ID:              inv.ID(),
		AmountCents:     inv.AmountCents(),
		Status:          string(inv.Status()),
		DueDate:         inv.DueDate().Format("2006-01-02"),
		PayerName:       inv.Payer().Name,
		PayerTaxID:      inv.Payer().TaxID,
		Provider:        string(inv.Provider()),
		ExternalID:      inv.ExternalID(),
		PixQRCode:       inv.PixQRCode(),
		PixKey:          inv.PixKey(),
		BoletoBarcode:   inv.BoletoBarcode(),
		BoletoDigitable: inv.BoletoDigitable(),
		CreatedAt:       inv.CreatedAt().Format(time.RFC3339),
		UpdatedAt:       inv.UpdatedAt().Format(time.RFC3339),
	}
}