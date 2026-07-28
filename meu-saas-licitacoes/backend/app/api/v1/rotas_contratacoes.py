"""
Rotas da API para consulta de contratações públicas (PNCP).
Endpoints públicos e autenticados para listar, filtrar e obter
detalhes de contratações com recebimento de propostas aberto.
"""
from math import ceil
from typing import Optional

from bson.objectid import ObjectId
from fastapi import APIRouter, Depends, HTTPException, Query, status
from pymongo import DESCENDING, ASCENDING
from pymongo.collection import Collection

from backend.app.core.database import mongo_db
from backend.app.schemas.contratacao import (
    ContratacaoEstatisticas,
    ContratacaoListaResponse,
    ContratacaoResponse,
    OrgaoEntidadeResponse,
    Paginacao,
)

router = APIRouter(prefix="/contratacoes", tags=["Contratações"])


def _doc_para_response(doc: dict) -> ContratacaoResponse:
    """Converte documento MongoDB para ContratacaoResponse."""
    orgao = doc.get("orgao_entidade")
    orgao_response = None
    if orgao:
        orgao_response = OrgaoEntidadeResponse(
            cnpj=orgao.get("cnpj"),
            nome=orgao.get("nome"),
            sigla=orgao.get("sigla"),
            uf=orgao.get("uf"),
            municipio=orgao.get("municipio"),
        )

    return ContratacaoResponse(
        id=str(doc.get("_id", "")),
        numero_controle_pncp=doc.get("numero_controle_pncp", ""),
        ano_compra=doc.get("ano_compra", 0),
        sequencial_compra=doc.get("sequencial_compra"),
        orgao_entidade=orgao_response,
        modalidade_nome=doc.get("modalidade_nome", ""),
        objeto_compra=doc.get("objeto_compra", ""),
        valor_total_estimado=doc.get("valor_total_estimado"),
        data_publicacao_pncp=doc.get("data_publicacao_pncp"),
        data_abertura_proposta=doc.get("data_abertura_proposta"),
        data_encerramento_proposta=doc.get("data_encerramento_proposta"),
        link_edital=doc.get("link_edital"),
        exclusividade_me_epp=doc.get("exclusividade_me_epp", False),
        cnaes_identificados=doc.get("cnaes_identificados", []),
        data_importacao=doc.get("data_importacao"),
        status_proposta=doc.get("status_proposta", "aberta"),
    )


@router.get("", response_model=ContratacaoListaResponse)
def listar_contratacoes(
    uf: Optional[str] = Query(None, description="Filtrar por UF do órgão"),
    modalidade: Optional[str] = Query(None, description="Filtrar por modalidade de licitação"),
    palavra_chave: Optional[str] = Query(None, description="Busca textual no objeto da compra"),
    exclusividade_me_epp: Optional[bool] = Query(
        None, description="Filtrar apenas exclusivas para ME/EPP"
    ),
    valor_min: Optional[float] = Query(None, description="Valor mínimo estimado"),
    valor_max: Optional[float] = Query(None, description="Valor máximo estimado"),
    ordenar_por: str = Query(
        "data_publicacao_pncp",
        description="Campo de ordenação",
    ),
    ordem: str = Query("desc", description="asc | desc"),
    pagina: int = Query(1, ge=1, description="Número da página"),
    tamanho: int = Query(20, ge=1, le=100, description="Itens por página"),
):
    """
    Lista contratações com recebimento de propostas aberto.
    Endpoint público (não exige autenticação).
    Suporta paginação, filtros combináveis e ordenação.
    """
    collection: Collection = mongo_db["contratacoes_pncp"]

    # Monta filtro MongoDB
    filtro: dict = {"status_proposta": "aberta"}

    if uf:
        filtro["orgao_entidade.uf"] = uf.upper()
    if modalidade:
        filtro["modalidade_nome"] = {"$regex": modalidade, "$options": "i"}
    if palavra_chave:
        filtro["$or"] = [
            {"objeto_compra": {"$regex": palavra_chave, "$options": "i"}},
            {"orgao_entidade.nome": {"$regex": palavra_chave, "$options": "i"}},
        ]
    if exclusividade_me_epp is not None:
        filtro["exclusividade_me_epp"] = exclusividade_me_epp
    if valor_min is not None or valor_max is not None:
        filtro["valor_total_estimado"] = {}
        if valor_min is not None:
            filtro["valor_total_estimado"]["$gte"] = valor_min
        if valor_max is not None:
            filtro["valor_total_estimado"]["$lte"] = valor_max

    # Valida campo de ordenação para evitar injeção
    campos_validos = [
        "data_publicacao_pncp", "data_abertura_proposta",
        "data_encerramento_proposta", "valor_total_estimado",
        "ano_compra",
    ]
    sort_field = ordenar_por if ordenar_por in campos_validos else "data_publicacao_pncp"
    direcao = DESCENDING if ordem == "desc" else ASCENDING

    total = collection.count_documents(filtro)
    skip = (pagina - 1) * tamanho

    cursor = (
        collection.find(filtro)
        .sort(sort_field, direcao)
        .skip(skip)
        .limit(tamanho)
    )

    resultados = [_doc_para_response(doc) for doc in cursor]

    return ContratacaoListaResponse(
        data=resultados,
        paginacao=Paginacao(
            pagina=pagina,
            tamanho=tamanho,
            total=total,
            total_paginas=ceil(total / tamanho) if total > 0 else 0,
        ),
    )


@router.get("/estatisticas", response_model=ContratacaoEstatisticas)
def estatisticas_contratacoes():
    """
    Retorna estatísticas agregadas das contratações abertas.
    Útil para dashboards e indicadores.
    Endpoint público.
    """
    collection: Collection = mongo_db["contratacoes_pncp"]
    pipeline_abertas = [
        {"$match": {"status_proposta": "aberta"}},
        {
            "$group": {
                "_id": None,
                "total": {"$sum": 1},
                "exclusivas_me_epp": {
                    "$sum": {"$cond": ["$exclusividade_me_epp", 1, 0]}
                },
                "valor_total_estimado": {"$sum": {"$ifNull": ["$valor_total_estimado", 0]}},
                "prazo_medio_dias": {"$avg": {
                    "$divide": [
                        {"$subtract": ["$data_encerramento_proposta", "$data_publicacao_pncp"]},
                        1000 * 60 * 60 * 24,  # ms para dias
                    ]
                }},
            }
        },
    ]

    # Contagem por UF
    pipeline_por_uf = [
        {"$match": {"status_proposta": "aberta", "orgao_entidade.uf": {"$ne": None}}},
        {
            "$group": {
                "_id": "$orgao_entidade.uf",
                "count": {"$sum": 1},
            }
        },
        {"$sort": {"count": -1}},
    ]

    aggr_abertas = list(collection.aggregate(pipeline_abertas))
    aggr_por_uf = list(collection.aggregate(pipeline_por_uf))

    stats = aggr_abertas[0] if aggr_abertas else {}
    total_contratacoes = collection.count_documents({})
    por_uf = {item["_id"]: item["count"] for item in aggr_por_uf}

    return ContratacaoEstatisticas(
        total_contratacoes=total_contratacoes,
        total_abertas=stats.get("total", 0),
        exclusivas_me_epp=stats.get("exclusivas_me_epp", 0),
        por_uf=por_uf,
        valor_total_estimado=stats.get("valor_total_estimado", 0.0),
        prazo_medio_propostas_dias=(
            round(stats["prazo_medio_dias"], 1)
            if stats.get("prazo_medio_dias") is not None
            else None
        ),
    )


@router.get("/{contratacao_id}", response_model=ContratacaoResponse)
def detalhe_contratacao(contratacao_id: str):
    """
    Retorna detalhes completos de uma contratação específica.
    """
    collection: Collection = mongo_db["contratacoes_pncp"]

    try:
        doc = collection.find_one({"_id": ObjectId(contratacao_id)})
    except Exception:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="ID da contratação inválido",
        )

    if not doc:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Contratação não encontrada",
        )

    return _doc_para_response(doc)


@router.get("/por-numero/{numero_controle}", response_model=ContratacaoResponse)
def detalhe_contratacao_por_numero(numero_controle: str):
    """
    Retorna detalhes de uma contratação pelo número de controle PNCP.
    Útil para consultar uma contratação específica sabendo seu identificador.
    """
    collection: Collection = mongo_db["contratacoes_pncp"]

    doc = collection.find_one({"numero_controle_pncp": numero_controle})

    if not doc:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Contratação não encontrada com esse número de controle",
        )

    return _doc_para_response(doc)