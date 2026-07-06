package domain

import "errors"

var (
	ErrAccountNotFound      = errors.New("conta não encontrada")
	ErrInsufficientBalance  = errors.New("saldo insuficiente")
	ErrAccountBlocked       = errors.New("conta bloqueada")
	ErrDuplicateIdempotency = errors.New("chave de idempotência já utilizada")
	ErrInvalidAmount        = errors.New("valor deve ser maior que zero")
)