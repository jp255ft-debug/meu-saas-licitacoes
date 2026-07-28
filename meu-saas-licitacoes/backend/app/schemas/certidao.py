from datetime import date, datetime
from enum import Enum
from typing import Optional

from pydantic import BaseModel, field_validator


class TipoCertidao(str, Enum):
    FEDERAL = "FEDERAL"
    ESTADUAL = "ESTADUAL"
    MUNICIPAL = "MUNICIPAL"
    FGTS = "FGTS"
    TRABALHISTA = "TRABALHISTA"
    SIMPLES = "SIMPLES"
    INSS = "INSS"


class StatusCertidao(str, Enum):
    VALIDA = "VALIDA"
    VENCENDO = "VENCENDO"
    VENCIDA = "VENCIDA"


# ---------- Request ----------


class CertidaoCreate(BaseModel):
    """Schema para cadastro de nova certidão."""
    tipo: TipoCertidao
    numero: Optional[str] = None
    orgao_emissor: Optional[str] = None
    data_emissao: Optional[date] = None
    data_validade: date
    arquivo_url: Optional[str] = None

    @field_validator("data_validade")
    @classmethod
    def validar_data_futura(cls, v: date) -> date:
        """A data de validade não pode estar no passado."""
        if v < date.today():
            raise ValueError("Data de validade não pode estar no passado")
        return v


class CertidaoUpdate(BaseModel):
    """Schema para atualização parcial de certidão."""
    tipo: Optional[TipoCertidao] = None
    numero: Optional[str] = None
    orgao_emissor: Optional[str] = None
    data_emissao: Optional[date] = None
    data_validade: Optional[date] = None
    arquivo_url: Optional[str] = None

    @field_validator("data_validade")
    @classmethod
    def validar_data_futura(cls, v: Optional[date]) -> Optional[date]:
        if v is not None and v < date.today():
            raise ValueError("Data de validade não pode estar no passado")
        return v


# ---------- Response ----------


class CertidaoResponse(BaseModel):
    """Schema de resposta com status calculado."""
    uuid: str
    tipo: TipoCertidao
    numero: Optional[str] = None
    orgao_emissor: Optional[str] = None
    data_emissao: Optional[date] = None
    data_validade: date
    status: StatusCertidao
    alerta_enviado_7d: bool
    alerta_enviado_30d: bool
    arquivo_url: Optional[str] = None
    data_criacao: datetime
    data_atualizacao: datetime

    model_config = {"from_attributes": True}