import uuid
from datetime import datetime
from typing import Optional

from sqlalchemy import String, Integer, Float, DateTime, Boolean, ForeignKey, func, JSON
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from backend.app.core.database import Base


class Assinatura(Base):
    """
    Tabela de assinaturas (planos) vinculadas ao usuário.
    Suporta múltiplos gateways de pagamento (Stripe, etc).
    """
    __tablename__ = "assinaturas"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    uuid: Mapped[str] = mapped_column(
        UUID(as_uuid=False),
        default=lambda: str(uuid.uuid4()),
        unique=True,
        nullable=False,
        index=True,
    )
    usuario_id: Mapped[int] = mapped_column(
        Integer, ForeignKey("usuarios.id"), nullable=False, unique=True
    )
    plano: Mapped[str] = mapped_column(
        String(20), nullable=False, default="trial"
    )  # trial | starter | pro
    ciclo_cobranca: Mapped[str] = mapped_column(
        String(10), nullable=False, default="mensal"
    )  # mensal | anual
    valor_mensal: Mapped[float] = mapped_column(
        Float, nullable=False, default=0.0
    )
    status: Mapped[str] = mapped_column(
        String(20), nullable=False, default="pendente"
    )  # pendente | ativa | cancelada | expirada | reembolsada
    gateway_pagamento: Mapped[Optional[str]] = mapped_column(
        String(30), nullable=True, default="stripe"
    )
    gateway_subscription_id: Mapped[Optional[str]] = mapped_column(
        String(255), nullable=True,
        comment="ID da assinatura no gateway (ex: sub_xxx no Stripe)"
    )
    recursos: Mapped[Optional[dict]] = mapped_column(
        JSON, nullable=True,
        comment="JSON com limites do plano (palavras_chave, recursos_liberados)"
    )
    tentativas_falhas: Mapped[int] = mapped_column(
        Integer, default=0, nullable=False,
        comment="Número de tentativas de cobrança falhas consecutivas"
    )
    data_inicio: Mapped[Optional[datetime]] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    data_proximo_ciclo: Mapped[Optional[datetime]] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    data_termino: Mapped[Optional[datetime]] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    data_criacao: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), nullable=False
    )
    data_atualizacao: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
        onupdate=func.now(),
        nullable=False,
    )

    # Relationships
    usuario = relationship("Usuario", back_populates="assinatura")