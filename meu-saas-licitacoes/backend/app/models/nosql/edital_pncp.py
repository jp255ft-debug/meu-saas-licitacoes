"""
Schema/documento para a coleção 'editais_pncp' no MongoDB.
Cada documento representa um edital bruto raspado do Portal Nacional de Compras Públicas (PNCP).
"""
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class EditalPNCP(BaseModel):
    """
    Estrutura de um edital no MongoDB.
    O _id do MongoDB gerado automaticamente serve como referência
    para a tabela 'editais_favoritos' no PostgreSQL (campo edital_mongodb_id).
    """
    id: Optional[str] = Field(None, alias="_id")
    pncp_id: str = Field(..., description="ID único do PNCP")
    orgao_nome: str = Field(..., description="Nome do órgão público")
    orgao_cnpj: str = Field(..., description="CNPJ do órgão")
    uf: str = Field(..., description="UF do órgão")
    cidade: Optional[str] = Field(None, description="Cidade do órgão")
    modalidade: str = Field(..., description="Modalidade da licitação")
    numero: str = Field(..., description="Número do edital")
    ano: int = Field(..., description="Ano do edital")
    objeto: str = Field(..., description="Descrição/resumo do objeto")
    valor_estimado: Optional[float] = Field(None, description="Valor estimado da licitação")
    data_publicacao: Optional[datetime] = Field(None)
    data_abertura: Optional[datetime] = Field(None)
    url_pncp: Optional[str] = Field(None, description="URL original no PNCP")
    documento_url: Optional[str] = Field(None, description="URL do PDF do edital")
    dados_brutos: Optional[dict] = Field(None, description="JSON completo retornado pela API do PNCP")
    data_coleta: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        populate_by_name = True
        json_encoders = {
            datetime: lambda v: v.isoformat(),
        }