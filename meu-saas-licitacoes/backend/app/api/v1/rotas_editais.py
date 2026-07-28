from math import ceil
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query, status
from pymongo.collection import Collection
from sqlalchemy.orm import Session

from backend.app.api.dependencias import get_current_user
from backend.app.core.database import get_db, get_mongo_db, mongo_db
from backend.app.models.sql.edital_favorito import EditalFavorito
from backend.app.models.sql.usuario import Usuario
from backend.app.schemas.edital import (
    EditalFiltrar,
    EditalListaResponse,
    EditalResponse,
    FavoritoCreate,
    FavoritoResponse,
    FavoritoUpdate,
    Paginacao,
)
from backend.app.services.smart_match import (
    buscar_editais_relevantes,
    calcular_relevancia,
)

router = APIRouter(prefix="/editais", tags=["Editais"])


def _doc_para_response(doc: dict) -> EditalResponse:
    """Converte documento MongoDB para EditalResponse."""
    return EditalResponse(
        id=doc.get("_id", ""),
        pncp_id=doc.get("pncp_id", ""),
        orgao_nome=doc.get("orgao_nome", ""),
        orgao_cnpj=doc.get("orgao_cnpj", ""),
        uf=doc.get("uf", ""),
        cidade=doc.get("cidade"),
        modalidade=doc.get("modalidade", ""),
        numero=doc.get("numero", ""),
        ano=doc.get("ano", 0),
        objeto=doc.get("objeto", ""),
        valor_estimado=doc.get("valor_estimado"),
        data_publicacao=doc.get("data_publicacao"),
        data_abertura=doc.get("data_abertura"),
        url_pncp=doc.get("url_pncp"),
        documento_url=doc.get("documento_url"),
        resumo=doc.get("resumo"),
        categorias=doc.get("categorias", []),
        score_relevancia=doc.get("score_relevancia"),
    )


@router.get("", response_model=EditalListaResponse)
def listar_editais(
    uf: Optional[str] = Query(None, description="Filtrar por UF"),
    modalidade: Optional[str] = Query(None, description="Filtrar por modalidade"),
    palavra_chave: Optional[str] = Query(None, description="Busca textual no objeto/órgão"),
    categoria: Optional[str] = Query(None, description="Filtrar por categoria (OBRAS, SAUDE, TI, etc.)"),
    data_inicio: Optional[str] = Query(None, description="Data início (ISO)"),
    data_fim: Optional[str] = Query(None, description="Data fim (ISO)"),
    valor_min: Optional[float] = Query(None, description="Valor mínimo estimado"),
    valor_max: Optional[float] = Query(None, description="Valor máximo estimado"),
    ordenar_por: str = Query("data_publicacao", description="Campo de ordenação"),
    ordem: str = Query("desc", description="asc | desc"),
    pagina: int = Query(1, ge=1, description="Número da página"),
    tamanho: int = Query(20, ge=1, le=100, description="Itens por página"),
):
    """
    Lista editais públicos com paginação e filtros combináveis.
    Não exige autenticação.
    """
    collection: Collection = mongo_db["editais_pncp"]

    # Monta filtro MongoDB
    filtro: dict = {}
    if uf:
        filtro["uf"] = uf.upper()
    if modalidade:
        filtro["modalidade"] = {"$regex": modalidade, "$options": "i"}
    if palavra_chave:
        filtro["$or"] = [
            {"objeto": {"$regex": palavra_chave, "$options": "i"}},
            {"orgao_nome": {"$regex": palavra_chave, "$options": "i"}},
        ]
    if categoria:
        # Busca na coleção processamento_ia as categorias e cruza
        proc_collection: Collection = mongo_db["processamento_ia"]
        proc_ids = [
            p["edital_mongodb_id"]
            for p in proc_collection.find(
                {"categorias": categoria.upper()}, {"edital_mongodb_id": 1}
            )
        ]
        if proc_ids:
            filtro["_id"] = {"$in": proc_ids}
        else:
            # Sem resultados, retorna vazio
            return EditalListaResponse(
                data=[], paginacao=Paginacao(pagina=pagina, tamanho=tamanho, total=0, total_paginas=0)
            )
    if data_inicio or data_fim:
        filtro["data_publicacao"] = {}
        if data_inicio:
            from datetime import datetime
            filtro["data_publicacao"]["$gte"] = datetime.fromisoformat(data_inicio)
        if data_fim:
            from datetime import datetime
            filtro["data_publicacao"]["$lte"] = datetime.fromisoformat(data_fim)
    if valor_min is not None or valor_max is not None:
        filtro["valor_estimado"] = {}
        if valor_min is not None:
            filtro["valor_estimado"]["$gte"] = valor_min
        if valor_max is not None:
            filtro["valor_estimado"]["$lte"] = valor_max

    from pymongo import DESCENDING, ASCENDING
    direcao = DESCENDING if ordem == "desc" else ASCENDING
    sort_field = ordenar_por if ordenar_por in (
        "data_publicacao", "valor_estimado", "data_abertura", "ano"
    ) else "data_publicacao"

    total = collection.count_documents(filtro)
    skip = (pagina - 1) * tamanho

    cursor = collection.find(filtro).sort(sort_field, direcao).skip(skip).limit(tamanho)

    # Coleta todos os IDs para bulk query ao processamento_ia
    docs_list = list(cursor)
    if docs_list:
        # Busca resumos IA em lote (1 query em vez de N)
        proc_collection: Collection = mongo_db["processamento_ia"]
        ids_str = []
        for doc in docs_list:
            doc["_id"] = str(doc["_id"])
            ids_str.append(doc["_id"])

        processados = list(
            proc_collection.find(
                {"edital_mongodb_id": {"$in": ids_str}},
                {"edital_mongodb_id": 1, "resumo": 1, "categorias": 1},
            )
        )
        proc_dict = {p["edital_mongodb_id"]: p for p in processados}

        for doc in docs_list:
            proc = proc_dict.get(doc["_id"])
            if proc:
                doc["resumo"] = proc.get("resumo")
                doc["categorias"] = proc.get("categorias", [])

    resultados = [_doc_para_response(doc) for doc in docs_list]

    return EditalListaResponse(
        data=resultados,
        paginacao=Paginacao(
            pagina=pagina,
            tamanho=tamanho,
            total=total,
            total_paginas=ceil(total / tamanho) if total > 0 else 0,
        ),
    )


@router.get("/relevantes", response_model=EditalListaResponse)
def listar_editais_relevantes(
    uf: Optional[str] = Query(None, description="Filtrar por UF"),
    modalidade: Optional[str] = Query(None, description="Filtrar por modalidade"),
    palavra_chave: Optional[str] = Query(None, description="Busca textual"),
    categoria: Optional[str] = Query(None, description="Filtrar por categoria"),
    ordenar_por: str = Query("score_relevancia", description="Campo de ordenação"),
    ordem: str = Query("desc", description="asc | desc"),
    pagina: int = Query(1, ge=1),
    tamanho: int = Query(20, ge=1, le=100),
    usuario: Usuario = Depends(get_current_user),
):
    """
    Lista editais ranqueados por score de relevância para o usuário autenticado.
    Exige token JWT.
    """
    try:
        data_inicio_dt = None
        data_fim_dt = None
    except Exception:
        data_inicio_dt = None
        data_fim_dt = None

    resultados, total = buscar_editais_relevantes(
        usuario=usuario,
        uf=uf,
        modalidade=modalidade,
        palavra_chave=palavra_chave,
        categoria=categoria,
        ordenar_por=ordenar_por,
        ordem=ordem,
        pagina=pagina,
        tamanho=tamanho,
    )

    # Se ordenar por score, reordena em memória
    if ordenar_por == "score_relevancia":
        resultados.sort(
            key=lambda r: (r.get("score_relevancia") or 0),
            reverse=(ordem == "desc"),
        )

    itens = [_doc_para_response(doc) for doc in resultados]

    return EditalListaResponse(
        data=itens,
        paginacao=Paginacao(
            pagina=pagina,
            tamanho=tamanho,
            total=total,
            total_paginas=ceil(total / tamanho) if total > 0 else 0,
        ),
    )


@router.get("/{edital_id}", response_model=EditalResponse)
def detalhe_edital(
    edital_id: str,
):
    """
    Retorna detalhes completos de um edital, incluindo resumo IA se disponível.
    """
    collection: Collection = mongo_db["editais_pncp"]
    from bson.objectid import ObjectId

    try:
        doc = collection.find_one({"_id": ObjectId(edital_id)})
    except Exception:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="ID do edital inválido",
        )

    if not doc:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Edital não encontrado",
        )

    doc["_id"] = str(doc["_id"])

    # Busca resumo IA
    proc_collection: Collection = mongo_db["processamento_ia"]
    proc = proc_collection.find_one({"edital_mongodb_id": doc["_id"]})
    if proc:
        doc["resumo"] = proc.get("resumo")
        doc["categorias"] = proc.get("categorias", [])

    return _doc_para_response(doc)


@router.post("/{edital_id}/favoritar", response_model=FavoritoResponse, status_code=201)
def favoritar_edital(
    edital_id: str,
    payload: Optional[FavoritoCreate] = None,
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """
    Marca um edital como favorito para o usuário autenticado.
    O edital_id é o _id do MongoDB.
    """
    # Verifica se já favoritou
    existente = (
        db.query(EditalFavorito)
        .filter(
            EditalFavorito.usuario_id == usuario.id,
            EditalFavorito.edital_mongodb_id == edital_id,
        )
        .first()
    )
    if existente:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Edital já está nos favoritos",
        )

    # Verifica se o edital existe no MongoDB
    collection: Collection = mongo_db["editais_pncp"]
    from bson.objectid import ObjectId

    try:
        doc = collection.find_one({"_id": ObjectId(edital_id)})
    except Exception:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="ID do edital inválido",
        )

    if not doc:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Edital não encontrado",
        )

    notas = payload.notas if payload else None

    favorito = EditalFavorito(
        usuario_id=usuario.id,
        edital_mongodb_id=edital_id,
        status_kanban="PARA_LER",
        notas=notas,
    )
    db.add(favorito)
    db.commit()
    db.refresh(favorito)

    return FavoritoResponse(
        uuid=favorito.uuid,
        edital_mongodb_id=favorito.edital_mongodb_id,
        status_kanban=favorito.status_kanban,
        notas=favorito.notas,
        data_criacao=favorito.data_criacao,
    )


@router.delete("/{edital_id}/favoritar", status_code=204)
def desfavoritar_edital(
    edital_id: str,
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """
    Remove um edital dos favoritos.
    """
    favorito = (
        db.query(EditalFavorito)
        .filter(
            EditalFavorito.usuario_id == usuario.id,
            EditalFavorito.edital_mongodb_id == edital_id,
        )
        .first()
    )
    if not favorito:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Favorito não encontrado",
        )

    db.delete(favorito)
    db.commit()


@router.get("/favoritos", response_model=list[FavoritoResponse])
def listar_favoritos(
    status_kanban: Optional[str] = Query(None, description="Filtrar por status (PARA_LER, LENDO, ANALISADO, ARQUIVADO)"),
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """
    Lista os editais favoritos do usuário autenticado.
    Opcionalmente filtra por coluna do Kanban.

    Otimizado: usa bulk query com $in no MongoDB em vez de N queries individuais.
    """
    query = db.query(EditalFavorito).filter(EditalFavorito.usuario_id == usuario.id)
    if status_kanban:
        query = query.filter(EditalFavorito.status_kanban == status_kanban.upper())
    query = query.order_by(EditalFavorito.data_criacao.desc())

    favoritos = query.all()
    if not favoritos:
        return []

    # Extrai todos os IDs do MongoDB para busca em lote
    from bson.objectid import ObjectId
    mongo_ids = []
    for fav in favoritos:
        try:
            mongo_ids.append(ObjectId(fav.edital_mongodb_id))
        except Exception:
            pass

    # 1 query bulk com $in em vez de N queries individuais
    collection: Collection = mongo_db["editais_pncp"]
    editais_cursor = collection.find({"_id": {"$in": mongo_ids}})

    # Monta dicionário {edital_mongodb_id: doc} para lookup O(1)
    editais_dict: dict[str, dict] = {}
    mongo_ids_str = [str(oid) for oid in mongo_ids]

    for doc in editais_cursor:
        str_id = str(doc["_id"])
        if str_id in mongo_ids_str:
            doc["_id"] = str_id
            editais_dict[str_id] = doc

    # 1 query bulk para processamento_ia
    proc_collection: Collection = mongo_db["processamento_ia"]
    processados_cursor = proc_collection.find(
        {"edital_mongodb_id": {"$in": mongo_ids_str}},
        {"edital_mongodb_id": 1, "resumo": 1, "categorias": 1},
    )
    processados_dict: dict[str, dict] = {}
    for proc in processados_cursor:
        processados_dict[proc["edital_mongodb_id"]] = proc

    # Monta resposta
    resultados = []
    for fav in favoritos:
        edital_data = None
        doc = editais_dict.get(fav.edital_mongodb_id)
        if doc:
            proc = processados_dict.get(fav.edital_mongodb_id)
            if proc:
                doc["resumo"] = proc.get("resumo")
                doc["categorias"] = proc.get("categorias", [])
            edital_data = _doc_para_response(doc)

        resultados.append(
            FavoritoResponse(
                uuid=fav.uuid,
                edital_mongodb_id=fav.edital_mongodb_id,
                status_kanban=fav.status_kanban,
                notas=fav.notas,
                data_criacao=fav.data_criacao,
                edital=edital_data,
            )
        )

    return resultados


@router.patch("/favoritos/{favorito_uuid}", response_model=FavoritoResponse)
def atualizar_favorito(
    favorito_uuid: str,
    payload: FavoritoUpdate,
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """
    Atualiza o status Kanban e/ou notas de um favorito.
    Usado para mover editais entre colunas do Kanban.
    """
    status_validos = ["PARA_LER", "LENDO", "ANALISADO", "ARQUIVADO"]
    if payload.status_kanban.upper() not in status_validos:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Status inválido. Valores: {', '.join(status_validos)}",
        )

    favorito = (
        db.query(EditalFavorito)
        .filter(
            EditalFavorito.uuid == favorito_uuid,
            EditalFavorito.usuario_id == usuario.id,
        )
        .first()
    )
    if not favorito:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Favorito não encontrado",
        )

    favorito.status_kanban = payload.status_kanban.upper()
    if payload.notas is not None:
        favorito.notas = payload.notas

    db.commit()
    db.refresh(favorito)

    return FavoritoResponse(
        uuid=favorito.uuid,
        edital_mongodb_id=favorito.edital_mongodb_id,
        status_kanban=favorito.status_kanban,
        notas=favorito.notas,
        data_criacao=favorito.data_criacao,
    )