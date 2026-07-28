"""
Schemas Pydantic para a API de contratações públicas (PNCP).
Define as estruturas de request e response para os endpoints de contratações.
"""
from datetime import datetime
from typing import Optional

from pydantic import BaseModel


class OrgaoEntidadeResponse(BaseModel):
    """Dados do órgão/entidade na resposta da API."""
    cnpj: Optional[str] = None
    nome: Optional[str] = None
    sigla: Optional[str] = None
    uf: Optional[str] = None
    municipio: Optional[str] = None


class ContratacaoResponse(BaseModel):
    """Schema de resposta para uma contratação do MongoDB."""
    id: str
    numero_controle_pncp: str
    ano_compra: int
    sequencial_compra: Optional[int] = None
    orgao_entidade: Optional[OrgaoEntidadeResponse] = None
    modalidade_nome: Optional[str] = None
    objeto_compra: str
    valor_total_estimado: Optional[float] = None
    data_publicacao_pncp: Optional[datetime] = None
    data_abertura_proposta: Optional[datetime] = None
    data_encerramento_proposta: Optional[datetime] = None
    link_edital: Optional[str] = None
    exclusividade_me_epp: bool = False
    cnaes_identificados: list[str] = []
    data_importacao: Optional[datetime] = None
    status_proposta: str = "aberta"


class ContratacaoFiltrar(BaseModel):
    """Query params para filtragem de contratações."""
    uf: Optional[str] = None
    modalidade: Optional[str] = None
    palavra_chave: Optional[str] = None
    exclusividade_me_epp: Optional[bool] = None
    valor_min: Optional[float] = None
    valor_max: Optional[float] = None
    ordenar_por: str = "data_publicacao_pncp"
    ordem: str = "desc"  # asc | desc
    pagina: int = 1
    tamanho: int = 20


class Paginacao(BaseModel):
    """Metadados de paginação."""
    pagina: int
    tamanho: int
    total: int
    total_paginas: int


class ContratacaoListaResponse(BaseModel):
    """Resposta paginada de contratações."""
    data: list[ContratacaoResponse]
    paginacao: Paginacao


class ContratacaoEstatisticas(BaseModel):
    """Estatísticas do dashboard de contratações."""
    total_contratacoes: int
    total_abertas: int
    exclusivas_me_epp: int
    por_uf: dict[str, int]
    valor_total_estimado: float
    prazo_medio_propostas_dias: Optional[float] = None