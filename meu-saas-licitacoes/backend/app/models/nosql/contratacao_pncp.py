"""
Schema/documento para a coleção 'contratacoes_pncp' no MongoDB.
Cada documento representa uma contratação com recebimento de propostas aberto,
obtida via API de consulta do PNCP (Portal Nacional de Contratações Públicas).
"""
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class OrgaoEntidade(BaseModel):
    """Dados do órgão/entidade responsável pela contratação."""
    cnpj: Optional[str] = Field(None, description="CNPJ do órgão")
    nome: Optional[str] = Field(None, description="Nome do órgão")
    sigla: Optional[str] = Field(None, description="Sigla do órgão")
    uf: Optional[str] = Field(None, description="UF do órgão")
    municipio: Optional[str] = Field(None, description="Município do órgão")


class ContratacaoPNCP(BaseModel):
    """
    Estrutura de uma contratação no MongoDB.
    O _id do MongoDB gerado automaticamente serve como referência
    para tabelas relacionais no PostgreSQL (ex: favoritos, histórico).
    """
    id: Optional[str] = Field(None, alias="_id")
    numero_controle_pncp: str = Field(
        ..., description="Número único de controle do PNCP (formato: XXXXXX)")
    ano_compra: int = Field(..., description="Ano da compra")
    sequencial_compra: Optional[int] = Field(None, description="Sequencial da compra")
    orgao_entidade: Optional[OrgaoEntidade] = Field(
        None, description="Dados do órgão responsável")
    modalidade_nome: Optional[str] = Field(
        None, description="Nome da modalidade de licitação")
    objeto_compra: str = Field(..., description="Descrição do objeto da compra")
    valor_total_estimado: Optional[float] = Field(
        None, description="Valor total estimado da contratação")
    data_publicacao_pncp: Optional[datetime] = Field(
        None, description="Data de publicação no PNCP")
    data_abertura_proposta: Optional[datetime] = Field(
        None, description="Data de abertura das propostas")
    data_encerramento_proposta: Optional[datetime] = Field(
        None, description="Data de encerramento das propostas")
    link_edital: Optional[str] = Field(
        None, description="URL do edital / link de acesso")
    exclusividade_me_epp: bool = Field(
        default=False,
        description="Indica se a contratação é exclusiva para ME/EPP (Lei Complementar 123/2006)")
    cnaes_identificados: list[str] = Field(
        default_factory=list,
        description="CNAES identificados na contratação para smart match")
    dados_brutos: Optional[dict] = Field(
        None, description="JSON completo retornado pela API do PNCP")
    data_importacao: datetime = Field(
        default_factory=datetime.utcnow, description="Data de importação no sistema")
    status_proposta: str = Field(
        default="aberta",
        description="Status da proposta: aberta | encerrada ")

    class Config:
        populate_by_name = True
        json_encoders = {
            datetime: lambda v: v.isoformat(),
        }