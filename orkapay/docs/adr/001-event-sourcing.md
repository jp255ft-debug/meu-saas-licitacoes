# ADR 001: Event Sourcing

## Decisão
Adotar Event Sourcing como padrão para persistência do estado financeiro.

## Contexto
Sistemas financeiros exigem rastreabilidade total e auditoria imutável.

## Motivação
- Auditoria completa e imutável
- Reconstrução de estado em qualquer ponto no tempo
- Suporte nativo a padrões de conciliação

## Consequências
- Curva de aprendizado mais íngreme
- Maior complexidade de infraestrutura (Kafka)
