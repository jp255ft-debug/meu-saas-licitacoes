# Regras Fundamentais do Projeto OrkaPay

## Anti-Alucinação
- **NUNCA invente código ou APIs que não existem.**
- **Sempre verifique a documentação oficial** (Asaas, Stark Bank, Go, NestJS).
- **NUNCA use `any` em TypeScript.** Use tipos explícitos ou `unknown`.

## Arquitetura
- **Domínio é sagrado:** Lógica de negócio NÃO pode ter dependências externas.
- **Event Sourcing é obrigatório:** Toda mudança de estado gera um evento imutável.
- **Adaptadores são plugáveis:** Troca de BaaS = mudança de configuração.

## Commits
- **Commits atômicos:** Um commit por funcionalidade.
- **Mensagens em português:** Descreva o QUE e o PORQUÊ.
