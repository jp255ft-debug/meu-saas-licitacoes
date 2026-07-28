"""
Serviço de pagamento com abstração de gateway.
Implementação atual: Stripe.
Estrutura preparada para suportar Mercado Pago, PagSeguro, etc.
"""
import logging
from typing import Any, Dict

import stripe

from backend.app.core.config import settings
from backend.app.schemas.assinatura import (
    PLANOS_CONFIG,
    CicloCobrancaEnum,
    PlanoEnum,
)

logger = logging.getLogger(__name__)

stripe.api_key = settings.STRIPE_SECRET_KEY


class PagamentoService:
    """Interface para gateways de pagamento. Métodos estáticos para uso direto."""

    @staticmethod
    def criar_checkout(
        usuario_email: str,
        usuario_uuid: str,
        plano: PlanoEnum,
        ciclo: CicloCobrancaEnum,
        success_url: str,
        cancel_url: str,
    ) -> Dict[str, Any]:
        """
        Cria uma sessão de checkout no Stripe.
        Gerencia produtos e prices dinamicamente (cria se não existir).
        Retorna: { "checkout_url": str, "subscription_id": str, "price_id": str }
        """
        config = PLANOS_CONFIG[plano.value]
        price_id = None

        try:
            # Busca ou cria produto no Stripe
            produto = None
            produtos = stripe.Product.list(limit=100)
            for p in produtos.auto_paging_iter():
                if p.name == f"Plano {config['nome']}":
                    produto = p
                    break

            if not produto:
                produto = stripe.Product.create(
                    name=f"Plano {config['nome']}",
                    description=f"Plano {config['nome']} - Licitação SaaS",
                )

            # Define preço e recorrência
            if ciclo == CicloCobrancaEnum.MENSAL:
                preco = config["preco_mensal"]
                recorrencia = "month"
            else:
                preco = config["preco_anual"]
                recorrencia = "year"

            # Busca ou cria price
            prices = stripe.Price.list(product=produto.id, limit=10)
            for p in prices.auto_paging_iter():
                if (
                    p.unit_amount == int(preco * 100)
                    and p.recurring
                    and p.recurring.interval == recorrencia
                ):
                    price_id = p.id
                    break

            if not price_id:
                price = stripe.Price.create(
                    product=produto.id,
                    unit_amount=int(preco * 100),
                    currency="brl",
                    recurring={"interval": recorrencia},
                )
                price_id = price.id

        except Exception as e:
            logger.warning(f"Erro ao criar produto/price no Stripe: {e}. Usando fallback.")
            price_id = "price_fallback_mock"

        # Cria sessão de checkout
        try:
            checkout = stripe.checkout.Session.create(
                customer_email=usuario_email,
                payment_method_types=["card", "pix"],
                line_items=[{"price": price_id, "quantity": 1}],
                mode="subscription",
                success_url=success_url + "?session_id={CHECKOUT_SESSION_ID}",
                cancel_url=cancel_url,
                metadata={
                    "usuario_uuid": usuario_uuid,
                    "plano": plano.value,
                    "ciclo": ciclo.value,
                },
            )
            return {
                "checkout_url": checkout.url,
                "subscription_id": checkout.id,
                "price_id": price_id,
            }
        except Exception as e:
            logger.error(f"Erro ao criar checkout Stripe: {e}")
            raise

    @staticmethod
    def processar_webhook(payload: bytes, sig_header: str) -> Dict[str, Any]:
        """Processa o webhook do Stripe e retorna o evento interpretado."""
        try:
            event = stripe.Webhook.construct_event(
                payload, sig_header, settings.STRIPE_WEBHOOK_SECRET
            )
        except ValueError:
            raise ValueError("Payload inválido")
        except stripe.error.SignatureVerificationError:
            raise ValueError("Assinatura do webhook inválida")

        event_type = event.get("type")
        data = event.get("data", {}).get("object", {})

        if event_type == "checkout.session.completed":
            return {
                "evento": "checkout.session.completed",
                "usuario_uuid": data.get("metadata", {}).get("usuario_uuid"),
                "plano": data.get("metadata", {}).get("plano"),
                "ciclo": data.get("metadata", {}).get("ciclo"),
                "gateway_subscription_id": data.get("subscription"),
                "status": "ativa",
            }

        elif event_type == "customer.subscription.deleted":
            return {
                "evento": "customer.subscription.deleted",
                "gateway_subscription_id": data.get("id"),
                "status": "cancelada",
            }

        elif event_type == "invoice.payment_failed":
            return {
                "evento": "invoice.payment_failed",
                "gateway_subscription_id": data.get("subscription"),
                "status": "expirada",
            }

        return {"evento": event_type, "processado": False}

    @staticmethod
    def cancelar_assinatura(gateway_subscription_id: str) -> bool:
        """Cancela a assinatura no Stripe (no final do ciclo atual)."""
        try:
            stripe.Subscription.modify(
                gateway_subscription_id, cancel_at_period_end=True
            )
            return True
        except stripe.error.StripeError as e:
            logger.error(f"Erro ao cancelar assinatura Stripe: {e}")
            return False

    @staticmethod
    def obter_status(gateway_subscription_id: str) -> str:
        """Obtém status da assinatura diretamente do Stripe."""
        try:
            sub = stripe.Subscription.retrieve(gateway_subscription_id)
            status_map = {
                "active": "ativa",
                "past_due": "pendente",
                "canceled": "cancelada",
                "incomplete": "pendente",
            }
            return status_map.get(sub.status, "expirada")
        except stripe.error.StripeError:
            return "expirada"