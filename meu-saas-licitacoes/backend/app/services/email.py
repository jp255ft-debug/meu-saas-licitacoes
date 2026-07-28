"""
Serviço de envio de e-mails transacionais via SendGrid.
Responsável por enviar alertas de novas contratações aos usuários.
"""
import logging
from typing import Optional

import sendgrid
from sendgrid.helpers.mail import Mail

from backend.app.core.config import settings

logger = logging.getLogger(__name__)


def enviar_email_alerta(
    destinatario: str,
    assunto: str,
    mensagem: str,
) -> bool:
    """
    Envia e-mail via SendGrid.
    Retorna True se enviado com sucesso, False caso contrário.

    Se SENDGRID_API_KEY não estiver configurada, apenas loga o envio
    (útil para ambiente de desenvolvimento).
    """
    if not settings.SENDGRID_API_KEY:
        logger.warning(
            "SENDGRID_API_KEY não configurado. "
            f"E-mail não enviado para {destinatario}. "
            f"Assunto: {assunto}"
        )
        return False

    try:
        sg = sendgrid.SendGridAPIClient(api_key=settings.SENDGRID_API_KEY)
        mail = Mail(
            from_email=settings.EMAIL_FROM,
            to_emails=destinatario,
            subject=assunto,
            plain_text_content=mensagem,
        )
        response = sg.send(mail)

        if response.status_code == 202:
            logger.info(f"✅ E-mail enviado para {destinatario} - assunto: {assunto}")
            return True
        else:
            logger.error(
                f"Erro SendGrid: {response.status_code} - {response.body}"
            )
            return False

    except Exception as e:
        logger.error(f"Erro ao enviar e-mail para {destinatario}: {e}")
        return False


def formatar_email_contratacao(contratacao: dict) -> str:
    """
    Formata o corpo do e-mail de alerta de nova contratação.
    Aceita dicionário no formato da API do PNCP (camelCase) ou do MongoDB (snake_case).
    """
    objeto = (
        contratacao.get("objetoCompra")
        or contratacao.get("objeto_compra")
        or "Licitação"
    )
    orgao = (
        contratacao.get("orgaoEntidade", {}).get("nome")
        or contratacao.get("orgao_entidade", {}).get("nome")
        or "Órgão público"
    )
    uf = (
        contratacao.get("orgaoEntidade", {}).get("uf")
        or contratacao.get("orgao_entidade", {}).get("uf")
        or ""
    )
    valor = (
        contratacao.get("valorTotalEstimado")
        or contratacao.get("valor_total_estimado")
        or 0
    )
    link = (
        contratacao.get("linkEdital")
        or contratacao.get("link_edital")
        or ""
    )

    return f"""
Nova licitação com propostas abertas!

Objeto: {objeto}
Órgão: {orgao} - {uf}
Valor estimado: R$ {valor:,.2f}

Acesse o edital completo: {link}

--
Este é um alerta automático do Licitação SaaS.
"""


def formatar_assunto_contratacao(contratacao: dict) -> str:
    """Gera assunto para o e-mail de alerta."""
    objeto = (
        contratacao.get("objetoCompra")
        or contratacao.get("objeto_compra")
        or ""
    )
    return f"Nova licitação: {objeto[:80]}"