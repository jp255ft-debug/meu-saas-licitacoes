package asaas

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

// ValidateSignature valida a assinatura do webhook (header asaas-signature).
func ValidateSignature(body []byte, signature, apiKey string) error {
	if signature == "" {
		return fmt.Errorf("assinatura do webhook ausente")
	}
	if apiKey == "" {
		return fmt.Errorf("apiKey nao configurada para validacao de webhook")
	}

	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return fmt.Errorf("assinatura do webhook invalida")
	}

	return nil
}

// ParseWebhook parseia o corpo do webhook para um evento Asaas.
func ParseWebhook(body io.Reader) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.NewDecoder(body).Decode(&event); err != nil {
		return nil, fmt.Errorf("erro ao decodificar webhook: %w", err)
	}

	if event.Event == "" {
		return nil, fmt.Errorf("webhook invalido: evento vazio")
	}
	if event.Payment.ID == "" {
		return nil, fmt.Errorf("webhook invalido: payment ID vazio")
	}

	return &event, nil
}