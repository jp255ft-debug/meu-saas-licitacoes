from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from backend.app.core.config import settings
from backend.app.api.v1 import (
    rotas_usuarios,
    rotas_auth,
    rotas_editais,
    rotas_certidoes,
    rotas_assinaturas,
    rotas_contratacoes,
    rotas_pca,
)

app = FastAPI(
    title="Licitação SaaS - API",
    description="Curadoria de editais públicos com IA (Lei 14.133/2021)",
    version="1.0.0",
)

# CORS — origens configuradas via settings (nunca "*" em produção)
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.CORS_ORIGINS,
    allow_credentials=True,  # necessário para cookies httpOnly
    allow_methods=["*"],
    allow_headers=["*"],
)

# Rotas
app.include_router(rotas_usuarios.router, prefix="/api/v1")
app.include_router(rotas_auth.router, prefix="/api/v1")
app.include_router(rotas_editais.router, prefix="/api/v1")
app.include_router(rotas_certidoes.router, prefix="/api/v1")
app.include_router(rotas_assinaturas.router, prefix="/api/v1")
app.include_router(rotas_contratacoes.router, prefix="/api/v1")
app.include_router(rotas_pca.router, prefix="/api/v1")


@app.get("/health")
def health_check():
    """Endpoint de saúde para verificar se a API está rodando."""
    return {"status": "ok", "version": "1.0.0"}