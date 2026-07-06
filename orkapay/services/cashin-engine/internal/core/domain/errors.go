package domain

import "errors"

// Erros sentinela do domínio cashin-engine.
var (
	ErrInvalidInvoice      = errors.New("fatura inválida")
	ErrInvoiceNotFound     = errors.New("fatura não encontrada")
	ErrInvoiceAlreadyPaid  = errors.New("fatura já foi paga")
	ErrInvoiceExpired      = errors.New("fatura expirada")
	ErrInvalidAmount       = errors.New("valor inválido para cash-in")
	ErrInvalidPayer        = errors.New("dados do pagador inválidos")
	ErrInvalidProvider     = errors.New("provedor de cash-in inválido")
	ErrProviderRejection   = errors.New("provedor rejeitou a transação")
	ErrInvalidWebhook      = errors.New("webhook inválido")
	ErrInvalidSignature    = errors.New("assinatura do webhook inválida")
	ErrInvoiceCannotBeCancelled = errors.New("fatura não pode ser cancelada neste status")
)
