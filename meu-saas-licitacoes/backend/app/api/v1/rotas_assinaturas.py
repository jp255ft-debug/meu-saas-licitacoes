from datetime import datetime, timedelta

from fastapi import APIRouter, Depends, HTTPException, Request, status
from sqlalchemy.orm import Session

from backend.app.api.dependencias import get_current_user
from backend.app.core.database import get_db
from backend.app.models.sql.assinatura import Assinatura
from backend.app.models.sql.usuario import Usuario
from backend.app.schemas.assinatura import (
    PLANOS_CONFIG,
    AssinaturaCriar,
    AssinaturaResponse,
    CheckoutResponse,
    CicloCobrancaEnum,
    PlanoEnum,
    StatusAssinaturaEnum,
)
from backend.app.services.pagamento import PagamentoService

router = APIRouter(prefix="/assinaturas", tags=["Assinaturas"])


@router.post("/criar-checkout", response_model=CheckoutResponse)
def criar_checkout(
    data: AssinaturaCriar,
    db: Session = Depends(get_db),
    usuario: Usuario = Depends(get_current_user),
):
    """
    Cria uma sessão de checkout no Stripe para o usuário autenticado.
    Retorna a URL para redirecionamento ao checkout.
    """
    # Valida se o usuário já tem uma assinatura ativa
    assinatura_ativa = (
        db.query(Assinatura)
        .filter(
            Assinatura.usuario_id == usuario.id,
            Assinatura.status == "ativa",
        )
        .first()
    )
    if assinatura_ativa:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Usuário já possui uma assinatura ativa",
        )

    # Cria checkout no Stripe
    try:
        checkout_data = PagamentoService.criar_checkout(
            usuario_email=usuario.email,
            usuario_uuid=usuario.uuid,
            plano=data.plano,
            ciclo=data.ciclo,
            success_url=data.success_url,
            cancel_url=data.cancel_url,
        )
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Erro ao criar checkout: {str(e)}",
        )

    # Cria registro da assinatura no banco (status PENDENTE)
    config = PLANOS_CONFIG[data.plano.value]
    nova_assinatura = Assinatura(
        usuario_id=usuario.id,
        plano=data.plano.value,
        ciclo_cobranca=data.ciclo.value,
        valor_mensal=config["preco_mensal"],
        status="pendente",
        gateway_pagamento="stripe",
        gateway_subscription_id=checkout_data["subscription_id"],
        recursos=config["recursos"],
        data_inicio=datetime.now(),
        data_proximo_ciclo=datetime.now() + timedelta(days=30),
    )
    db.add(nova_assinatura)
    db.commit()
    db.refresh(nova_assinatura)

    return CheckoutResponse(
        checkout_url=checkout_data["checkout_url"],
        subscription_uuid=nova_assinatura.uuid,
    )


@router.get("/me", response_model=AssinaturaResponse)
def obter_minha_assinatura(
    db: Session = Depends(get_db),
    usuario: Usuario = Depends(get_current_user),
):
    """Retorna a assinatura atual do usuário autenticado."""
    assinatura = (
        db.query(Assinatura)
        .filter(Assinatura.usuario_id == usuario.id)
        .order_by(Assinatura.data_criacao.desc())
        .first()
    )

    if not assinatura:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Nenhuma assinatura encontrada",
        )
    return assinatura


@router.post("/cancelar/{uuid}")
def cancelar_assinatura(
    uuid: str,
    db: Session = Depends(get_db),
    usuario: Usuario = Depends(get_current_user),
):
    """Cancela a assinatura (no final do ciclo vigente)."""
    assinatura = (
        db.query(Assinatura)
        .filter(
            Assinatura.uuid == uuid,
            Assinatura.usuario_id == usuario.id,
        )
        .first()
    )

    if not assinatura:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Assinatura não encontrada",
        )

    if assinatura.status != "ativa":
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Assinatura não está ativa",
        )

    # Cancela no Stripe
    if (
        assinatura.gateway_pagamento == "stripe"
        and assinatura.gateway_subscription_id
    ):
        PagamentoService.cancelar_assinatura(
            assinatura.gateway_subscription_id
        )

    assinatura.status = "cancelada"
    assinatura.data_termino = datetime.now()
    db.commit()

    return {"message": "Assinatura cancelada com sucesso. Válida até o final do ciclo atual."}


@router.post("/webhook")
async def webhook_stripe(request: Request, db: Session = Depends(get_db)):
    """
    Endpoint para receber webhooks do Stripe.
    Processa: checkout.session.completed, customer.subscription.deleted, invoice.payment_failed
    """
    payload = await request.body()
    sig_header = request.headers.get("stripe-signature")

    if not sig_header:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Falta assinatura do Stripe",
        )

    try:
        evento = PagamentoService.processar_webhook(payload, sig_header)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST, detail=str(e)
        )

    # Processa checkout.completed
    if evento.get("evento") == "checkout.session.completed":
        usuario_uuid = evento.get("usuario_uuid")
        gateway_subscription_id = evento.get("gateway_subscription_id")

        if not usuario_uuid or not gateway_subscription_id:
            return {"status": "ignored", "reason": "dados faltando"}

        assinatura = (
            db.query(Assinatura)
            .filter(
                Assinatura.gateway_subscription_id == gateway_subscription_id,
                Assinatura.status == "pendente",
            )
            .first()
        )

        if assinatura:
            assinatura.status = "ativa"
            assinatura.data_inicio = datetime.now()
            assinatura.data_proximo_ciclo = datetime.now() + timedelta(days=30)
            db.commit()

    # Processa subscription.deleted
    elif evento.get("evento") == "customer.subscription.deleted":
        gateway_subscription_id = evento.get("gateway_subscription_id")
        if gateway_subscription_id:
            assinatura = (
                db.query(Assinatura)
                .filter(
                    Assinatura.gateway_subscription_id
                    == gateway_subscription_id
                )
                .first()
            )
            if assinatura:
                assinatura.status = "cancelada"
                assinatura.data_termino = datetime.now()
                db.commit()

    # Processa payment_failed
    elif evento.get("evento") == "invoice.payment_failed":
        gateway_subscription_id = evento.get("gateway_subscription_id")
        if gateway_subscription_id:
            assinatura = (
                db.query(Assinatura)
                .filter(
                    Assinatura.gateway_subscription_id
                    == gateway_subscription_id
                )
                .first()
            )
            if assinatura:
                assinatura.tentativas_falhas += 1
                if assinatura.tentativas_falhas >= 3:
                    assinatura.status = "expirada"
                db.commit()

    return {"status": "ok"}