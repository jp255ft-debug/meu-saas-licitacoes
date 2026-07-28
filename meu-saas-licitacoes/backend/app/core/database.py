"""
Configuração das conexões com PostgreSQL (SQLAlchemy) e MongoDB.
Índices do MongoDB são criados automaticamente na importação deste módulo.
"""
import logging

from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, declarative_base
from pymongo import MongoClient
from pymongo.collection import Collection

from backend.app.core.config import settings

logger = logging.getLogger(__name__)

# ─── SQLAlchemy (PostgreSQL) ────────────────────────────────────────────────
# Driver síncrono psycopg2-binary para compatibilidade com PgBouncer
engine = create_engine(
    settings.DATABASE_URL,
    pool_pre_ping=True,       # verifica conexão antes de usar
    pool_size=10,             # conexões no pool
    max_overflow=20,          # conexões extras se o pool estourar
)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()


def get_db():
    """
    Injeta uma sessão SQLAlchemy nas rotas FastAPI.
    Uso: `def minha_rota(db: Session = Depends(get_db))`
    """
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


# ─── MongoDB ────────────────────────────────────────────────────────────────
mongo_client = MongoClient(settings.MONGODB_URL)
mongo_db = mongo_client.get_database("licitacao_db")


def get_mongo_db():
    """
    Injeta a instância do banco MongoDB nas rotas/services.
    """
    return mongo_db


# ─── Índices MongoDB (criados uma vez na inicialização) ─────────────────────

def criar_indices_mongodb():
    """
    Cria todos os índices necessários para o MongoDB.
    Chamado automaticamente na importação deste módulo.
    Separado em função para facilitar testes e recriação manual.
    """
    colecoes_indices = {
        "contratacoes_pncp": [
            ("numero_controle_pncp", {"unique": True}),
            ([("orgao_entidade.uf", 1), ("status_proposta", 1)], {}),
            ([("cnaes_identificados", 1), ("status_proposta", 1)], {}),
            ([("status_proposta", 1), ("data_publicacao_pncp", -1)], {}),
        ],
        "editais_pncp": [
            ("pncp_id", {"unique": True}),
            ("uf",),
            ([("uf", 1), ("data_publicacao", -1)], {}),
        ],
        "processamento_ia": [
            ("edital_mongodb_id", {"unique": True}),
            ("categorias",),
        ],
        "pca_pncp": [
            ("id_pca", {"unique": True}),
            ("ano",),
            ([("ano", -1), ("orgao_entidade.uf", 1)], {}),
            ([("orgao_entidade.cnpj", 1)], {}),
        ],
    }

    for colecao_nome, indices in colecoes_indices.items():
        collection: Collection = mongo_db[colecao_nome]
        for idx_def in indices:
            try:
                if isinstance(idx_def, str):
                    # Campo único: create_index("campo", unique=True)
                    collection.create_index(idx_def, unique=True)
                elif isinstance(idx_def, tuple):
                    campo = idx_def[0]
                    kwargs = idx_def[1] if len(idx_def) > 1 else {}
                    if isinstance(campo, str):
                        collection.create_index(campo, **kwargs)
                    elif isinstance(campo, list):
                        collection.create_index(campo, **kwargs)
            except Exception as e:
                logger.warning(f"Erro ao criar índice {idx_def} em {colecao_nome}: {e}")

    logger.info("✅ Índices MongoDB verificados/criados")


# Cria índices automaticamente na importação
criar_indices_mongodb()