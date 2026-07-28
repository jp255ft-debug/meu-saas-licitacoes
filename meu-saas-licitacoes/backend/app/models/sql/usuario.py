import uuid
from datetime import datetime
from typing import Optional

from sqlalchemy import String, Boolean, DateTime, func, JSON
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from backend.app.core.database import Base


class Usuario(Base):
    """
    Tabela central de usuários do SaaS.
    IDs auto-incrementados NUNCA expostos; usa-se uuid externo.
    """
    __tablename__ = "usuarios"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    uuid: Mapped[str] = mapped_column(
        UUID(as_uuid=False),
        default=lambda: str(uuid.uuid4()),
        unique=True,
        nullable=False,
        index=True,
    )
    nome: Mapped[str] = mapped_column(String(255), nullable=False)
    email: Mapped[str] = mapped_column(
        String(255), unique=True, nullable=False, index=True
    )
    cpf_cnpj: Mapped[str] = mapped_column(
        String(18), unique=True, nullable=False
    )
    senha_hash: Mapped[str] = mapped_column(String(255), nullable=False)
    tipo: Mapped[str] = mapped_column(
        String(20), nullable=False, default="MEI"
    )  # MEI | EMPRESA
    ativo: Mapped[bool] = mapped_column(Boolean, default=True, nullable=False)

    # --- Preferências de alerta ---
    alertas_whatsapp: Mapped[bool] = mapped_column(
        Boolean, default=True, nullable=False
    )
    alertas_email: Mapped[bool] = mapped_column(
        Boolean, default=True, nullable=False
    )
    alertas_in_app: Mapped[bool] = mapped_column(
        Boolean, default=True, nullable=False
    )
    telefone_whatsapp: Mapped[Optional[str]] = mapped_column(
        String(20), nullable=True,
        comment="Telefone para WhatsApp (separado do telefone comum)"
    )
    configuracoes_match: Mapped[Optional[dict]] = mapped_column(
        JSON, nullable=True,
        comment="JSON com ufs, cnaes, palavras_chave para smart match"
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
    assinatura = relationship("Assinatura", back_populates="usuario", uselist=False)
    certidoes = relationship("Certidao", back_populates="usuario")
    editais_favoritos = relationship("EditalFavorito", back_populates="usuario")
    historico_acoes = relationship("HistoricoAcao", back_populates="usuario")
    notificacoes = relationship("Notificacao", back_populates="usuario", cascade="all, delete-orphan")
