# Regras de Arquitetura - OrkaPay

## Anti-Alucinação em Arquitetura
- **NUNCA invente padrões arquiteturais que não existem no projeto.**
- **Sempre siga a Clean/Hexagonal Architecture:** domínio isolado, adaptadores nas bordas.
- **NUNCA coloque lógica de negócio em adaptadores ou infraestrutura.**
- **NUNCA crie dependências de frameworks no domínio** (ex: importar NestJS ou GORM em `internal/core/domain/`).

## Estrutura Obrigatória

### Camadas (do centro para fora)
1. **Domain** (`internal/core/domain/`) — Entidades, Value Objects, Eventos de domínio
   - ⚠️ **Proibido:** importar qualquer pacote externo, banco de dados, framework web
   - ✅ **Permitido:** apenas tipos nativos Go e erros customizados
2. **Ports** (`internal/core/port/`) — Interfaces (inbound/outbound)
   - Inbound: use cases / serviços de aplicação
   - Outbound: repositories, event stores, gateways, messaging
3. **Service** (`internal/service/`) — Implementação dos use cases
   - Depende apenas de interfaces (ports), nunca de implementações concretas
4. **Adapter** (`internal/adapter/`) — Implementações concretas
   - Inbound: REST handlers, gRPC, consumers (Kafka/RabbitMQ)
   - Outbound: PostgreSQL, Event Store, Gateways (Asaas, Stark Bank), Messaging

### Regras de Dependência
- **SEMPRE** aponte do externo para o interno (adaptador → port → domínio)
- **NUNCA** o domínio pode depender de adaptadores
- **NUNCA** crie dependência cíclica entre camadas

## Event Sourcing
- **Toda mudança de estado DEVE gerar um evento imutável**
- Eventos são a fonte da verdade (source of truth)
- Projeções/read models são derivados dos eventos
- **NUNCA** atualize estado diretamente no banco sem gerar evento

## BaaS Providers
- **Cada provedor (Asaas, Stark Bank, Celcoin, Fitbank) é um adaptador**
- Trocar de provedor = mudar configuração, NUNCA código de domínio
- **NUNCA** invente endpoints de API — consulte documentação oficial
