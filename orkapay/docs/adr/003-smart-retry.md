# ADR 003: Smart Retry

## Decisão
Implementar serviço dedicado de retentativa inteligente com ML (DeepSeek).

## Contexto
Cobranças recorrentes podem falhar por saldo insuficiente em horários específicos.

## Motivação
- Recuperar até 85% do churn involuntário
- Analisar padrões de transação para prever melhor janela de retry
- Diferencial competitivo

## Consequências
- Dependência de API externa (DeepSeek)
- Complexidade adicional no agendamento
