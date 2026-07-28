"""
Serviço de Smart Match: correlaciona o perfil do usuário com editais disponíveis
para calcular scores de relevância personalizados.
Suporta filtragem por categorias, palavras-chave e ordenação por relevância.
"""
from __future__ import annotations
from datetime import datetime, timezone
from typing import Optional

from pymongo.collection import Collection
from pymongo import DESCENDING, ASCENDING

from backend.app.core.database import mongo_db
from backend.app.models.sql.usuario import Usuario

# Tamanho do lote para processamento de match de usuários
# Evita carregar todos os usuários em memória de uma vez
_MATCH_BATCH_SIZE = 200


def calcular_relevancia(
    usuario: Usuario,
    edital: dict,
) -> float:
    """
    Calcula um score de relevância (0.0 a 1.0) entre um usuário e um edital.
    Leva em conta:
    - Compatibilidade de porte (MEI vs. valor estimado)
    - Palavras-chave do objeto do edital vs. perfil do usuário
    - Prazo de abertura (quanto mais próximo, maior o score)
    - Categorias do edital
    """
    score = 0.5  # score base neutro

    # Fator 1: Porte da empresa vs. valor do edital
    valor_estimado = edital.get("valor_estimado")
    if valor_estimado is not None:
        if usuario.tipo == "MEI" and valor_estimado > 500000:
            score -= 0.2
        elif usuario.tipo == "EMPRESA" and valor_estimado < 10000:
            score -= 0.1

    # Fator 2: Palavras-chave no objeto
    objeto = (edital.get("objeto") or "").lower()
    palavras_usuario = _extrair_palavras_chave(usuario)
    if palavras_usuario:
        match_count = sum(
            1 for p in palavras_usuario if p.lower() in objeto
        )
        score += min(match_count * 0.1, 0.3)

    # Fator 3: Categorias casadas (se houver categorias no edital)
    categorias = edital.get("categorias") or []
    if categorias:
        score += min(len(categorias) * 0.05, 0.15)

    # Fator 4: Data próxima da abertura
    data_abertura = edital.get("data_abertura")
    if data_abertura:
        if isinstance(data_abertura, str):
            try:
                data_abertura = datetime.fromisoformat(
                    data_abertura.replace("Z", "+00:00")
                )
            except (ValueError, TypeError):
                data_abertura = None
        if data_abertura:
            dias_restantes = (data_abertura - datetime.now(timezone.utc)).days
            if 0 < dias_restantes <= 30:
                score += 0.2
            elif dias_restantes <= 0:
                score -= 0.1

    return max(0.0, min(1.0, score))


def buscar_editais_relevantes(
    usuario: Usuario,
    uf: Optional[str] = None,
    modalidade: Optional[str] = None,
    palavra_chave: Optional[str] = None,
    categoria: Optional[str] = None,
    data_inicio: Optional[datetime] = None,
    data_fim: Optional[datetime] = None,
    valor_min: Optional[float] = None,
    valor_max: Optional[float] = None,
    ordenar_por: str = "data_publicacao",
    ordem: str = "desc",
    pagina: int = 1,
    tamanho: int = 20,
) -> tuple[list[dict], int]:
    """
    Busca editais no MongoDB aplicando filtros e calcula score de relevância.
    Retorna (lista_de_editais_com_score, total_de_documentos).
    """
    collection: Collection = mongo_db["editais_pncp"]

    # Monta filtro MongoDB
    filtro: dict = {}

    if uf:
        filtro["uf"] = uf.upper()
    if modalidade:
        filtro["modalidade"] = {"$regex": modalidade, "$options": "i"}
    if palavra_chave:
        # Busca textual no objeto
        filtro["$or"] = [
            {"objeto": {"$regex": palavra_chave, "$options": "i"}},
            {"orgao_nome": {"$regex": palavra_chave, "$options": "i"}},
        ]
    if categoria:
        filtro["categorias"] = categoria.upper()
    if data_inicio or data_fim:
        filtro["data_publicacao"] = {}
        if data_inicio:
            filtro["data_publicacao"]["$gte"] = data_inicio
        if data_fim:
            filtro["data_publicacao"]["$lte"] = data_fim
    if valor_min is not None or valor_max is not None:
        filtro["valor_estimado"] = {}
        if valor_min is not None:
            filtro["valor_estimado"]["$gte"] = valor_min
        if valor_max is not None:
            filtro["valor_estimado"]["$lte"] = valor_max

    # Ordenação
    direcao = DESCENDING if ordem == "desc" else ASCENDING
    sort_field = ordenar_por if ordenar_por in (
        "data_publicacao", "valor_estimado", "data_abertura", "ano"
    ) else "data_publicacao"

    # Total antes da paginação
    total = collection.count_documents(filtro)

    # Paginação
    skip = (pagina - 1) * tamanho
    cursor = (
        collection.find(filtro)
        .sort(sort_field, direcao)
        .skip(skip)
        .limit(tamanho)
    )

    # Processa resultados e calcula score
    resultados = []
    for doc in cursor:
        doc["_id"] = str(doc["_id"])
        doc["score_relevancia"] = calcular_relevancia(usuario, doc)

        # Busca resumo IA se disponível
        proc_collection: Collection = mongo_db["processamento_ia"]
        processamento = proc_collection.find_one(
            {"edital_mongodb_id": doc["_id"]}
        )
        if processamento:
            doc["resumo"] = processamento.get("resumo")
            doc["categorias"] = processamento.get("categorias", [])

        resultados.append(doc)

    return resultados, total


def _extrair_palavras_chave(usuario: Usuario) -> list[str]:
    """
    Extrai palavras-chave do perfil do usuário para matching.
    Usa nome e tipo; idealmente viria de categorias/ramo cadastrado.
    """
    palavras = []
    if usuario.nome:
        palavras.extend(usuario.nome.lower().split())
    palavras.append(usuario.tipo.lower())
    return list(set(palavras))


def usuarios_compatíveis_com_contratacao(
    db: "Session",
    contratacao: dict,
    apenas_assinantes: bool = True,
) -> list[Usuario]:
    """
    Retorna lista de usuários que têm match com uma contratação específica.

    Critérios (avaliados em AND):
    1. UF do órgão está na lista de UFs de interesse do usuário
    2. CNAE do usuário está na lista de CNAEs identificados da contratação
    3. Pelo menos uma palavra-chave do usuário aparece no objeto da contratação

    Otimizado para processamento em lotes (batch) — não carrega todos
    os usuários em memória de uma vez.

    Parâmetros:
        db: Sessão SQLAlchemy.
        contratacao: Dicionário com dados da contratação (formato MongoDB ou API).
        apenas_assinantes: Se True, filtra apenas usuários com assinatura ativa.

    Retorna:
        Lista de objetos Usuario compatíveis.
    """
    from sqlalchemy.orm import Session
    from backend.app.models.sql.assinatura import Assinatura

    # Extrair dados da contratação (aceita camelCase da API ou snake_case do MongoDB)
    orgao = contratacao.get("orgaoEntidade") or contratacao.get("orgao_entidade") or {}
    uf_orgao = orgao.get("uf", "")
    objeto = (contratacao.get("objetoCompra") or contratacao.get("objeto_compra") or "").lower()
    cnaes_contratacao = (
        contratacao.get("cnaesIdentificados")
        or contratacao.get("cnaes_identificados")
        or []
    )

    # Monta query base
    query_base = db.query(Usuario).filter(Usuario.ativo.is_(True))
    if apenas_assinantes:
        query_base = query_base.join(
            Assinatura, Assinatura.usuario_id == Usuario.id
        ).filter(Assinatura.status == "ativa")

    # Conta total para saber quantos lotes processar
    total_usuarios = query_base.count()

    compatíveis = []
    offset = 0

    while offset < total_usuarios:
        # Processa em lotes de _MATCH_BATCH_SIZE
        lote = (
            query_base
            .order_by(Usuario.id)
            .offset(offset)
            .limit(_MATCH_BATCH_SIZE)
            .all()
        )

        for usuario in lote:
            config = usuario.configuracoes_match or {}
            ufs_interesse = config.get("ufs", [])
            cnaes_usuario = config.get("cnaes", [])
            palavras_chave = config.get("palavras_chave", [])

            # 1. Verificar UF
            if ufs_interesse and uf_orgao not in ufs_interesse:
                continue

            # 2. Verificar CNAE (interseção)
            if cnaes_usuario and not (set(cnaes_usuario) & set(cnaes_contratacao)):
                continue

            # 3. Verificar palavras-chave (pelo menos uma aparece no objeto)
            if palavras_chave:
                encontrou = any(palavra.lower() in objeto for palavra in palavras_chave)
                if not encontrou:
                    continue

            compatíveis.append(usuario)

        offset += _MATCH_BATCH_SIZE

    return compatíveis


# ─── Smart Match para PCA ───────────────────────────────────────────────────


def buscar_pca_para_usuario(usuario: Usuario) -> list[dict]:
    """
    Busca itens de PCA compatíveis com o perfil do usuário.
    Usa aggregation pipeline para calcular score de relevância.

    Retorna lista de PCAs ordenados por score total decrescente.
    """
    from datetime import datetime

    config = usuario.configuracoes_match or {}
    cnaes_usuario = config.get("cnaes", [])
    ufs_interesse = config.get("ufs", [])
    palavras_chave = [p.lower() for p in config.get("palavras_chave", [])]

    # Se não tem configuração, retorna PCAs recentes (com baixa relevância)
    if not cnaes_usuario and not ufs_interesse and not palavras_chave:
        return list(
            mongo_db["pca_pncp"]
            .find({"ano": datetime.now().year})
            .sort("data_publicacao", -1)
            .limit(20)
        )

    # Pipeline de agregação
    pipeline = [
        {"$match": {"ano": datetime.now().year}},
        {"$unwind": "$itens"},
        {
            "$addFields": {
                "score": {
                    "$sum": [
                        # +30 se CNAE match
                        {
                            "$cond": [
                                {"$in": ["$itens.cnae_identificado", cnaes_usuario]},
                                30,
                                0,
                            ]
                        },
                        # +20 se UF match
                        {
                            "$cond": [
                                {"$in": ["$orgao_entidade.uf", ufs_interesse]},
                                20,
                                0,
                            ]
                        },
                        # +10 por cada palavra-chave na descrição
                        {
                            "$multiply": [
                                {
                                    "$size": {
                                        "$setIntersection": [
                                            palavras_chave,
                                            {
                                                "$split": [
                                                    {"$toLower": "$itens.descricao_item"},
                                                    " ",
                                                ]
                                            },
                                        ]
                                    }
                                },
                                10,
                            ]
                        },
                        # +10 se valor total > 100k
                        {
                            "$cond": [
                                {"$gt": ["$itens.valor_total_estimado", 100000]},
                                10,
                                0,
                            ]
                        },
                    ]
                }
            }
        },
        {"$match": {"score": {"$gt": 0}}},  # apenas itens com algum score
        {"$sort": {"score": -1}},
        {
            "$group": {
                "_id": "$id_pca",
                "id_pca": {"$first": "$id_pca"},
                "ano": {"$first": "$ano"},
                "orgao_entidade": {"$first": "$orgao_entidade"},
                "itens": {"$push": "$itens"},
                "data_publicacao": {"$first": "$data_publicacao"},
                "score_total": {"$sum": "$score"},
            }
        },
        {"$sort": {"score_total": -1}},
        {"$limit": 50},
    ]

    resultados = list(mongo_db["pca_pncp"].aggregate(pipeline))
    # Converte ObjectId para string
    for r in resultados:
        r["_id"] = str(r["_id"])
    return resultados
