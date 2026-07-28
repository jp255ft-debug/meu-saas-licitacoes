# 📋 RELATÓRIO DE AUDITORIA TÉCNICA — `meu-saas-licitacoes`

**Data:** 28/07/2026  
**Analista:** IA Cline  
**Versão do Código:** `39c81bd`  
**Tipo:** Auditoria Completa de Software SaaS (Pre-Production)

---

## 📌 1. VISÃO GERAL DO PROJETO

| Atributo | Descrição |
|----------|-----------|
| **Nome** | `meu-saas-licitacoes` |
| **Descrição** | SaaS B2G de curadoria de editais públicos (Lei 14.133/2021) com IA |
| **Público-alvo** | MEIs e pequenas empresas |
| **Arquitetura** | Backend FastAPI + Frontend Next.js + Persistência Poliglota |
| **Maturação** | MVP funcional — pré-produção |
| **Testes** | ❌ Nenhum (diretório `tests/` vazio) |
| **Documentação** | ✅ `CLINE.md` presente; ❌ Swagger incompleto em algumas rotas |

### Stack Implementado

| Camada | Tecnologia | Status |
|--------|------------|--------|
| Backend API | Python 3.11 + FastAPI 0.115 | ✅ Implementado |
| ORM | SQLAlchemy 2.0 (Declarativo) | ✅ Implementado |
| PostgreSQL 16 | Via PgBouncer (pool_mode=transaction) | ✅ Dockerizado |
| MongoDB 7 | PyMongo 4.10 | ✅ Dockerizado |
| Redis 7 | Cache/Filas Celery | ✅ Dockerizado |
| Autenticação | JWT (python-jose) + Bcrypt (passlib) | ✅ Implementado |
| Frontend | Next.js 14 + Tailwind CSS | ✅ Parcial |
| Containerização | Docker Compose | ✅ Configurado |

---

## 🏗️ 2. AUDITORIA DE ARQUITETURA E ESTRUTURA

### 2.1. Organização de Diretórios

| Critério | Status | Observação |
|----------|--------|------------|
| Estrutura segue `CLINE.md` | ✅ | Exata |
| Arquivos órfãos/duplicados | ✅ | Nenhum detectado |
| `__init__.py` presentes | ⚠️ | Faltam em `api/v1/`, `models/sql/`, `models/nosql/`, `schemas/`, `services/`, `worker/`, `utils/` — podem causar problemas de import em alguns contextos |
| Dockerfiles organizados | ✅ | Bem estruturados |

**Problema:** Faltam arquivos `__init__.py` nos subdiretórios de `app/`. Embora o Python 3.3+ tenha namespace packages, a falta explícita pode causar problemas com ferramentas de análise estática e em alguns cenários de import.

### 2.2. Separação de Responsabilidades

| Camada | Conformidade | Observação |
|--------|-------------|------------|
| **API** (`api/`) | ✅ | Rotas HTTP com injeção de dependências via `Depends()` |
| **Dependências** (`dependencias.py`) | ✅ | `get_current_user`, `get_usuario_assinatura_ativa` corretos |
| **Serviços** (`services/`) | ⚠️ | Serviços bem isolados, mas `pagamento.py` mistura lógica de negócio com chamada Stripe |
| **Modelos SQL** (`models/sql/`) | ✅ | Apenas definições de schema com `Mapped[]` |
| **Modelos NoSQL** (`models/nosql/`) | ✅ | Schemas Pydantic para MongoDB |
| **Schemas** (`schemas/`) | ✅ | Pydantic v2 com validações |
| **Worker** (`worker/`) | ✅ | Tasks Celery isoladas |

### 2.3. Persistência Poliglota

| Critério | Status | Detalhes |
|----------|--------|----------|
| Divisão PostgreSQL vs MongoDB | ✅ | Transacional (usuários, assinaturas, certidões) → PostgreSQL; Não-estruturado (editais, contratações, IA) → MongoDB |
| Referência cruzada | ✅ | `usuario.id` (PK auto) ↔ `edital_mongodb_id` (string) — sem FK real |
| Índices no PostgreSQL | ⚠️ | `UUID` e `email` com index; faltam índices compostos para consultas frequentes (ex: `usuario_id + data_validade` em certidões) |
| Índices no MongoDB | ⚠️ | `create_index("numero_controle_pncp", unique=True)` chamado em cada execução do scraper em vez de uma vez na inicialização |

### 2.4. Escalabilidade e Pontos de Gargalo

| Item | Risco | Detalhes |
|------|-------|----------|
| PgBouncer | ✅ | `pool_mode = transaction`, `default_pool_size = 10` — adequado |
| Celery concurrency | ✅ | `--concurrency=4` — razoável para MVP |
| MongoDB índices | 🟠 **Alto** | Consultas frequentes em `editais_pncp` sem índice em `data_publicacao`, `uf`, `modalidade` |
| N+1 queries no SQLAlchemy | 🟢 Baixo | ORM usa `relationship()` mas consultas são diretas via `query()` — baixo risco |
| **N+1 no MongoDB** | 🔴 **Crítico** | `listar_favoritos()` faz 1 query na lista + N queries no MongoDB para cada favorito. `listar_editais()` faz 1 query + N queries no `processamento_ia`. `buscar_novas_contratacoes()` faz 1 query + N queries individuais de verificação |
| **Carregamento em memória** | 🟠 **Alto** | `usuarios_compatíveis_com_contratacao()` carrega **todos os usuários ativos** do banco e itera em Python — inviável com > 1000 usuários |

---

## 🔒 3. AUDITORIA DE SEGURANÇA

### 3.1. Autenticação e Autorização (JWT)

| Item | Status | Detalhes |
|------|--------|----------|
| Senhas hasheadas com bcrypt | ✅ | `passlib[bcrypt]` com salt rounds = 12 (default) |
| JWT com expiração | ⚠️ | 30 minutos, mas **sem refresh token** — usuário é forçado a relogar |
| Rotas protegidas | ✅ | `Depends(get_current_user)` corretamente aplicado |
| Middleware frontend | ✅ | `middleware.ts` protege rotas `/painel/:path*` |
| **Exposição de senhas em logs** | ✅ | Não detectado |
| **Refresh Token** | ❌ **Crítico** | Ausente. Token de 30 min sem renovação causa má experiência e pode levar usuários a armazenarem token inseguramente |

### 3.2. Gestão de Tokens no Frontend

| Item | Severidade | Detalhes |
|------|------------|----------|
| **Token no localStorage** | 🔴 **Crítico** | `useAuth.ts` linha 17: `localStorage.setItem('access_token', access_token)` — vulnerável a XSS. O ideal é usar `httpOnly` cookies |
| Cookie sem `httpOnly` | 🟠 **Alto** | Linha 18: `document.cookie = 'access_token=...'` — acessível via JavaScript |
| Interceptor 401 | ✅ | Redireciona para `/login` ao expirar |

### 3.3. Validação de Entrada

| Item | Status | Detalhes |
|------|--------|----------|
| Inputs validados com Pydantic | ✅ | Todos os schemas usam Pydantic |
| Validação de CPF/CNPJ | ⚠️ | `validators.py` tem funções de validação de dígitos verificadores, mas o schema `UsuarioCreate` **só valida comprimento** (11 ou 14 dígitos) — não usa `validar_cpf()`/`validar_cnpj()` |
| Validação de data | ✅ | `data_validade` não pode estar no passado |
| SQL Injection | ✅ | SQLAlchemy ORM previne (queries parametrizadas) |
| **NoSQL Injection** | 🟠 **Alto** | Uso de `$regex` com input do usuário em `smart_match.py` linha 100 e `rotas_editais.py` linha 79 — embora seja apenas regex textual, a injeção no MongoDB é possível se usuário enviar `$where`, `$ne`, etc. Recomenda-se sanitização adicional |

### 3.4. Segurança de Dados (LGPD)

| Item | Status | Detalhes |
|------|--------|----------|
| Dados sensíveis protegidos | ⚠️ | CPF/CNPJ expostos na resposta da API sem mascaramento |
| Logs de auditoria | ✅ | `HistoricoAcao` implementado |
| Direito ao esquecimento | ❌ **Médio** | `DELETE /me` faz apenas **soft-delete** (`ativo = False`) — dados permanecem no banco |
| Consentimento explícito | ❌ **Médio** | Não há registro de consentimento para coleta/uso de dados |
| Criptografia em trânsito | ⚠️ | Docker expõe HTTP direto; HTTPS depende de reverse proxy (não configurado no docker-compose) |
| Criptografia em repouso | ❌ | Não detectada — PostgreSQL e MongoDB sem criptografia de dados |

### 3.5. Variáveis de Ambiente e Segredos

| Item | Status | Detalhes |
|------|--------|----------|
| `.env` no `.gitignore` | ✅ | Presente |
| `.env.local` do frontend | ⚠️ | Não listado no `.gitignore` — risco se contiver secrets |
| **SECRET_KEY hardcoded** | 🔴 **Crítico** | `config.py` linha 30: `SECRET_KEY: str = "supersecretkeychangeinproduction"` — valor padrão inseguro |
| **STRIPE_SECRET_KEY placeholder** | 🟠 **Alto** | `sk_test_placeholder` — se não for sobrescrito no `.env`, aceitará pagamentos sem processamento real |

### 3.6. CORS e Headers de Segurança

| Item | Status | Detalhes |
|------|--------|----------|
| **CORS** | 🔴 **Crítico** | `allow_origins=["*"]` — permite qualquer origem em produção |
| HSTS | ❌ | Não configurado |
| X-Frame-Options | ❌ | Não configurado |
| CSP (Content-Security-Policy) | ❌ | Não configurado |
| Rate Limiting | ❌ | Nenhum endpoint tem rate limiting — risk de brute force e DoS |

---

## ⚡ 4. AUDITORIA DE PERFORMANCE

### 4.1. Consultas ao Banco de Dados

| Problema | Arquivo | Severidade |
|----------|---------|------------|
| **N+1 queries**: `listar_favoritos()` busca N editais no MongoDB | `rotas_editais.py:353-366` | 🟠 **Alto** |
| **N+1 queries**: `listar_editais()` busca N processamentos IA | `rotas_editais.py:131-135` | 🟠 **Alto** |
| **N+1 queries**: `buscar_novas_contratacoes()` verifica existência individual | `scraper_contratacoes.py:240-244` | 🟡 **Médio** |
| **Sem projeção MongoDB**: `collection.find(filtro)` sem `{}` de projeção em várias queries | `rotas_contratacoes.py`, `rotas_editais.py` | 🟡 **Médio** |
| **CREATE INDEX repetido**: `salvar_contratacoes_mongodb` chama `create_index` em cada inserção | `scraper_contratacoes.py:143` | 🟢 **Baixo** |

### 4.2. Processamento Assíncrono (Celery)

| Item | Status | Detalhes |
|------|--------|----------|
| Tarefas pesadas delegadas | ✅ | Scraping, IA, alertas como tasks Celery |
| Rate limiting PNCP | ❌ | Sem controle de taxa para chamadas à API do PNCP — risco de bloqueio |
| Timeout configurado | ✅ | `task_soft_time_limit=300`, `task_time_limit=600` |
| Retry | ❌ | Tasks não têm `retry`/`max_retries` configurados |

### 4.3. Cache

| Item | Status | Detalhes |
|------|--------|----------|
| Cache de dados frequentes | ❌ | UFs, modalidades e listas sem cache — Redis usado apenas como broker Celery |
| Redis além de filas | ❌ | Não há uso de cache Redis para acelerar consultas repetitivas |

### 4.4. Frontend (Next.js)

| Item | Status | Detalhes |
|------|--------|----------|
| Imagens otimizadas | ⚠️ | Não há uso de `<Image>` do Next.js |
| Code splitting | ✅ | App Router já faz nativamente |
| React Query com `staleTime` | ❌ | `useEditais.ts` não define `staleTime` — refetch a cada mount |
| Loading/Error states | ⚠️ | Hooks retornam dados sem tratamento de loading/error nos componentes |

---

## 🐛 5. AUDITORIA DE CÓDIGO E QUALIDADE

### 5.1. Python (Backend)

| Item | Status | Detalhes |
|------|--------|----------|
| PEP 8 | ✅ | Nomes `snake_case`, indentação 4 espaços |
| Funções longas | ⚠️ | `verificar_novas_contratacoes()` em `tarefas_cron.py` tem ~110 linhas — precisa refatorar |
| DRY | ⚠️ | Formatação de mensagens (WhatsApp e e-mail) tem lógica duplicada para extrair dados da contratação |
| Type hints | ✅ | Presentes em todas as funções |
| Tratamento de exceções | ⚠️ | `except Exception` genérico em vários lugares — perde contexto do erro |
| **`import uuid` dentro de funções** | 🟢 **Baixo** | `tarefas_cron.py` linhas 273, 301 — import dentro do loop |
| **`ObjectId` importado dentro de funções** | 🟢 **Baixo** | `rotas_editais.py` linhas 212, 268, 352 |
| **`resumir_edital` usa string como `_id`** | 🟡 **Médio** | Linha 27: `find_one({"_id": edital_id})` — precisa converter para `ObjectId(edital_id)` |

### 5.2. TypeScript/React (Frontend)

| Item | Status | Detalhes |
|------|--------|----------|
| Componentes funcionais | ✅ | Todos com hooks |
| Tipos definidos | ✅ | Interfaces para props e estados |
| Hooks encapsulados | ✅ | `useAuth`, `useEditais` bem estruturados |
| `useMemo`/`useCallback` | ❌ | Não utilizados — pode causar re-renders desnecessários |
| Loading/Error states | ⚠️ | `useEditais.ts` não expõe `isLoading`/`error` para os componentes | 
| `staleTime` no React Query | ❌ | Não configurado — chamadas repetitivas à API |

### 5.3. Erros Específicos Encontrados

| # | Arquivo | Linha | Problema | Severidade |
|---|---------|-------|----------|------------|
| 1 | `integracao_whatsapp.py` | 62-63 | `WHATSAPP_API_URL` e `WHATSAPP_TOKEN` são constantes hardcoded, ignoram `settings.WHATSAPP_API_KEY` | 🔴 **Crítico** |
| 2 | `integracao_openai.py` | 27 | `find_one({"_id": edital_id})` — `edital_id` é string, deveria ser `ObjectId(edital_id)` | 🟡 **Médio** |
| 3 | `rotas_editais.py` | 165-170 | Bloco morto: `try: data_inicio_dt = None / except: pass` | 🟢 **Baixo** |
| 4 | `scraper_contratacoes.py` | 143 | `create_index` chamado em cada execução upsert | 🟢 **Baixo** |
| 5 | `tarefas_cron.py` | 147-148 | Alertas WhatsApp para certidões com `# TODO` — não implementados | 🟡 **Médio** |

### 5.4. Testes

| Item | Status | Detalhes |
|------|--------|----------|
| Testes unitários | ❌ | Nenhum |
| Testes de integração | ❌ | Nenhum |
| Testes E2E | ❌ | Nenhum |
| Cobertura | ❌ | 0% |
| Diretório `tests/` | ❌ | Vazio |

### 5.5. Documentação

| Item | Status | Detalhes |
|------|--------|----------|
| `CLINE.md` atualizado | ✅ | Completo |
| Docstrings | ✅ | Presentes nas funções principais |
| Swagger `/docs` | ⚠️ | Gerado automaticamente pelo FastAPI, mas algumas rotas têm descrições genéricas |
| `README.md` | ❌ | Não encontrado na raiz do projeto |

---

## 🔗 6. AUDITORIA DE INTEGRAÇÕES EXTERNAS

### 6.1. API PNCP

| Item | Status | Detalhes |
|------|--------|----------|
| Endpoints corretos v3 | ⚠️ | Usa `api/consulta/v1/` — verificar se é v1 ou v3 |
| Rate limiting/retry | ❌ | Sem tratamento — chamadas concorrentes podem levar a bloqueio |
| Validação de dados | ✅ | `_mapear_dados_contratacao` e `salvar_editais_mongodb` validam antes de salvar |
| Fallback v1 | ❌ | Não há fallback |

### 6.2. Stripe (Pagamentos)

| Item | Status | Detalhes |
|------|--------|----------|
| Webhooks configurados | ✅ | Endpoint `/webhook` implementado |
| Verificação de assinatura | ✅ | `stripe.Webhook.construct_event()` com `STRIPE_WEBHOOK_SECRET` |
| Fluxo de cancelamento | ✅ | Cancelamento no Stripe + atualização local |
| Sincronização de status | ⚠️ | Apenas eventos `completed`, `deleted`, `payment_failed` tratados — faltam `subscription.updated`, `invoice.paid` |
| **Dados de pagamento no servidor** | ✅ | Stripe.js/Checkout mantém dados no Stripe |

### 6.3. WhatsApp (Evolution API)

| Item | Status | Detalhes |
|------|--------|----------|
| URL configurável via `.env` | ⚠️ | `settings.WHATSAPP_API_URL` existe mas **não é usado** no módulo — constante hardcoded `WHATSAPP_API_URL = "https://api.whatsapp.com/send"` |
| Retry/fallback | ❌ | Sem retry em falhas de envio |
| Validação de telefone | ✅ | `validators.py` tem `validar_telefone()` |

### 6.4. SendGrid (E-mail)

| Item | Status | Detalhes |
|------|--------|----------|
| Chave via `.env` | ✅ | `settings.SENDGRID_API_KEY` |
| Template de e-mail | ❌ | Apenas texto plano (`plain_text_content`) sem templates HTML |
| Tratamento de falhas | ✅ | Try/except com log |

---

## 🐳 7. AUDITORIA DE INFRAESTRUTURA E DEPLOY

### 7.1. Docker e Docker Compose

| Item | Status | Detalhes |
|------|--------|----------|
| Serviços sobem corretamente | ✅ | 6 serviços configurados (postgres, pgbouncer, mongodb, redis, backend, celery-worker) |
| Healthchecks | ✅ | Todos os serviços de infra têm healthcheck |
| Volumes persistentes | ✅ | `postgres_data`, `mongo_data`, `redis_data` |
| Variáveis de ambiente | ✅ | Passadas via `env_file: .env` |
| **Celery Worker sem healthcheck** | 🟢 **Baixo** | Não tem healthcheck — pode estar morto sem detecção |

### 7.2. Segurança da Infraestrutura

| Item | Status | Detalhes |
|------|--------|----------|
| PostgreSQL senha forte | ⚠️ | `licitacao_pass` — senha simples, default |
| MongoDB autenticação | ⚠️ | `MONGO_INITDB_ROOT_USERNAME/PASSWORD` configurados, mas não há confirmação de `--auth` no comando |
| Redis com senha | ✅ | `redis-server --requirepass redis_pass` |
| Portas expostas | ⚠️ | Portas do PostgreSQL (5432) e MongoDB (27017) expostas externamente — deveriam ser internas apenas |
| **MongoDB sem `--auth` explícito** | 🟡 **Médio** | A imagem `mongo:7` por padrão NÃO ativa autenticação sem `--auth` no comando |

### 7.3. Monitoramento e Logs

| Item | Status | Detalhes |
|------|--------|----------|
| Logs estruturados | ❌ | `logging` padrão do Python, sem formato JSON |
| Ferramentas de monitoramento | ❌ | Sem Sentry, Prometheus, ou similar |
| Logs centralizados | ❌ | Apenas stdout dos containers |

---

## ✅ 8. CHECKLIST DE CONFORMIDADE

### 8.1. Lei 14.133/2021 (Nova Lei de Licitações)

| Item | Status | Detalhes |
|------|--------|----------|
| Dados PNCP conforme legislação | ✅ | Scraper usa API pública |
| Exclusividade ME/EPP | ✅ | Campo `exclusividade_me_epp` presente e filtrável |
| Prazos e modalidades extraídos | ✅ | `data_abertura_proposta`, `data_encerramento_proposta`, `modalidade_nome` |
| Tratamento de dados públicos | ✅ | Dados via API pública governamental |

### 8.2. LGPD (Lei Geral de Proteção de Dados)

| Item | Status | Detalhes |
|------|--------|----------|
| Política de privacidade | ❌ | Não encontrada |
| Usuário pode acessar/corrigir dados | ✅ | Endpoints `GET /me` e `PUT /me` |
| Usuário pode excluir dados | ❌ | Apenas soft-delete (`ativo = False`) |
| Registro de consentimento | ❌ | Não há tabela ou campo de consentimento |
| Armazenamento por tempo necessário | ❌ | Sem política de retenção definida |

### 8.3. Stripe (PCI-DSS)

| Item | Status | Detalhes |
|------|--------|----------|
| Dados de pagamento não tocam servidor | ✅ | Stripe Checkout |
| Webhooks com verificação | ✅ | `stripe-signature` verificado |
| Logging de eventos | ✅ | Eventos registrados/processados |

---

## 📊 9. RELATÓRIO FINAL

### 9.1. Score de Saúde do Projeto: **54/100**

| Categoria | Peso | Score | Justificativa |
|-----------|------|-------|---------------|
| **Arquitetura** | 15% | **75** | Estrutura sólida e bem organizada. Separação de responsabilidades correta. Débitos: `__init__.py` faltantes, índices MongoDB ausentes |
| **Segurança** | 25% | **30** | 🔴 **Crítico**: CORS aberto, SECRET_KEY hardcoded, token no localStorage, sem refresh token, WhatsApp token ignorado, sem rate limiting, sem headers de segurança |
| **Performance** | 15% | **45** | Múltiplos N+1 no MongoDB, carregamento em memória de todos usuários, sem cache, sem projeções |
| **Qualidade de Código** | 20% | **55** | Funções longas, imports dentro de loops, validação CPF/CNPJ incompleta, bloco morto, **zero testes** |
| **Integrações** | 10% | **50** | WhatsApp não funcional (hardcoded), OpenAI sem ObjectId, PNCP sem rate limit, SendGrid sem template HTML |
| **Infraestrutura** | 10% | **65** | Docker bem configurado, healthchecks ok, mas MongoDB sem `--auth`, portas expostas, sem monitoramento |
| **Conformidade** | 5% | **35** | LGPD: sem exclusão real, sem consentimento, sem política de privacidade |
| **Total** | **100%** | **54** | 🟠 **Risco Alto** — necessita correções críticas antes de produção |

### 9.2. Top 5 Itens Mais Urgentes

| # | Item | Categoria | Severidade | Solução Proposta |
|---|------|-----------|------------|------------------|
| 1 | **CORS `allow_origins=["*"]`** + **SECRET_KEY hardcoded** + **Token JWT no localStorage** | Segurança | 🔴 Crítico | **CORS**: Configurar `allow_origins` com origem do frontend. **SECRET_KEY**: Mover para `.env` com valor forte (32+ chars). **Token**: Migrar para cookie `httpOnly` + `SameSite=Strict` com refresh token |
| 2 | **WhatsApp não funcional**: `WHATSAPP_API_URL` e `WHATSAPP_TOKEN` são constantes hardcoded que ignoram `settings.WHATSAPP_API_KEY` | Integrações | 🔴 Crítico | Substituir constantes por `settings.WHATSAPP_API_URL` e `settings.WHATSAPP_API_KEY`. Remover constantes soltas. Testar integração |
| 3 | **N+1 queries MongoDB** em `listar_favoritos()`, `listar_editais()`, `buscar_novas_contratacoes()` + **carregamento de todos usuários em memória** | Performance | 🟠 Alto | **Favoritos**: Fazer join via aggregation pipeline (`$lookup`) em vez de N queries. **Editais**: Usar `$lookup` com `processamento_ia`. **Usuários**: Adicionar filtro no SQL + paginação |
| 4 | **Zero testes automatizados** + **Sem validação real de CPF/CNPJ** + **`resumir_edital()` sem ObjectId** | Qualidade | 🟡 Médio | **Testes**: Implementar testes unitários para serviços core (auth, scraper, pagamento). **CPF/CNPJ**: Usar `validar_cpf()`/`validar_cnpj()` de `validators.py` nos schemas. **ObjectId**: Corrigir `find_one({"_id": edital_id})` para `ObjectId(edital_id)` |
| 5 | **MongoDB sem `--auth`** + **Portas expostas** + **Sem rate limiting** + **LGPD sem exclusão real** | Infra/Conformidade | 🟡 Médio | **MongoDB**: Adicionar `--auth` ao comando. **Portas**: Remover `ports` expostos de postgres/mongodb (apenas rede interna). **Rate limit**: Adicionar `slowapi` ou middleware custom. **LGPD**: Implementar hard-delete real com script de anonimização |

### 9.3. Recomendações Estratégicas

#### 🔴 Curto Prazo (1-2 semanas) — Prioridade Absoluta
1. **Corrigir vulnerabilidades críticas de segurança**: CORS, SECRET_KEY, token JWT no localStorage
2. **Corrigir integração WhatsApp**: Substituir constantes hardcoded por settings
3. **Adicionar refresh token** com `httpOnly` cookie
4. **Implementar rate limiting** nos endpoints de autenticação (evitar brute force)
5. **Corrigir MongoDB sem `--auth`** e portas expostas

#### 🟡 Médio Prazo (1-2 meses)
1. **Resolver N+1 queries no MongoDB** com aggregation pipelines
2. **Adicionar índices MongoDB** para consultas frequentes (`data_publicacao`, `uf`, `modalidade`)
3. **Implementar testes unitários** para serviços core (mínimo 30% de cobertura)
4. **Adicionar validação real de CPF/CNPJ** usando `validators.py`
5. **Corrigir `resumir_edital()`** para usar `ObjectId()`
6. **Adicionar cache Redis** para listas de UFs, modalidades, estatísticas
7. **Implementar LGPD**: hard-delete com anonimização + registro de consentimento
8. **Adicionar `__init__.py`** nos subdiretórios de `app/`

#### 🟢 Longo Prazo (3+ meses)
1. **Monitoramento**: Integrar Sentry para erros + Prometheus/Grafana para métricas
2. **CI/CD**: Adicionar pipeline de testes com GitHub Actions
3. **Logs estruturados**: Migrar para formato JSON com `structlog` ou `python-json-logger`
4. **Retry automático**: Configurar retry com backoff para tasks Celery e chamadas PNCP
5. **Templates de e-mail HTML**: Melhorar experiência do SendGrid com templates
6. **Otimização de frontend**: Adicionar `staleTime` no React Query, `useMemo`, imagens otimizadas
7. **Documentação**: Criar `README.md` completo, melhorar descrições no Swagger
8. **Cobertura de testes > 70%**: Testes de integração e E2E

---

## 📝 NOTAS FINAIS

- **O projeto tem uma base sólida**: arquitetura limpa, separação de responsabilidades correta, escolhas tecnológicas adequadas para o domínio
- **Risco principal é segurança**: os 3 itens críticos (CORS, SECRET_KEY, JWT no localStorage) precisam ser corrigidos **antes de qualquer deploy em produção**
- **Integração WhatsApp não funcional**: a configuração hardcoded impede o funcionamento do principal canal de alerta do sistema
- **Performance aceitável para MVP**, mas os N+1 no MongoDB degradarão rapidamente com crescimento de dados
- **Zero testes** é o maior débito técnico — cada nova feature aumenta o risco de regressão

---

*Relatório gerado automaticamente por IA Cline em 28/07/2026.*