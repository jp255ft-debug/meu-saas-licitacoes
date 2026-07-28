import uuid
from datetime import datetime
from typing import Optional

from sqlalchemy import String, Integer, DateTime, Text, ForeignKey, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from backend.app.core.database import Base


class EditalFavorito(Base):
    """
    Tabela de editais marcados como favoritos pelo usuário.
    Referencia o MongoDB via string edital_mongodb_id (sem FK real).
    Suporta status do Kanban (PARA_LER, LENDO, ANALISADO, ARQUIVADO).
    """
    __tablename__ = "editais_favoritos"

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
    edital_mongodb_id: Mapped[str] = mapped_column(
        String(255), nullable=False, index=True
    )
    status_kanban: Mapped[str] = mapped_column(
        String(20), nullable=False, default="PARA_LER"
    )  # PARA_LER | LENDO | ANALISADO | ARQUIVADO
    notas: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    data_criacao: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), nullable=False
    )

    # Relationships
    usuario = relationship("Usuario", back_populates="editais_favoritos")