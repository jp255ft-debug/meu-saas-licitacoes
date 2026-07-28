"""
Rotas da API para consulta do Plano de Contratações Anual (PCA) do PNCP.
Endpoints públicos e autenticados para listar, filtrar e obter
detalhes de PCAs com score de relevância personalizado.
"""
from math import ceil
from datetime import datetime
from typing import Optional

from bson.objectid import ObjectId
from fastapi import APIRouter, Depends, HTTPException, Query, status
from pymongo import DESCENDING
from pymongo.collection import Collection

from backend.app.core.database import mongo_db
from backend.app.api.dependencias import get_current_user
from backend.app.models.sql.usuario import Usuario
from backend.app.services.smart_match import buscar_pca_para_usuario

router = APIRouter(prefix="/pca", tags=["Plano de Contratações Anual"])


@router.get("")
def listar_pca(
    ano: int = Query(default=datetime.now().year, description="Ano de referência"),
    uf: Optional[str] = Query(None, description="Filtrar por UF do órgão"),
    orgao: Optional[str] = Query(None, description="Busca textual no nome do órgão"),
    pagina: int = Query(1, ge=1, description="Número da página"),
    tamanho: int = Query(20, ge=1, le=100, description="Itens por página"),
):
    """
    Lista Planos de Contratação Anual por ano e UF.
    Endpoint público (não exige autenticação).
    """
    collection: Collection = mongo_db["pca_pncp"]

    # Monta filtro
    filtro: dict = {"ano": ano}
    if uf:
        filtro["orgao_entidade.uf"] = uf.upper()
    if orgao:
        filtro["orgao_entidade.nome"] = {"$regex": orgao, "$options": "i"}

    total = collection.count_documents(filtro)
    skip = (pagina - 1) * tamanho

    cursor = (
        collection.find(filtro)
        .sort("data_publicacao", DESCENDING)
        .skip(skip)
        .limit(tamanho)
    )

    resultados = []
    for doc in cursor:
        doc["_id"] = str(doc["_id"])
        resultados.append(doc)

    return {
        "items": resultados,
        "total": total,
        "pagina": pagina,
        "tamanho": tamanho,
        "total_paginas": ceil(total / tamanho) if total > 0 else 0,
    }


@router.get("/meus-pca")
def listar_meus_pca(
    usuario: Usuario = Depends(get_current_user),
):
    """
    Lista itens de PCA compatíveis com o perfil do usuário autenticado.
    Usa o Smart Match para ranquear por relevância (CNAE, UF, palavras-chave).
    Exige token JWT.
    """
    resultados = buscar_pca_para_usuario(usuario)
    return {
        "items": resultados,
        "total": len(resultados),
    }


@router.get("/estatisticas")
def estatisticas_pca():
    """
    Retorna estatísticas agregadas dos PCAs disponíveis.
    Útil para dashboard.
    Endpoint público.
    """
    collection: Collection = mongo_db["pca_pncp"]
    ano_atual = datetime.now().year

    # Total de PCAs por ano
    total_ano_atual = collection.count_documents({"ano": ano_atual})
    total_ano_anterior = collection.count_documents({"ano": ano_atual - 1})

    # Itens por UF (aggregation)
    pipeline_por_uf = [
        {"$match": {"ano": ano_atual}},
        {"$unwind": "$itens"},
        {
            "$group": {
                "_id": "$orgao_entidade.uf",
                "total_itens": {"$sum": 1},
                "valor_total": {"$sum": {"$ifNull": ["$itens.valor_total_estimado", 0]}},
            }
        },
        {"$sort": {"total_itens": -1}},
    ]

    por_uf = list(collection.aggregate(pipeline_por_uf))
    uf_result = {
        item["_id"]: {"itens": item["total_itens"], "valor": item["valor_total"]}
        for item in por_uf if item["_id"]
    }

    return {
        "ano_atual": ano_atual,
        "total_pcas_ano_atual": total_ano_atual,
        "total_pcas_ano_anterior": total_ano_anterior,
        "por_uf": uf_result,
    }


@router.get("/{id_pca}")
def detalhe_pca(id_pca: str):
    """
    Retorna detalhes completos de um PCA específico pelo ID.
    """
    collection: Collection = mongo_db["pca_pncp"]

    doc = collection.find_one({"id_pca": id_pca})
    if not doc:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="PCA não encontrado",
        )

    doc["_id"] = str(doc["_id"])
    return doc