from datetime import datetime
from typing import Optional

from pydantic import BaseModel


class EditalFiltrar(BaseModel):
    """Query params para listagem de editais."""
    uf: Optional[str] = None
    modalidade: Optional[str] = None
    palavra_chave: Optional[str] = None
    categoria: Optional[str] = None
    data_inicio: Optional[datetime] = None
    data_fim: Optional[datetime] = None
    valor_min: Optional[float] = None
    valor_max: Optional[float] = None
    ordenar_por: str = "data_publicacao"
    ordem: str = "desc"  # asc | desc
    pagina: int = 1
    tamanho: int = 20


class EditalResponse(BaseModel):
    """Schema de resposta para um edital do MongoDB."""
    id: str
    pncp_id: str
    orgao_nome: str
    orgao_cnpj: str
    uf: str
    cidade: Optional[str] = None
    modalidade: str
    numero: str
    ano: int
    objeto: str
    valor_estimado: Optional[float] = None
    data_publicacao: Optional[datetime] = None
    data_abertura: Optional[datetime] = None
    url_pncp: Optional[str] = None
    documento_url: Optional[str] = None
    resumo: Optional[str] = None
    categorias: list[str] = []
    score_relevancia: Optional[float] = None


class Paginacao(BaseModel):
    """Metadados de paginação."""
    pagina: int
    tamanho: int
    total: int
    total_paginas: int


class EditalListaResponse(BaseModel):
    """Resposta paginada de editais."""
    data: list[EditalResponse]
    paginacao: Paginacao


class FavoritoCreate(BaseModel):
    """Schema para favoritar um edital."""
    edital_mongodb_id: str
    notas: Optional[str] = None


class FavoritoUpdate(BaseModel):
    """Schema para atualizar favorito (kanban)."""
    status_kanban: str  # PARA_LER | LENDO | ANALISADO | ARQUIVADO
    notas: Optional[str] = None


class FavoritoResponse(BaseModel):
    """Schema de resposta para favorito."""
    uuid: str
    edital_mongodb_id: str
    status_kanban: str
    notas: Optional[str] = None
    data_criacao: datetime
    edital: Optional[EditalResponse] = None

    model_config = {"from_attributes": True}