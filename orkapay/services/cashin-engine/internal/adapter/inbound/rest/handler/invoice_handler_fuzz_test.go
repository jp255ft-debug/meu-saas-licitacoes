//go:build gofuzz

package handler

import (
	"encoding/json"
	"testing"

	"github.com/jp255ft-debug/cashin-engine/internal/adapter/inbound/rest/dto"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
)

func FuzzCreateInvoicePayload(f *testing.F) {
	// Seeds com payloads válidos e inválidos
	f.Add([]byte(`{"amount_cents":10000,"due_date":"2026-07-10","payer_name":"João","payer_tax_id":"12345678901","payer_email":"joao@test.com","provider":"ASAAS"}`))
	f.Add([]byte(`{"amount_cents":0,"due_date":"","payer_name":"","payer_tax_id":""}`))
	f.Add([]byte(`{"amount_cents":-1,"due_date":"invalid","payer_name":"","payer_tax_id":"' OR '1'='1"}`))
	f.Add([]byte(`{"amount_cents":999999999999999999,"due_date":"2026-07-10","payer_name":"A","payer_tax_id":"12"}`))
	f.Add([]byte(`not-json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var req dto.CreateInvoiceRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return // Payload inválido é esperado, não é bug
		}

		// Tenta criar invoice com dados do parse
		// Se o parse foi bem-sucedido, valida o domínio
		payer := domain.InvoicePayer{
			Name:  req.PayerName,
			TaxID: req.PayerTaxID,
			Email: req.PayerEmail,
		}

		_, err := domain.NewInvoice(
			req.AmountCents,
			req.DueDate,
			payer,
			domain.CashInProvider(req.Provider),
			req.Metadata,
		)
		// Não deve panic em nenhum caso
		_ = err
	})
}