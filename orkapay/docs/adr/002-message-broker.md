# ADR 002: Message Broker

## Decisão
Usar Kafka para Event Sourcing e RabbitMQ para filas transacionais.

## Contexto
Sistema precisa processar alto volume de eventos e garantir entrega confiável para retentativas.

## Motivação
- Kafka: alto throughput, imutabilidade, replay
- RabbitMQ: filas de atraso, TTL, DLX para Smart Retry

## Consequências
- Duas infraestruturas de mensageria para gerenciar
- Maior robustez e resiliência
