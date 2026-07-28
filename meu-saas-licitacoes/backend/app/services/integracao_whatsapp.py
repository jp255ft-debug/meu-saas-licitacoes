"""
Serviço de integração com WhatsApp (Evolution API) para envio de alertas.
Responsável por notificar usuários sobre novos editais compatíveis,
certidões próximas do vencimento e contratações com propostas abertas.
Usa requests (síncrono) para compatibilidade com tasks Celery.
"""
import logging
import re
from datetime import datetime
from typing import Optional

import requests

from backend.app.core.config import settings

logger = logging.getLogger(__name__)


# ─── Helpers de validação ───────────────────────────────────────────────────


def _limpar_numero(telefone: str) -> str:
    """
    Remove caracteres não numéricos e garante prefixo 55 (Brasil).
    Retorna número limpo ou string vazia se inválido.
    """
    cleaned = re.sub(r"\D", "", telefone)
    if not cleaned:
        return ""

    # Se não começar com 55, adiciona código do Brasil
    if not cleaned.startswith("55"):
        # Remove 0 inicial se existir
        if cleaned.startswith("0"):
            cleaned = cleaned[1:]
        cleaned = "55" + cleaned

    return cleaned


def _validar_configuracao() -> bool:
    """Verifica se as variáveis de ambiente do WhatsApp estão configuradas."""
    if not settings.WHATSAPP_API_URL:
        logger.error("WHATSAPP_API_URL não configurada no .env")
        return False
    if not settings.WHATSAPP_API_KEY:
        logger.error("WHATSAPP_API_KEY não configurada no .env")
        return False
    if not settings.WHATSAPP_INSTANCE:
        logger.error("WHATSAPP_INSTANCE não configurada no .env")
        return False
    return True


# ─── Formatação de mensagens ───────────────────────────────────────────────


def formatar_mensagem_contratacao(contratacao: dict) -> str:
    """
    Formata a mensagem para envio por WhatsApp sobre uma nova contratação.
    Aceita dicionário no formato da API do PNCP (camelCase) ou do MongoDB (snake_case).
    """
    objeto = contratacao.get("objetoCompra", "Licitação")
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
    valor = contratacao.get("valorTotalEstimado") or contratacao.get("valor_total_estimado", 0)
    data_encerramento = (
        contratacao.get("dataEncerramentoProposta")
        or contratacao.get("data_encerramento_proposta")
        or ""
    )
    link = contratacao.get("linkEdital") or contratacao.get("link_edital", "")

    # Formata data de encerramento
    data_str = ""
    if data_encerramento:
        if isinstance(data_encerramento, str):
            try:
                data_encerramento = datetime.fromisoformat(
                    data_encerramento.replace("Z", "+00:00")
                )
            except (ValueError, TypeError):
                pass
        if isinstance(data_encerramento, datetime):
            data_str = data_encerramento.strftime("%d/%m/%Y")
        else:
            data_str = str(data_encerramento)

    msg = (
        f"🚨 NOVA LICITAÇÃO COM PROPOSTAS ABERTAS!\n\n"
        f"📌 {objeto}\n"
        f"🏛️ {orgao} - {uf}\n"
        f"💰 R$ {valor:,.2f}\n"
        f"📅 Encerramento: {data_str}\n\n"
        f"🔗 {link}\n\n"
        f"Acesse o portal para mais detalhes e envie sua proposta!"
    )
    return msg


# ─── Envio de mensagens via Evolution API ───────────────────────────────────


def _enviar_mensagem(telefone: str, mensagem: str) -> bool:
    """
    Envia mensagem via Evolution API.
    Usa as configurações do .env: WHATSAPP_API_URL, WHATSAPP_API_KEY, WHATSAPP_INSTANCE.

    Retorna True se enviado com sucesso, False caso contrário.
    """
    if not _validar_configuracao():
        return False

    numero_limpo = _limpar_numero(telefone)
    if not numero_limpo or len(numero_limpo) < 12:
        logger.error(f"Número de WhatsApp inválido: {telefone}")
        return False

    url = f"{settings.WHATSAPP_API_URL}/message/sendText/{settings.WHATSAPP_INSTANCE}"
    payload = {
        "number": numero_limpo,
        "text": mensagem,
        "delay": 1,
    }
    headers = {
        "apikey": settings.WHATSAPP_API_KEY,
        "Content-Type": "application/json",
    }

    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
        response.raise_for_status()
        logger.info(f"✅ WhatsApp enviado para {numero_limpo}")
        return True
    except requests.Timeout:
        logger.error("Timeout ao enviar WhatsApp")
        return False
    except requests.HTTPError as e:
        logger.error(f"Erro HTTP WhatsApp ({e.response.status_code}): {e.response.text}")
        return False
    except requests.RequestException as e:
        logger.error(f"Erro ao enviar WhatsApp: {e}")
        return False


# ─── Funções públicas de envio ──────────────────────────────────────────────


def enviar_alerta_edital(
    telefone: str,
    nome_usuario: str,
    orgao: str,
    objeto_resumido: str,
    url_edital: Optional[str] = None,
) -> bool:
    """
    Envia mensagem de WhatsApp para o usuário sobre um novo edital relevante.
    Retorna True se enviado com sucesso.
    """
    mensagem = (
        f"📢 *Novo Edital Encontrado!*\n\n"
        f"Olá {nome_usuario}, identificamos um edital compatível com seu perfil:\n\n"
        f"🏛 *Órgão:* {orgao}\n"
        f"📄 *Objeto:* {objeto_resumido[:200]}...\n"
    )

    if url_edital:
        mensagem += f"\n🔗 *Link:* {url_edital}"

    mensagem += "\n\nAcesse o painel para mais detalhes."

    return _enviar_mensagem(telefone, mensagem)


def enviar_alerta_certidao(
    telefone: str,
    nome_usuario: str,
    tipo_certidao: str,
    data_validade: str,
    dias_restantes: int,
) -> bool:
    """
    Envia alerta sobre certidão próxima do vencimento via WhatsApp.
    """
    mensagem = (
        f"⚠️ *Alerta de Certidão*\n\n"
        f"Olá {nome_usuario}, sua certidão *{tipo_certidao}* "
        f"vence em *{dias_restantes} dias* ({data_validade}).\n\n"
        f"Renove-a para continuar participando de licitações!"
    )

    return _enviar_mensagem(telefone, mensagem)


def enviar_alerta_whatsapp_contratacao(
    telefone: str,
    contratacao: dict,
) -> bool:
    """
    Envia alerta de WhatsApp sobre uma nova contratação com propostas abertas.
    Usa a função formatar_mensagem_contratacao() e envia via _enviar_mensagem().

    Retorna True se enviado com sucesso.
    """
    mensagem = formatar_mensagem_contratacao(contratacao)
    return _enviar_mensagem(telefone, mensagem)