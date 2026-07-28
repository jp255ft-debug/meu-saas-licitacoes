"""
Schema/documento para a coleção 'pca_pncp' no MongoDB.
Cada documento representa um Plano de Contratações Anual (PCA) de um órgão público,
contendo todos os itens que o órgão pretende contratar no ano.
Fonte: API de consulta do PNCP (Portal Nacional de Contratações Públicas).
"""
from datetime import datetime
from typing import Optional

from pydantic import BaseModel, Field


class OrgaoEntidadePCA(BaseModel):
    """Dados do órgão/entidade responsável pelo PCA."""
    cnpj: Optional[str] = Field(None, description="CNPJ do órgão")
    nome: Optional[str] = Field(None, description="Nome do órgão")
    sigla: Optional[str] = Field(None, description="Sigla do órgão")
    uf: Optional[str] = Field(None, description="UF do órgão")
    municipio: Optional[str] = Field(None, description="Município do órgão")


class ItemPCA(BaseModel):
    """Item individual do Plano de Contratações Anual."""
    codigo_classificacao: Optional[str] = Field(
        None, description="Código de classificação do item"
    )
    descricao_item: str = Field(
        ..., description="Descrição do item a ser contratado"
    )
    quantidade_estimada: Optional[float] = Field(
        None, description="Quantidade estimada"
    )
    valor_estimado: Optional[float] = Field(
        None, description="Valor unitário estimado"
    )
    valor_total_estimado: Optional[float] = Field(
        None, description="Valor total estimado"
    )
    unidade_medida: Optional[str] = Field(
        None, description="Unidade de medida"
    )
    cnae_identificado: Optional[str] = Field(
        None, description="CNAE identificado via classificação do item"
    )


class PCAPNCP(BaseModel):
    """
    Estrutura de um PCA no MongoDB.
    O _id do MongoDB gerado automaticamente serve como referência.
    O campo idPca é o identificador único do PNCP.
    """
    id: Optional[str] = Field(None, alias="_id")
    id_pca: str = Field(..., description="ID único do PCA no PNCP")
    ano: int = Field(..., description="Ano de referência do PCA")
    orgao_entidade: Optional[OrgaoEntidadePCA] = Field(
        None, description="Dados do órgão responsável"
    )
    itens: list[ItemPCA] = Field(
        default_factory=list, description="Itens planejados para contratação"
    )
    data_publicacao: Optional[datetime] = Field(
        None, description="Data de publicação do PCA"
    )
    data_importacao: datetime = Field(
        default_factory=datetime.utcnow,
        description="Data de importação no sistema",
    )
    origem: str = Field(
        default="api_pncp", description="Origem dos dados"
    )
    versao: str = Field(default="1.0", description="Versão do schema")

    class Config:
        populate_by_name = True
        json_encoders = {
            datetime: lambda v: v.isoformat(),
        }