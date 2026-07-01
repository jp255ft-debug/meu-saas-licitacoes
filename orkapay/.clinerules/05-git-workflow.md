# Fluxo de Git - OrkaPay

## Anti-Alucinação em Git
- **NUNCA faça commits gigantes.** Um commit = uma funcionalidade lógica.
- **NUNCA commite arquivos gerados** (node_modules/, dist/, bin/, .env).
- **NUNCA force push em branches compartilhadas** (main, develop).
- **NUNCA commite segredos** (API keys, senhas, tokens).

## Estrutura de Branches

### Branches Principais
- `main` — Produção (protegida, sem pushes diretos)
- `develop` — Desenvolvimento (protegida, sem pushes diretos)

### Branches de Suporte
- `feature/<nome-da-feature>` — Nova funcionalidade (branch de `develop`)
- `fix/<nome-do-fix>` — Correção de bug (branch de `develop`)
- `hotfix/<nome-do-hotfix>` — Correção urgente (branch de `main`)
- `release/<versao>` — Preparação de release (branch de `develop`)

## Convenção de Commits

### Formato (Conventional Commits)
```
<tipo>(<escopo>): <descrição em português>

<corpo opcional explicando o PORQUÊ>
```

### Tipos
- `feat` — Nova funcionalidade
- `fix` — Correção de bug
- `refactor` — Refatoração (sem mudança de comportamento)
- `test` — Adição ou correção de testes
- `docs` — Documentação
- `chore` — Tarefas de manutenção (dependências, config)
- `infra` — Infraestrutura (Docker, K8s, CI/CD)

### Exemplos
```
feat(payment-core): adiciona entidade Payment com value objects

Cria a entidade Payment com CPF/CNPJ como Value Object tipado.
Implementa factory method NewPayment com validações de domínio.
```

```
fix(cashin-engine): corrige validação de webhook do Asaas

O webhook estava aceitando payloads sem assinatura.
Adiciona verificação de HMAC-SHA256 conforme docs do Asaas.
```

## Fluxo de Trabalho
1. Crie branch de `develop`: `git checkout -b feature/nome-da-feature`
2. Faça commits atômicos seguindo a convenção
3. Abra Pull Request para `develop`
4. Code review obrigatório (mínimo 1 aprovador)
5. Merge com squash (manter histórico limpo)
6. Delete branch após merge

## Regras de Pull Request
- **Título:** mesmo formato de commit
- **Descrição:** o que foi feito, porquê, como testar
- **Checklist:** testes passando, lint ok, sem conflitos
- **Tamanho:** máximo 400 linhas de diff (excedeu? divida em PRs menores)
