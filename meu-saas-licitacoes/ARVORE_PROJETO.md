# 🌳 Árvore Completa do Projeto — `meu-saas-licitacoes`

```
meu-saas-licitacoes/
│
├── .gitignore
├── AUDITORIA_RELATORIO.md             # Relatório completo de auditoria técnica
├── ARVORE_PROJETO.md                  # Este arquivo — árvore do projeto
├── CLINE.md                           # Documentação do projeto (CLINE)
│
├── backend/                           # API Backend (FastAPI + Celery)
│   ├── .env.example                   # Template de variáveis de ambiente
│   ├── alembic.ini                    # Configuração do Alembic (migrações)
│   ├── docker-compose.yml             # Orquestração Docker (6 serviços)
│   ├── Dockerfile                     # Dockerfile do backend
│   ├── requirements.txt               # Dependências Python
│   │
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py                    # Entrypoint FastAPI (CORS, rotas)
│   │   │
│   │   ├── api/
│   │   │   ├── dependencias.py        # get_current_user, get_usuario_assinatura_ativa
│   │   │   └── v1/                    # Rotas da API versão 1
│   │   │       ├── rotas_assinaturas.py   # Stripe checkout + webhook
│   │   │       ├── rotas_auth.py          # Login JWT + logout (cookie httpOnly)
│   │   │       ├── rotas_certidoes.py     # CRUD certidões
│   │   │       ├── rotas_contratacoes.py  # Contratações PNCP abertas
│   │   │       ├── rotas_editais.py       # Editais + favoritos + Kanban
│   │   │       ├── rotas_pca.py           # ★ PCA (Plano de Contratações Anual)
│   │   │       └── rotas_usuarios.py      # CRUD usuários + preferências
│   │   │
│   │   ├── core/
│   │   │   ├── config.py              # Settings (pydantic-settings)
│   │   │   ├── database.py            # SQLAlchemy + MongoDB + índices
│   │   │   └── security.py            # JWT (create/decode) + bcrypt hash
│   │   │
│   │   ├── models/
│   │   │   ├── sql/                   # Modelos PostgreSQL (SQLAlchemy)
│   │   │   │   ├── assinatura.py      # Assinaturas (planos, gateway)
│   │   │   │   ├── certidao.py        # Certidões negativas/regularidade
│   │   │   │   ├── edital_favorito.py # Editais favoritados + Kanban
│   │   │   │   ├── historico_acao.py  # Auditoria de ações do usuário
│   │   │   │   ├── notificacao.py     # Notificações enviadas (WhatsApp, e-mail, in-app)
│   │   │   │   └── usuario.py         # Tabela central de usuários
│   │   │   └── nosql/                 # Schemas MongoDB (Pydantic)
│   │   │       ├── contratacao_pncp.py    # Contratações com propostas abertas
│   │   │       ├── edital_pncp.py         # Editais raspados do PNCP
│   │   │       ├── pca_pncp.py            # ★ PCA (itens de contratação anual)
│   │   │       └── processamento_ia.py    # Resumos de IA + embeddings
│   │   │
│   │   ├── schemas/                   # Schemas Pydantic (Request/Response)
│   │   │   ├── assinatura.py          # Planos + checkout + response
│   │   │   ├── certidao.py            # Certidão create/update/response
│   │   │   ├── contratacao.py         # Contratação response + estatísticas
│   │   │   ├── edital.py              # Edital + favoritos + paginação
│   │   │   ├── token.py               # LoginRequest + TokenResponse
│   │   │   └── usuario.py             # Usuario create/update/response
│   │   │
│   │   ├── services/                  # Lógica de negócio
│   │   │   ├── email.py               # SendGrid — envio de alertas
│   │   │   ├── integracao_openai.py   # Resumo de editais com GPT
│   │   │   ├── integracao_whatsapp.py # Evolution API — WhatsApp
│   │   │   ├── pagamento.py           # Stripe — checkout + webhooks
│   │   │   ├── scraper_contratacoes.py # Scraper de contratações PNCP
│   │   │   ├── scraper_pca.py         # ★ Scraper do PCA (ano atual + anterior)
│   │   │   ├── scraper_pncp.py        # Scraper de editais PNCP
│   │   │   └── smart_match.py         # Match perfil ↔ editais/PCA (batch otimizado)
│   │   │
│   │   ├── utils/
│   │   │   └── validators.py          # Validação CPF, CNPJ, telefone
│   │   │
│   │   └── worker/                    # Tarefas assíncronas (Celery)
│   │       ├── celery_app.py          # Config Celery + schedules
│   │       └── tarefas_cron.py        # Scraping, IA, certidões, PCA
│   │
│   ├── docker/                        # Configurações Docker
│   │   ├── mongodb/
│   │   │   └── init.js                # Script de inicialização MongoDB
│   │   ├── pgbouncer/
│   │   │   ├── pgbouncer.ini          # Pool mode: transaction
│   │   │   └── userlist.txt           # Lista de usuários PgBouncer
│   │   └── postgres/
│   │       └── init.sql               # Script de inicialização PostgreSQL
│   │
│   ├── migrations/                    # Migrações Alembic
│   │   └── versions/                  # (vazio — migrações futuras)
│   │
│   └── tests/                         # Testes automatizados
│                                       # (vazio — a implementar)
│
└── frontend/                          # Frontend (Next.js 14 + Tailwind)
    ├── .env.local                     # NEXT_PUBLIC_API_URL
    ├── package.json                   # Dependências Next.js
    ├── package-lock.json
    ├── tailwind.config.ts             # Configuração do Tailwind CSS
    │
    ├── public/                        # Assets públicos
    │                                   # (vazio)
    │
    └── src/
        ├── middleware.ts              # Middleware Next.js (proteção de rotas)
        │
        ├── app/                       # App Router
        │   ├── globals.css            # Estilos globais Tailwind
        │   ├── layout.tsx             # Root layout (Providers)
        │   ├── page.tsx               # Landing Page
        │   ├── providers.tsx          # QueryClientProvider
        │   │
        │   ├── (auth)/                # Rotas de autenticação
        │   │   ├── login/
        │   │   │   └── page.tsx       # Página de login (JSON body, cookie httpOnly)
        │   │   └── cadastro/
        │   │       └── page.tsx       # Página de cadastro (cria + login automático)
        │   │
        │   └── painel/                # Dashboard (autenticado)
        │       ├── layout.tsx         # Sidebar + menu + logout
        │       ├── page.tsx           # ★ Dashboard (cards de estatísticas)
        │       │
        │       ├── editais/
        │       │   └── page.tsx       # ★ Lista de editais com filtros
        │       ├── favoritos/
        │       │   └── page.tsx       # ★ Kanban (PARA_LER → ARQUIVADO)
        │       ├── pca/
        │       │   └── page.tsx       # ★ PCA (recomendados + listagem)
        │       ├── checkout/
        │       │   └── page.tsx       # ★ Planos (Starter R$97 / Pro R$197)
        │       └── configuracoes/
        │           └── page.tsx       # Configurações (Alertas, Smart Match, Histórico)
        │
        ├── components/
        │   ├── ConfiguracoesAlerta.tsx     # Preferências de notificação
        │   ├── ConfiguracoesSmartMatch.tsx # Perfil de matching (UF, CNAE, palavras)
        │   ├── HistoricoNotificacoes.tsx   # Histórico de notificações enviadas
        │   └── ui/                        # Componentes UI (shadcn)
        │       ├── badge.tsx
        │       ├── button.tsx
        │       ├── card.tsx
        │       ├── input.tsx
        │       ├── label.tsx
        │       ├── select.tsx
        │       ├── switch.tsx
        │       ├── table.tsx
        │       └── tabs.tsx
        │
        ├── hooks/
        │   ├── useAuth.ts             # Login/logout (sem localStorage)
        │   └── useEditais.ts          # Queries de editais (React Query)
        │
        └── lib/
            ├── api.ts                 # Axios instance (withCredentials)
            └── utils.ts               # Utilitários (cn)
```

## 📊 Resumo

| Categoria | Total |
|-----------|-------|
| Arquivos backend (Python) | ~40 |
| Arquivos frontend (TSX/TS) | ~25 |
| Arquivos de configuração/infra | ~10 |
| **Total aproximado** | **~75 arquivos** |

### Stack
- **Backend:** Python 3.11, FastAPI, SQLAlchemy 2.0, PostgreSQL 16, MongoDB 7, Redis 7, Celery 5
- **Frontend:** Next.js 14, Tailwind CSS, React Query, Axios
- **Infra:** Docker Compose, PgBouncer, Nginx (produção)

### Legenda
- **★** = Arquivo criado durante o sprint de correções
- **(vazio)** = Diretório existente aguardando conteúdo