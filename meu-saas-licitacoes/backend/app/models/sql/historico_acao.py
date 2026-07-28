import uuid
from datetime import datetime

from sqlalchemy import String, Integer, DateTime, ForeignKey, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from backend.app.core.database import Base


class HistoricoAcao(Base):
    """
    Auditoria simples: registra ações do usuário no sistema.
    Ex: "usuario_criou_conta", "usuario_favoritou_edital", "certidao_expirou".
    """
    __tablename__ = "historico_acoes"

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
    acao: Mapped[str] = mapped_column(
        String(100), nullable=False, index=True
    )
    entidade: Mapped[str] = mapped_column(
        String(50), nullable=True
    )  # ex: "edital", "certidao", "assinatura"
    entidade_id: Mapped[str] = mapped_column(
        String(255), nullable=True
    )  # uuid da entidade afetada
    data: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), nullable=False
    )

    # Relationships
    usuario = relationship("Usuario", back_populates="historico_acoes")