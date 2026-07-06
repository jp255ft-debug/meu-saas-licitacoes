package domain

import "errors"

// Erros sentinela do domínio cashout-engine.
var (
	ErrInvalidTransfer       = errors.New("transferência inválida")
	ErrTransferNotFound      = errors.New("transferência não encontrada")
	ErrTransferAlreadySettled = errors.New("transferência já foi liquidada")
	ErrTransferFailed        = errors.New("transferência falhou")
	ErrInvalidAmount         = errors.New("valor inválido para cash-out")
	ErrInvalidPayee          = errors.New("dados do beneficiário inválidos")
	ErrInvalidBankInfo       = errors.New("informações bancárias inválidas")
	ErrInvalidProvider       = errors.New("provedor de cash-out inválido")
	ErrProviderRejection     = errors.New("provedor rejeitou a transferência")
	ErrInsufficientBalance   = errors.New("saldo insuficiente para transferência")
	ErrTransferCannotBeCancelled = errors.New("transferência não pode ser cancelada neste status")
	ErrInvalidPixKey         = errors.New("chave Pix do beneficiário inválida")
)
