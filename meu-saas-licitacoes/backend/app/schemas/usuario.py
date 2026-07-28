import re
from datetime import datetime
from typing import Optional

from pydantic import BaseModel, EmailStr, field_validator, model_validator


class PreferenciasAlertasUpdate(BaseModel):
    """Schema para atualização das preferências de alerta do usuário."""
    alertas_whatsapp: bool | None = None
    alertas_email: bool | None = None
    alertas_in_app: bool | None = None
    telefone_whatsapp: str | None = None
    configuracoes_match: dict | None = None


class UsuarioCreate(BaseModel):
    """Schema para criação de usuário (registro)."""
    nome: str
    email: EmailStr
    cpf_cnpj: str
    senha: str
    tipo: str = "MEI"  # MEI | EMPRESA

    @field_validator("cpf_cnpj")
    @classmethod
    def validar_cpf_cnpj(cls, v: str) -> str:
        """Remove máscara e valida formato (11 dígitos CPF ou 14 CNPJ)."""
        cleaned = re.sub(r"\D", "", v)
        if len(cleaned) not in (11, 14):
            raise ValueError("CPF deve ter 11 dígitos ou CNPJ 14 dígitos")
        return cleaned

    @field_validator("senha")
    @classmethod
    def validar_senha(cls, v: str) -> str:
        """Mínimo 8 caracteres."""
        if len(v) < 8:
            raise ValueError("Senha deve ter no mínimo 8 caracteres")
        return v

    @field_validator("tipo")
    @classmethod
    def validar_tipo(cls, v: str) -> str:
        if v.upper() not in ("MEI", "EMPRESA"):
            raise ValueError("Tipo deve ser MEI ou EMPRESA")
        return v.upper()

    @model_validator(mode="after")
    def validar_cpf_cnpj_por_tipo(self) -> "UsuarioCreate":
        """Valida se o tamanho do documento é compatível com o tipo."""
        cleaned = re.sub(r"\D", "", self.cpf_cnpj)
        if self.tipo == "MEI" and len(cleaned) != 11:
            raise ValueError("MEI deve informar CPF (11 dígitos)")
        if self.tipo == "EMPRESA" and len(cleaned) != 14:
            raise ValueError("Empresa deve informar CNPJ (14 dígitos)")
        return self


class UsuarioResponse(BaseModel):
    """Schema de resposta que expõe o UUID e dados públicos (nunca a senha)."""
    uuid: str
    nome: str
    email: str
    cpf_cnpj: str
    tipo: str
    ativo: bool
    data_criacao: datetime

    model_config = {"from_attributes": True}


class UsuarioUpdate(BaseModel):
    """Schema para atualização parcial do perfil."""
    nome: Optional[str] = None
    email: Optional[EmailStr] = None
    tipo: Optional[str] = None