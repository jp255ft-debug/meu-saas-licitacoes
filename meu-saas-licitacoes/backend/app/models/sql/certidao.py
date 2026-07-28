import uuid
from datetime import date, datetime
from typing import Optional

from sqlalchemy import String, Integer, Date, DateTime, Boolean, ForeignKey, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from backend.app.core.database import Base


class Certidao(Base):
    """
    Tabela de certidões negativas/de regularidade fiscal do usuário.
    Suporta radar de vencimentos: status automático + flags de alerta.
    """
    __tablename__ = "certidoes"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    uuid: Mapped[str] = mapped_column(
        UUID(as_uuid=False),
        default=lambda: str(uuid.uuid4()),
        unique=True,
        nullable=False,
        index=True,
    )
    usuario_id: Mapped[int] = mapped_column(
        Integer, ForeignKey("usuarios.id"), nullable=False
    )
    tipo: Mapped[str] = mapped_column(
        String(30), nullable=False
    )  # FEDERAL | ESTADUAL | MUNICIPAL | FGTS | TRABALHISTA | SIMPLES | INSS
    numero: Mapped[Optional[str]] = mapped_column(
        String(100), nullable=True, comment="Número do documento"
    )
    orgao_emissor: Mapped[Optional[str]] = mapped_column(
        String(255), nullable=True, comment="Órgão que emitiu a certidão"
    )
    data_emissao: Mapped[Optional[date]] = mapped_column(
        Date, nullable=True
    )
    data_validade: Mapped[date] = mapped_column(
        Date, nullable=False
    )
    status: Mapped[str] = mapped_column(
        String(20), nullable=False, default="VALIDA"
    )  # VALIDA | VENCENDO | VENCIDA
    alerta_enviado_7d: Mapped[bool] = mapped_column(
        Boolean, default=False, nullable=False,
        comment="Já enviou alerta de 7 dias?"
    )
    alerta_enviado_30d: Mapped[bool] = mapped_column(
        Boolean, default=False, nullable=False,
        comment="Já enviou alerta de 30 dias?"
    )
    arquivo_url: Mapped[Optional[str]] = mapped_column(
        String(500), nullable=True
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
    usuario = relationship("Usuario", back_populates="certidoes")