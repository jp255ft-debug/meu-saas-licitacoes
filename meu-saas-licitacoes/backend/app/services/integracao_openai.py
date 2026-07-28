"""
Serviço de integração com LLM (Large Language Model) para processamento de editais.
Suporta múltiplos provedores configuráveis via settings.LLM_PROVIDER:

- "openai"    → GPT-4o-mini (padrão)  — custo ~$0.60/1M tokens
- "deepseek"  → DeepSeek Chat         — custo ~$0.14/1M tokens (recomendado)
- "gemini"    → Gemini 1.5 Flash       — custo ~$0.35/1M tokens

Função principal: resumir_edital(edital_id) — assinatura mantida para compatibilidade.
"""
import logging
from typing import Optional

from openai import OpenAI
from pymongo.collection import Collection

from backend.app.core.config import settings
from backend.app.core.database import mongo_db
from backend.app.models.nosql.processamento_ia import ProcessamentoIA

logger = logging.getLogger(__name__)


# ─── Factory de clientes LLM ───────────────────────────────────────────────


def _criar_cliente() -> OpenAI:
    """
    Cria e retorna um cliente OpenAI (ou compatível) baseado no provedor
    configurado em settings.LLM_PROVIDER.

    DeepSeek é compatível com a API OpenAI, basta alterar base_url e api_key.
    """
    provider = settings.LLM_PROVIDER.lower()

    if provider == "deepseek":
        return OpenAI(
            api_key=settings.DEEPSEEK_API_KEY,
            base_url="https://api.deepseek.com/v1",
        )
    elif provider == "gemini":
        # Gemini via OpenAI-compatible proxy (Google Generative AI)
        # Nota: Gemini nativo requer google-generativeai, mas mantemos
        # compatibilidade via API key padrão da OpenAI para simplificar
        logger.warning(
            "Gemini provider requer configuração adicional. "
            "Usando fallback para OpenAI."
        )
        return OpenAI(api_key=settings.OPENAI_API_KEY)
    else:
        # OpenAI (padrão)
        return OpenAI(api_key=settings.OPENAI_API_KEY)


def _obter_modelo() -> str:
    """Retorna o nome do modelo baseado no provedor configurado."""
    provider = settings.LLM_PROVIDER.lower()
    if provider == "deepseek":
        return settings.DEEPSEEK_MODEL
    elif provider == "gemini":
        return settings.GEMINI_MODEL
    return settings.OPENAI_MODEL


# ─── Função principal (assinatura compatível) ──────────────────────────────


def resumir_edital(edital_id: str) -> Optional[ProcessamentoIA]:
    """
    Gera resumo de 5 tópicos para um edital usando o LLM configurado.
    Busca o edital no MongoDB, envia para a API e salva o resultado.

    Mantém a assinatura original: recebe edital_id (string do MongoDB).

    Parâmetros:
        edital_id: ID do edital no MongoDB (ObjectId como string).

    Retorna:
        ProcessamentoIA salvo ou None em caso de erro.
    """
    collection: Collection = mongo_db["editais_pncp"]
    from bson.objectid import ObjectId
    try:
        edital = collection.find_one({"_id": ObjectId(edital_id)})
    except Exception:
        edital = collection.find_one({"_id": edital_id})

    if not edital:
        logger.warning(f"Edital {edital_id} não encontrado no MongoDB")
        return None

    objeto = edital.get("objeto", "")
    orgao = edital.get("orgao_nome", "")
    modalidade = edital.get("modalidade", "")
    valor = edital.get("valor_estimado")
    uf = edital.get("uf", "")

    # Constrói prompt completo
    prompt = (
        f"Resuma o seguinte edital de licitação em exatamente 5 tópicos "
        f"objetivos e diretos (máximo 2 linhas cada):\n\n"
        f"Órgão: {orgao} - {uf}\n"
        f"Modalidade: {modalidade}\n"
        f"Objeto: {objeto}\n"
    )
    if valor:
        prompt += f"Valor estimado: R$ {valor:,.2f}\n"

    modelo = _obter_modelo()
    provider = settings.LLM_PROVIDER.lower()

    try:
        client = _criar_cliente()
        response = client.chat.completions.create(
            model=modelo,
            messages=[
                {
                    "role": "system",
                    "content": "Você é um especialista em licitações públicas brasileiras. "
                    "Seus resumos são precisos, concisos e voltados para pequenas empresas.",
                },
                {"role": "user", "content": prompt},
            ],
            max_tokens=500,
            temperature=0.3,
        )

        resumo_texto = response.choices[0].message.content or ""

        # Salva no MongoDB
        processamento = ProcessamentoIA(
            edital_mongodb_id=str(edital["_id"]),
            resumo=resumo_texto,
            score_relevancia=0.5,  # score inicial; smart_match recalcula depois
            categorias=_extrair_categorias(objeto),
            modelo_ia=f"{provider}/{modelo}",
            tokens_utilizados=response.usage.total_tokens if response.usage else None,
        )

        proc_collection: Collection = mongo_db["processamento_ia"]
        proc_collection.insert_one(
            processamento.model_dump(exclude={"id"})
        )

        logger.info(
            f"Resumo gerado para edital {edital_id} "
            f"(provedor={provider}, modelo={modelo})"
        )
        return processamento

    except Exception as e:
        logger.error(
            f"Erro ao chamar LLM ({provider}/{modelo}) para edital {edital_id}: {e}"
        )
        return None


def _extrair_categorias(objeto: str) -> list[str]:
    """
    Extrai categorias básicas do objeto do edital por palavras-chave.
    Placeholder até termos classificação por IA.
    """
    objeto_lower = objeto.lower()
    categorias = []

    palavra_categoria = {
        "obra": "OBRAS",
        "construção": "OBRAS",
        "engenharia": "OBRAS",
        "saúde": "SAUDE",
        "hospital": "SAUDE",
        "medicamento": "SAUDE",
        "educação": "EDUCACAO",
        "escola": "EDUCACAO",
        "informática": "TI",
        "software": "TI",
        "tecnologia": "TI",
        "transporte": "TRANSPORTE",
        "veículo": "TRANSPORTE",
        "alimentação": "ALIMENTACAO",
        "gênero alimentício": "ALIMENTACAO",
        "limpeza": "LIMPEZA",
        "higiene": "LIMPEZA",
    }

    for palavra, categoria in palavra_categoria.items():
        if palavra in objeto_lower and categoria not in categorias:
            categorias.append(categoria)

    return categorias if categorias else ["GERAL"]