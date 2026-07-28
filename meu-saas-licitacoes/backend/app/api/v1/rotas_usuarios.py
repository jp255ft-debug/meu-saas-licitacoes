from math import ceil

from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy.orm import Session

from backend.app.core.database import get_db
from backend.app.core.security import hash_password, verify_password
from backend.app.api.dependencias import get_current_user
from backend.app.models.sql.usuario import Usuario
from backend.app.models.sql.notificacao import Notificacao
from backend.app.schemas.usuario import (
    PreferenciasAlertasUpdate,
    UsuarioCreate,
    UsuarioResponse,
    UsuarioUpdate,
)

router = APIRouter(prefix="/usuarios", tags=["Usuários"])


@router.post("", response_model=UsuarioResponse, status_code=status.HTTP_201_CREATED)
def criar_usuario(payload: UsuarioCreate, db: Session = Depends(get_db)):
    """Registra um novo usuário no sistema."""
    # Verifica se email já existe
    if db.query(Usuario).filter(Usuario.email == payload.email).first():
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Email já cadastrado",
        )

    # Verifica se CPF/CNPJ já existe
    if db.query(Usuario).filter(Usuario.cpf_cnpj == payload.cpf_cnpj).first():
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="CPF/CNPJ já cadastrado",
        )

    usuario = Usuario(
        nome=payload.nome,
        email=payload.email,
        cpf_cnpj=payload.cpf_cnpj,
        senha_hash=hash_password(payload.senha),
        tipo=payload.tipo,
    )
    db.add(usuario)
    db.commit()
    db.refresh(usuario)
    return usuario


@router.get("/me", response_model=UsuarioResponse)
def obter_perfil(usuario: Usuario = Depends(get_current_user)):
    """Retorna os dados do usuário autenticado."""
    return usuario


@router.put("/me", response_model=UsuarioResponse)
def atualizar_perfil(
    payload: UsuarioUpdate,
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """Atualiza dados do perfil do usuário autenticado."""
    if payload.nome is not None:
        usuario.nome = payload.nome
    if payload.email is not None:
        if db.query(Usuario).filter(Usuario.email == payload.email, Usuario.id != usuario.id).first():
            raise HTTPException(
                status_code=status.HTTP_409_CONFLICT,
                detail="Email já está em uso",
            )
        usuario.email = payload.email
    if payload.tipo is not None:
        usuario.tipo = payload.tipo.upper()

    db.commit()
    db.refresh(usuario)
    return usuario


@router.delete("/me", status_code=status.HTTP_204_NO_CONTENT)
def desativar_conta(
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """Desativa (soft-delete) a conta do usuário autenticado."""
    usuario.ativo = False
    db.commit()


@router.patch("/me/preferencias-alertas")
def atualizar_preferencias_alertas(
    dados: PreferenciasAlertasUpdate,
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """
    Atualiza as preferências de alerta do usuário autenticado.
    Permite ativar/desativar notificações por WhatsApp, e-mail e in-app,
    além de cadastrar telefone para WhatsApp e configurações de match.
    """
    if dados.alertas_whatsapp is not None:
        usuario.alertas_whatsapp = dados.alertas_whatsapp
    if dados.alertas_email is not None:
        usuario.alertas_email = dados.alertas_email
    if dados.alertas_in_app is not None:
        usuario.alertas_in_app = dados.alertas_in_app
    if dados.telefone_whatsapp is not None:
        usuario.telefone_whatsapp = dados.telefone_whatsapp
    if dados.configuracoes_match is not None:
        usuario.configuracoes_match = dados.configuracoes_match

    db.commit()

    return {
        "message": "Preferências atualizadas com sucesso",
        "preferencias": {
            "alertas_whatsapp": usuario.alertas_whatsapp,
            "alertas_email": usuario.alertas_email,
            "alertas_in_app": usuario.alertas_in_app,
            "telefone_whatsapp": usuario.telefone_whatsapp,
            "configuracoes_match": usuario.configuracoes_match,
        },
    }


@router.get("/me/notificacoes")
def listar_notificacoes(
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
    pagina: int = Query(1, ge=1, description="Número da página"),
    tamanho: int = Query(20, ge=1, le=100, description="Itens por página"),
    tipo: str | None = Query(None, description="Filtrar por tipo: whatsapp | email | in_app"),
    status: str | None = Query(None, description="Filtrar por status: pendente | enviado | falha | lido"),
):
    """
    Lista o histórico de notificações do usuário autenticado.
    Suporta paginação e filtros opcionais por tipo e status.
    """
    query = db.query(Notificacao).filter(Notificacao.usuario_id == usuario.id)

    if tipo:
        query = query.filter(Notificacao.tipo == tipo)
    if status:
        query = query.filter(Notificacao.status == status)

    total = query.count()
    skip = (pagina - 1) * tamanho

    items = (
        query.order_by(Notificacao.criado_em.desc())
        .offset(skip)
        .limit(tamanho)
        .all()
    )

    return {
        "items": [
            {
                "uuid": n.uuid,
                "tipo": n.tipo.value,
                "status": n.status.value,
                "contratacao_id": n.contratacao_id,
                "mensagem": n.mensagem[:200],  # preview
                "data_envio": n.data_envio.isoformat() if n.data_envio else None,
                "data_leitura": n.data_leitura.isoformat() if n.data_leitura else None,
                "erro": n.erro,
                "tentativas": n.tentativas,
                "criado_em": n.criado_em.isoformat(),
            }
            for n in items
        ],
        "total": total,
        "pagina": pagina,
        "tamanho": tamanho,
        "total_paginas": ceil(total / tamanho) if total > 0 else 0,
    }
