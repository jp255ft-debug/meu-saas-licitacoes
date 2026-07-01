# Políticas de Teste - OrkaPay

## Anti-Alucinação em Testes
- **NUNCA invente frameworks de teste que não estão no projeto.**
- **NUNCA escreva testes que dependem de ambiente externo sem mock explícito.**
- **NUNCA use `it.only` ou `describe.only` em commits.**
- **NUNCA pule testes com `.skip` sem justificativa documentada.**

## Pirâmide de Testes

### 1. Testes Unitários (70% do código)
- **Obrigatórios para:** entidades de domínio, value objects, serviços
- **Proibido:** chamadas reais a banco, APIs externas, Kafka
- **Ferramentas:** Go `testing` + `testify`, NestJS `@nestjs/testing` + `jest`

### 2. Testes de Integração (20% do código)
- **Obrigatórios para:** adaptadores (PostgreSQL, Kafka, gateways)
- **Use:** testcontainers para banco de dados, mock servers para APIs
- **NUNCA:** use banco de dados de produção ou staging

### 3. Testes End-to-End (10% do código)
- **Obrigatórios para:** fluxos críticos (criação de pagamento, conciliação)
- **Use:** ambiente isolado com docker-compose

## Regras Obrigatórias

### Cobertura
- **Domínio:** 100% de cobertura (entidades, regras de negócio)
- **Serviços:** 90%+ de cobertura
- **Adaptadores:** 80%+ de cobertura
- **Controllers:** 70%+ de cobertura

### Nomenclatura
- **Go:** `TestEntidade_Cenario_ResultadoEsperado`
- **NestJS:** `describe('PaymentService')` / `it('deve criar pagamento com sucesso')`

### O que NÃO testar
- Código gerado (ORM entities, migrations)
- Configurações
- Código de terceiros (bibliotecas)

### Checklist de Teste para cada Feature
- [ ] Teste unitário da entidade/agregado
- [ ] Teste unitário do use case (serviço)
- [ ] Teste de integração do adaptador
- [ ] Teste de contrato (se aplicável)
