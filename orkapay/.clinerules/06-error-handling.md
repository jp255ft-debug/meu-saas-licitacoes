# Tratamento de Erros - OrkaPay

## Anti-Alucinação em Erros
- **NUNCA use `panic` para erros de negócio.** Use `error` retornado.
- **NUNCA engula erros** com `_` sem justificativa explícita.
- **NUNCA exponha detalhes internos em mensagens de erro para o cliente.**
- **NUNCA logue dados sensíveis** (CPF, senhas, tokens, chaves de API).

## Hierarquia de Erros

### 1. Erros de Domínio (internal/core/domain/entity/errors.go)
```go
var (
    ErrInvalidCPF          = errors.New("cpf inválido")
    ErrInvalidAmount       = errors.New("valor deve ser maior que zero")
    ErrPaymentNotFound     = errors.New("pagamento não encontrado")
    ErrPaymentAlreadyPaid  = errors.New("pagamento já foi pago")
    ErrInsufficientBalance = errors.New("saldo insuficiente")
)
```
- **Regra:** erros de domínio são sentinelas (`var Err... = errors.New(...)`)
- **Regra:** NUNCA adicione contexto variável em erros de domínio (use wrapping)

### 2. Erros de Aplicação (internal/service/)
```go
// Use fmt.Errorf com %w para wrapping
if err := repo.Save(ctx, payment); err != nil {
    return fmt.Errorf("falha ao salvar pagamento %s: %w", payment.ID, err)
}
```

### 3. Erros de Adaptador (internal/adapter/)
- **Traduza** erros de bibliotecas externas para erros de domínio
- **NUNCA** propague erros de banco de dados ou HTTP diretamente para o domínio

## Regras de Log

### Níveis
- `ERROR` — Erro que impede operação (banco fora, Kafka indisponível)
- `WARN` — Erro recuperável (retry de mensagem, webhook inválido)
- `INFO` — Eventos de negócio (pagamento criado, boleto emitido)
- `DEBUG` — Informação detalhada para desenvolvimento

### O que logar
- ID da transação/correlação
- Tipo de evento
- Resultado da operação (sucesso/falha)
- Tempo de execução

### O que NUNCA logar
- Senhas, tokens, chaves de API
- Dados completos de cartão de crédito
- CPF/CNPJ completos (mascarar: `***.***.***-**`)
- Payloads completos de webhook (logar apenas metadados)

## Respostas de Erro (API Gateway)

### Formato Padrão
```json
{
    "success": false,
    "error": {
        "code": "PAYMENT_NOT_FOUND",
        "message": "Pagamento não encontrado",
        "requestId": "req-abc123"
    }
}
```

### Códigos HTTP
- `400` — Erro de validação (cliente mandou dados inválidos)
- `404` — Recurso não encontrado
- `409` — Conflito (pagamento já processado)
- `422` — Erro de negócio (saldo insuficiente)
- `500` — Erro interno (não exponha detalhes)

## Smart Retry
- Erros de gateway (timeout, 5xx) → retentativa com backoff exponencial
- Erros de domínio (422, 409) → NUNCA retentar (erro lógico)
- Erros de validação (400) → NUNCA retentar (erro do cliente)
