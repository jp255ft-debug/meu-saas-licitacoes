# 🧠 PROJETO: LICITAÇÃO SAAS - MVP
**Diretório**: `C:\Source\Repos\meu-saas-licitacoes`

---

## 📌 1. OBJETIVO GERAL
Construir um SaaS B2G de curadoria de editais públicos (Lei 14.133/2021) com inteligência artificial.
- **Diferencial**: Resumir editais de 80 páginas em 5 temas usando IA (OpenAI).
- **Público-alvo**: MEIs e pequenas empresas que querem vender para o governo.

---

## ⚙️ 2. STACK TECNOLÓGICA (Regra Fixa)
| Camada | Tecnologia | Versão |
|--------|------------|--------|
| **Backend API** | Python + FastAPI | 3.11+ / 0.115+ |
| **ORM Relacional** | SQLAlchemy | 2.0+ (Declarativo) |
| **Banco Relacional** | PostgreSQL (via PgBouncer) | 16+ |
| **Banco NoSQL** | MongoDB (PyMongo) | 7.0+ |
| **Cache/Filas** | Redis (Celery) | 7+ |
| **Autenticação** | JWT (python-jose) + Bcrypt (passlib) | - |
| **Validação** | Pydantic v2 | - |
| **Frontend** | Next.js (App Router) + Tailwind | 14+ |

---

## 🚫 3. REGRAS DE NEGÓCIO E RESTRIÇÕES
1. **NUNCA** use `async/await` para SQLAlchemy a menos que explicitamente autorizado (para evitar conflitos com PgBouncer, use driver síncrono `psycopg2-binary`).
2. **Todas as senhas** devem ser hashadas com `bcrypt` (salt rounds = 12).
3. **UUIDs** devem ser gerados com `uuid.uuid4()` e expostos na API, NUNCA o ID auto-incrementado.
4. **Datas** sempre em UTC (`datetime.utcnow()` ou `func.now()` no SQLAlchemy).
5. **Evite joins** MongoDB/PostgreSQL; a referência entre PostgreSQL (usuario.id) e MongoDB (edital._id) é feita via string no campo `edital_mongodb_id`.

---

## 📂 4. ESTRUTURA DE DIRETÓRIOS (OBRIGATÓRIA)
```plaintext
backend/
├── app/
│   ├── api/               # Rotas HTTP
│   │   ├── v1/            # Versão 1 da API
│   │   └── dependencias.py # get_current_user, get_db, etc
│   ├── core/              # Configurações e segurança
│   │   ├── config.py      # Carrega .env (pydantic-settings)
│   │   ├── database.py    # Sessões SQLAlchemy + conexão MongoDB
│   │   └── security.py    # JWT e hashing
│   ├── models/            # SQLAlchemy (PostgreSQL) e Schemas (Mongo)
│   │   ├── sql/           # Tabelas: usuario, assinatura, certidao, etc
│   │   └── nosql/         # Coleções: edital_pncp, processamento_ia
│   ├── schemas/           # Pydantic (Request/Response)
│   ├── services/          # Lógica de negócio (scraping, IA, whatsapp)
│   ├── worker/            # Celery tasks (scraping, alertas)
│   ├── utils/             # Funções auxiliares (validadores, formatação)
│   └── main.py            # Entrypoint FastAPI
├── migrations/            # Alembic (controle de versão do banco)
├── docker/                # Configs para Postgres, PgBouncer, MongoDB
├── .env.example           # Template de variáveis
└── requirements.txt       # Dependências Python
```

---

## 🔗 5. ORDEM DE CONSTRUÇÃO (File-by-File)
Siga esta ordem estritamente para evitar erros de importação circular:

1. **Core**: `config.py` -> `database.py` -> `security.py`
2. **Models (SQL)**: `usuario.py` -> `assinatura.py` -> `certidao.py` -> `edital_favorito.py` -> `historico_acao.py`
3. **Models (NoSQL)**: `edital_pncp.py` -> `processamento_ia.py`
4. **Schemas (Pydantic)**: `usuario.py`, `token.py`, `certidao.py`
5. **API Dependencies**: `dependencias.py` (injeta `get_db`, `get_current_user`)
6. **API Routes**: `rotas_usuarios.py` -> `rotas_auth.py`
7. **Main**: `app/main.py` (conecta tudo)

---

## 🎯 6. CONVENÇÕES DE CÓDIGO
- **SQLAlchemy**: Use `Mapped[]` e `mapped_column()` (SQLAlchemy 2.0 style).
- **Pydantic**: Use `model_validator` para validações customizadas (ex: CPF/CNPJ).
- **Nomes**: `snake_case` para variáveis/funções; `PascalCase` para classes.
- **Docstrings**: Sempre que uma função tiver lógica complexa, adicione um comentário de "Por quê?".
- **Erros**: Use `raise HTTPException(status_code=..., detail=...)` do FastAPI.

---

## 🔧 7. AMBIENTE (Variáveis de Ambiente)
O arquivo `.env` deve conter:
```
# PostgreSQL (via PgBouncer)
DATABASE_URL=postgresql://licitacao_user:licitacao_pass@pgbouncer:6432/licitacao_db
# MongoDB
MONGODB_URL=mongodb://admin:admin_pass@mongodb:27017/licitacao_db?authSource=admin
# Redis
REDIS_URL=redis://:redis_pass@redis:6379/0
# JWT
SECRET_KEY=supersecretkeychangeinproduction
ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=30
```

---

## 🧪 8. TESTES MANUAIS (Swagger)
Após cada bloco de rotas implementado, valide via Swagger:
- `http://localhost:8000/docs`

---

## 📌 9. INSTRUÇÃO PARA O CLINE (Agente IA)
> **IMPORTANTE**: Antes de sugerir qualquer código, verifique se os arquivos mencionados já existem na estrutura. Se não existirem, CRIE-OS com o conteúdo adequado. Não invente caminhos diferentes do especificado acima. Mantenha a compatibilidade com a versão do Python e bibliotecas listadas.