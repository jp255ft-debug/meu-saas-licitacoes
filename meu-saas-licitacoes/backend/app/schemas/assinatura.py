from datetime import datetime
from enum import Enum
from typing import Any, Dict, Optional

from pydantic import BaseModel


class PlanoEnum(str, Enum):
    STARTER = "starter"
    PRO = "pro"
    TRIAL = "trial"


class CicloCobrancaEnum(str, Enum):
    MENSAL = "mensal"
    ANUAL = "anual"


class StatusAssinaturaEnum(str, Enum):
    ATIVA = "ativa"
    CANCELADA = "cancelada"
    EXPIRADA = "expirada"
    PENDENTE = "pendente"
    REEMBOLSADA = "reembolsada"


# ---------- Planos disponíveis (centralizado) ----------
PLANOS_CONFIG = {
    "starter": {
        "nome": "Starter",
        "preco_mensal": 97.00,
        "preco_anual": 970.00,  # 10% off
        "recursos": {
            "palavras_chave": 3,
            "alertas_diarios": True,
            "ia_resumos": False,
            "kanban": False,
            "whatsapp_alerts": False,
        },
    },
    "pro": {
        "nome": "Pro",
        "preco_mensal": 197.00,
        "preco_anual": 1970.00,  # 10% off
        "recursos": {
            "palavras_chave": -1,  # ilimitado
            "alertas_diarios": True,
            "ia_resumos": True,
            "kanban": True,
            "whatsapp_alerts": True,
        },
    },
}


# ---------- Request ----------


class AssinaturaCriar(BaseModel):
    """Schema para criação de checkout de assinatura."""
    plano: PlanoEnum
    ciclo: CicloCobrancaEnum = CicloCobrancaEnum.MENSAL
    success_url: str = "https://meusaas.com/dashboard"
    cancel_url: str = "https://meusaas.com/planos"


# ---------- Response ----------


class AssinaturaResponse(BaseModel):
    """Schema de resposta com dados da assinatura."""
    uuid: str
    plano: PlanoEnum
    ciclo_cobranca: CicloCobrancaEnum
    valor_mensal: float
    status: StatusAssinaturaEnum
    data_inicio: Optional[datetime] = None
    data_proximo_ciclo: Optional[datetime] = None
    data_termino: Optional[datetime] = None
    recursos: Optional[Dict[str, Any]] = None

    model_config = {"from_attributes": True}


class CheckoutResponse(BaseModel):
    """Schema de resposta com URL do checkout."""
    checkout_url: str
    subscription_uuid: str