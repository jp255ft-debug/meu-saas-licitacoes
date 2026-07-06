package domain

import "errors"

// Erros sentinela do domínio boleto-engine.
var (
	// Boleto
	ErrInvalidBoletoNumber = errors.New("número do boleto inválido")
	ErrInvalidBarcode      = errors.New("código de barras inválido")
	ErrInvalidDigitable    = errors.New("linha digitável inválida")
	ErrBoletoNotFound      = errors.New("boleto não encontrado")
	ErrBoletoAlreadyPaid   = errors.New("boleto já foi pago")
	ErrBoletoExpired       = errors.New("boleto expirado")
	ErrBoletoCannotBeCancelled = errors.New("boleto não pode ser cancelado neste status")
	ErrInvalidDueDate      = errors.New("data de vencimento inválida")
	ErrInvalidFine         = errors.New("valor de multa inválido")
	ErrInvalidInterest     = errors.New("valor de juros inválido")
	ErrInvalidDiscount     = errors.New("valor de desconto inválido")
	ErrInvalidPayerInfo    = errors.New("informações do pagador inválidas")
	ErrInvalidProvider     = errors.New("provedor de boleto inválido")
	ErrProviderRejection   = errors.New("provedor rejeitou a emissão do boleto")
)
