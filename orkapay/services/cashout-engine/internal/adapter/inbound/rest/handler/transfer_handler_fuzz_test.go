//go:build gofuzz

package handler

import (
	"encoding/json"
	"testing"

	"github.com/jp255ft-debug/cashout-engine/internal/adapter/inbound/rest/dto"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
)

func FuzzCreateTransferPayload(f *testing.F) {
	// Seeds com payloads válidos e inválidos
	f.Add([]byte(`{"amount_cents":50000,"payee_name":"Jane Doe","payee_tax_id":"98765432101","payee_pix_key":"jane@test.com","payee_pix_type":"EMAIL","provider":"STARK_BANK","description":"Teste"}`))
	f.Add([]byte(`{"amount_cents":0,"payee_name":"","payee_tax_id":"","payee_pix_key":"","description":""}`))
	f.Add([]byte(`{"amount_cents":-1,"payee_name":"","payee_tax_id":"' OR '1'='1","payee_pix_key":"","description":"test"}`))
	f.Add([]byte(`{"amount_cents":999999999999999999,"payee_name":"A","payee_tax_id":"12","payee_pix_key":"x","description":"y"}`))
	f.Add([]byte(`not-json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var req dto.CreateTransferRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return // Payload inválido é esperado, não é bug
		}

		// Tenta criar transfer com dados do parse
		payee := entity.TransferPayee{
			Name:    req.PayeeName,
			TaxID:   req.PayeeTaxID,
			PixKey:  req.PayeePixKey,
			PixType: req.PayeePixType,
		}

		_, err := entity.NewTransfer(
			req.AmountCents,
			payee,
			entity.CashOutProvider(req.Provider),
			req.Description,
			req.Metadata,
		)
		// Não deve panic em nenhum caso
		_ = err
	})
}