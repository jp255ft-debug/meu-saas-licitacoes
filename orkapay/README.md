# OrkaPay

**OrkaPay** é uma plataforma de orquestração de pagamentos B2B que conecta o mundo analógico do boleto ao digital do Pix. Oferecemos uma suíte financeira completa com cashback e participação nos lucros para nossos clientes.

## 🚀 Características

- Orquestração de Pix, Pix Automático e Boletos
- Boleto Híbrido (com QR Code Pix)
- Conciliação Automática com Ledger Imutável
- Cashback e Revenue Share para clientes
- Integração com ERPs e BaaS (Asaas, Stark Bank, Celcoin, Fitbank)
- Arquitetura de microsserviços em Go com Event Sourcing

## 📚 Documentação

- [API Documentation](./docs/api/openapi.yaml)
- [Architecture Decision Records](./docs/adr/)
- [Pricing & Plans](./docs/pricing/)
- [Business Model](./docs/business/)

## 🛠️ Tecnologias

- **Core**: Go (Golang) + NestJS (TypeScript)
- **Mensageria**: Apache Kafka + RabbitMQ
- **Banco de Dados**: PostgreSQL + Redis
- **Event Sourcing**: Kafka (Event Store) + PostgreSQL (Projeções)
- **Orquestração**: Kubernetes + Helm
- **IA**: DeepSeek (Smart Retry)

## 🏃 Como Executar

```bash
# Subir infraestrutura local
make up

# Executar migrações
make migrate

# Build dos serviços
make build

# Executar testes
make test
```

## 📄 Licença

MIT
