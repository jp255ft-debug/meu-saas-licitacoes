package domain

import "errors"

// Erros sentinela do domínio payment-core.
// Usados em validações e regras de negócio.
// Nunca retornam erros de infraestrutura (DB, HTTP, etc) —
// esses são mapeados para estes erros de domínio nos adaptadores.
var (
	// Money
	ErrNegativeAmount = errors.New("valor monetário não pode ser negativo")
	ErrInvalidFormat  = errors.New("formato de valor monetário inválido")

	// PixKey
	ErrInvalidPixKey     = errors.New("chave Pix inválida")
	ErrEmptyPixKey       = errors.New("chave Pix não pode ser vazia")
	ErrUnsupportedPixKey = errors.New("tipo de chave Pix não suportado")

	// Payment (Aggregate)
	ErrPaymentAlreadyProcessed = errors.New("pagamento já foi processado")
	ErrPaymentNotFound         = errors.New("pagamento não encontrado")
	ErrInvalidPaymentStatus    = errors.New("status de pagamento inválido para esta operação")
	ErrPaymentExpired          = errors.New("pagamento expirado")
	ErrPaymentAmountMismatch   = errors.New("valor do pagamento não corresponde ao esperado")
	ErrPaymentNotPending       = errors.New("pagamento não está pendente")
	ErrPaymentAlreadySettled   = errors.New("pagamento já foi liquidado")
	ErrPaymentAlreadyFailed    = errors.New("pagamento já falhou")
	ErrPaymentAlreadyCancelled = errors.New("pagamento já foi cancelado")
	ErrPaymentCannotBeCancelled = errors.New("pagamento não pode ser cancelado neste status")
	ErrInvalidPayer            = errors.New("dados do pagador inválidos")
	ErrInvalidPayee            = errors.New("dados do beneficiário inválidos")
	ErrMissingGatewayConfig    = errors.New("configuração do gateway de pagamento ausente")
	ErrGatewayRejection        = errors.New("gateway de pagamento rejeitou a transação")
)
