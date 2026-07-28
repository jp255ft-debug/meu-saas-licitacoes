from datetime import date, datetime
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy.orm import Session

from backend.app.api.dependencias import get_current_user
from backend.app.core.database import get_db
from backend.app.models.sql.certidao import Certidao
from backend.app.models.sql.usuario import Usuario
from backend.app.schemas.certidao import (
    CertidaoCreate,
    CertidaoResponse,
    CertidaoUpdate,
    StatusCertidao,
)

router = APIRouter(prefix="/certidoes", tags=["Certidões"])


def _calcular_status(data_validade: date) -> str:
    """
    Calcula o status da certidão com base na data de validade.
    - > 30 dias: VALIDA
    - 1 a 30 dias: VENCENDO
    - Hoje ou passado: VENCIDA
    """
    dias = (data_validade - date.today()).days
    if dias < 0:
        return "VENCIDA"
    elif dias <= 30:
        return "VENCENDO"
    return "VALIDA"


@router.post("", response_model=CertidaoResponse, status_code=status.HTTP_201_CREATED)
def criar_certidao(
    payload: CertidaoCreate,
    db: Session = Depends(get_db),
    usuario: Usuario = Depends(get_current_user),
):
    """Cadastra uma nova certidão para o usuário autenticado."""
    certidao = Certidao(
        usuario_id=usuario.id,
        tipo=payload.tipo.value,
        numero=payload.numero,
        orgao_emissor=payload.orgao_emissor,
        data_emissao=payload.data_emissao,
        data_validade=payload.data_validade,
        status=_calcular_status(payload.data_validade),
        arquivo_url=payload.arquivo_url,
    )
    db.add(certidao)
    db.commit()
    db.refresh(certidao)
    return certidao


@router.get("", response_model=list[CertidaoResponse])
def listar_certidoes(
    status: Optional[str] = Query(None, description="Filtrar por status: VALIDA, VENCENDO, VENCIDA"),
    db: Session = Depends(get_db),
    usuario: Usuario = Depends(get_current_user),
):
    """Lista as certidões do usuário autenticado. Opcionalmente filtra por status."""
    # Atualiza status antes de listar
    hoje = date.today()
    db.query(Certidao).filter(
        Certidao.usuario_id == usuario.id,
        Certidao.data_validade < hoje,
        Certidao.status != "VENCIDA",
    ).update({"status": "VENCIDA"}, synchronize_session=False)
    db.commit()

    query = db.query(Certidao).filter(Certidao.usuario_id == usuario.id)
    if status:
        query = query.filter(Certidao.status == status.upper())
    query = query.order_by(Certidao.data_validade.asc())
    return query.all()


@router.get("/{uuid}", response_model=CertidaoResponse)
def obter_certidao(
    uuid: str,
    db: Session = Depends(get_db),
    usuario: Usuario = Depends(get_current_user),
):
    """Retorna os detalhes de uma certidão específica."""
    certidao = (
        db.query(Certidao)
        .filter(Certidao.uuid == uuid, Certidao.usuario_id == usuario.id)
        .first()
    )
    if not certidao:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Certidão não encontrada",
        )
    return certidao


@router.put("/{uuid}", response_model=CertidaoResponse)
def atualizar_certidao(
    uuid: str,
    payload: CertidaoUpdate,
    db: Session = Depends(get_db),
    usuario: Usuario = Depends(get_current_user),
):
    """Atualiza uma certidão. Recalcula status se data_validade for alterada."""
    certidao = (
        db.query(Certidao)
        .filter(Certidao.uuid == uuid, Certidao.usuario_id == usuario.id)
        .first()
    )
    if not certidao:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Certidão não encontrada",
        )

    # Aplica campos fornecidos
    update_data = payload.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        if key == "tipo" and value is not None:
            setattr(certidao, key, value.value)
        else:
            setattr(certidao, key, value)

    # Recalcula status se data_validade foi alterada
    if "data_validade" in update_data:
        certidao.status = _calcular_status(certidao.data_validade)
        # Reseta alertas para permitir novo disparo
        certidao.alerta_enviado_7d = False
        certidao.alerta_enviado_30d = False

    certidao.data_atualizacao = datetime.now()
    db.commit()
    db.refresh(certidao)
    return certidao


@router.delete("/{uuid}", status_code=status.HTTP_204_NO_CONTENT)
def deletar_certidao(
    uuid: str,
    db: Session = Depends(get_db),
    usuario: Usuario = Depends(get_current_user),
):
    """Remove uma certidão."""
    certidao = (
        db.query(Certidao)
        .filter(Certidao.uuid == uuid, Certidao.usuario_id == usuario.id)
        .first()
    )
    if not certidao:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Certidão não encontrada",
        )
    db.delete(certidao)
    db.commit()


@router.post("/atualizar-status", status_code=200, include_in_schema=False)
def atualizar_status_em_massa(db: Session = Depends(get_db)):
    """
    Endpoint interno para o worker atualizar status de todas as certidões.
    Não exposto no Swagger (include_in_schema=False).
    """
    certidoes = db.query(Certidao).all()
    hoje = date.today()
    contagem = 0
    for cert in certidoes:
        novo_status = _calcular_status(cert.data_validade)
        if cert.status != novo_status:
            cert.status = novo_status
            if novo_status == "VALIDA":
                cert.alerta_enviado_7d = False
                cert.alerta_enviado_30d = False
            contagem += 1
    db.commit()
    return {"message": f"Status atualizado para {contagem} certidões"}