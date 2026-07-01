# AGENTS.md - Guia para IA no Projeto OrkaPay

## Visão Geral
OrkaPay é uma plataforma de orquestração de pagamentos B2B. O projeto utiliza uma arquitetura de microsserviços com Event Sourcing e mensageria assíncrona.

## Estrutura do Projeto
- `.clinerules/` - Regras fundamentais para IA (nunca desative)
- `.memory/` - Contexto persistente para a IA
- `api-gateway/` - Camada externa (NestJS)
- `services/` - Microsserviços Go
- `shared/` - Código compartilhado

## Regras Fundamentais
1. **Domínio é sagrado**: Não coloque dependências externas em `internal/core/domain/`
2. **Event Sourcing obrigatório**: Toda mudança de estado deve gerar um evento
3. **Adaptadores plugáveis**: Troca de BaaS = mudança de configuração, não de código
4. **NUNCA invente APIs**: Consulte a documentação oficial primeiro

## Fluxo de Trabalho Recomendado
1. **DeepSeek (Plan)**: Planeje a arquitetura
2. **Kimi Code (Act)**: Implemente seguindo as regras
3. **Cline (Plan)**: Revise o código antes do commit

## Tecnologias
- Go 1.26+
- NestJS 11+
- Kafka 3.8+
- RabbitMQ 4.0+
- PostgreSQL 16+
- Redis 7+
- Kubernetes 1.32+
