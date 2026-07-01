# Padrões de Código Go - OrkaPay

## Anti-Alucinação em Go
- **NUNCA invente pacotes da stdlib Go que não existem.** Consulte https://pkg.go.dev/std
- **NUNCA use bibliotecas que não estão no `go.mod`.** Sempre adicione com `go get`.
- **NUNCA invente funções/métodos de APIs externas** (Asaas, Stark Bank, etc.). Consulte a documentação oficial.
- **NUNCA use `interface{}` — use `any`** (Go 1.18+).
- **NUNCA use `panic` para erros de negócio.** Use `error` retornado.

## Estrutura de Código

### Naming Conventions
- **Arquivos:** `snake_case.go` (ex: `payment_created.go`)
- **Pacotes:** `snake_case` (ex: `package payment`)
- **Tipos/Structs:** `PascalCase` (ex: `type PaymentAggregate struct`)
- **Métodos/Funções:** `camelCase` para privadas, `PascalCase` para exportadas
- **Variáveis:** `camelCase` (ex: `paymentID`)
- **Constantes:** `PascalCase` (ex: `const PaymentStatusPending`)

### Organização de Arquivos
```
internal/core/domain/entity/
├── payment.go          # Entidade Payment
├── payment_test.go     # Testes da entidade
├── errors.go           # Erros de domínio
└── value_objects.go    # Value Objects (CPF, CNPJ, Email, etc.)
```

### Regras de Código
- **SEMPRE** retorne `error` como último valor
- **SEMPRE** trate erros com `if err != nil` — nunca ignore
- **NUNCA** use `_` para ignorar erros
- **SEMPRE** feche resources com `defer` (arquivos, conexões, respostas HTTP)
- **NUNCA** use `context.Background()` em produção — use `context.WithTimeout` ou `context.WithCancel`
- **SEMPRE** documente funções exportadas com comentários `// FuncaoX faz Y`

### Testes
- **SEMPRE** escreva testes para entidades de domínio (regras de negócio)
- **Use** `testing` padrão + `testify/assert` para asserts
- **NUNCA** use `reflect.DeepEqual` em testes — use `assert.Equal`
- **Nomeie** testes como `TestEntidade_Cenario_ResultadoEsperado`

### Imports
- Agrupe imports em 3 blocos: stdlib → pacotes internos → pacotes externos
- **NUNCA** use `import . "pacote"` (dot imports)
- **NUNCA** use `import _ "pacote"` sem justificativa documentada
