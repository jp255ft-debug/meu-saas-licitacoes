"""
Tabela de notificações enviadas aos usuários (WhatsApp, e-mail, in-app).
Cada registro representa uma tentativa de envio de alerta sobre uma contratação.
"""
import uuid
from datetime import datetime
from typing import Optional

from sqlalchemy import String, Integer, DateTime, Text, ForeignKey, Enum, Boolean, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from backend.app.core.database import Base

import enum


class TipoNotificacaoEnum(str, enum.Enum):
    WHATSAPP = "whatsapp"
    EMAIL = "email"
    IN_APP = "in_app"


class StatusNotificacaoEnum(str, enum.Enum):
    PENDENTE = "pendente"
    ENVIADO = "enviado"
    FALHA = "falha"
    LIDO = "lido"


class Notificacao(Base):
    """
    Tabela de notificações enviadas aos usuários.
    Referencia o MongoDB via contratacao_id (string), sem FK real.
    """
    __tablename__ = "notificacoes"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    uuid: Mapped[str] = mapped_column(
        UUID(as_uuid=False),
        default=lambda: str(uuid.uuid4()),
        unique=True,
        nullable=False,
        index=True,
    )
    usuario_id: Mapped[int] = mapped_column(
        Integer, ForeignKey("usuarios.id", ondelete="CASCADE"), nullable=False
    )
    contratacao_id: Mapped[str] = mapped_column(
        String(255), nullable=False, index=True,
        comment="ID do documento no MongoDB (coleção contratacoes_pncp)"
    )
    tipo: Mapped[TipoNotificacaoEnum] = mapped_column(
        Enum(TipoNotificacaoEnum), nullable=False
    )
    status: Mapped[StatusNotificacaoEnum] = mapped_column(
        Enum(StatusNotificacaoEnum), default=StatusNotificacaoEnum.PENDENTE
    )
    mensagem: Mapped[str] = mapped_column(Text, nullable=False)
    data_envio: Mapped[Optional[datetime]] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    data_leitura: Mapped[Optional[datetime]] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    erro: Mapped[Optional[str]] = mapped_column(
        String(255), nullable=True
    )
    tentativas: Mapped[int] = mapped_column(
        Integer, default=0, nullable=False
    )
    criado_em: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), nullable=False
    )

    # Relationships
    usuario = relationship("Usuario", back_populates="notificacoes")