"""
Serviço de scraping do Plano de Contratações Anual (PCA) do PNCP.
Consome o endpoint /pca da API de consulta v1 do PNCP,
que retorna os planos de contratação anuais dos órgãos públicos.
Usa requests (síncrono) para compatibilidade com tasks Celery.
"""
import logging
import re
from datetime import datetime
from typing import Optional

import requests

from backend.app.core.database import mongo_db
from backend.app.models.nosql.pca_pncp import (
    ItemPCA,
    OrgaoEntidadePCA,
    PCAPNCP,
)

logger = logging.getLogger(__name__)

# Base URL da API de consulta do PNCP (v1)
BASE_URL = "https://pncp.gov.br/api/consulta/v1/"


# ─── Mapeamento de palavras-chave para CNAE ────────────────────────────────

_PALAVRAS_CNAE = {
    "computador": "9511-8/00",
    "impressora": "9511-8/00",
    "notebook": "9511-8/00",
    "software": "6201-5/00",
    "desenvolvimento": "6201-5/00",
    "sistema": "6202-3/00",
    "consultoria": "6202-3/00",
    "informática": "6203-1/00",
    "ti": "6203-1/00",
    "tecnologia": "6203-1/00",
    "material": "4741-5/00",
    "papelaria": "4741-5/00",
    "limpeza": "8121-4/00",
    "higiene": "8121-4/00",
    "vigilância": "8011-1/00",
    "segurança": "8011-1/00",
    "alimentação": "5620-1/00",
    "refeição": "5620-1/00",
    "café": "5620-1/00",
    "transporte": "4923-0/00",
    "logística": "4930-2/00",
    "veículo": "4930-2/00",
    "manutenção": "3314-7/00",
    "construção": "4120-4/00",
    "obra": "4120-4/00",
    "engenharia": "4120-4/00",
    "saúde": "8610-1/01",
    "medicamento": "8640-2/00",
    "hospital": "8610-1/01",
    "educação": "8520-1/00",
    "escola": "8520-1/00",
    "treinamento": "8533-3/00",
    "curso": "8533-3/00",
}


def _extrair_cnae(descricao: str) -> str:
    """
    Extrai CNAE a partir da descrição do item via palavras-chave.
    Retorna string vazia se não encontrar correspondência.
    """
    desc_lower = descricao.lower()
    for palavra, cnae in _PALAVRAS_CNAE.items():
        if palavra in desc_lower:
            return cnae
    return ""


def _parse_data(data_str) -> Optional[datetime]:
    """Tenta converter string ISO para datetime."""
    if not data_str:
        return None
    try:
        return datetime.fromisoformat(data_str.replace("Z", "+00:00"))
    except (ValueError, TypeError):
        return None


# ─── Consulta à API do PNCP ────────────────────────────────────────────────


def buscar_pca_por_ano(
    ano: int,
    pagina: int = 1,
    tamanho: int = 50,
) -> list[dict]:
    """
    Consulta a API pública do PNCP e retorna PCAs de um ano específico.

    Endpoint: GET /pca?ano={ano}&pagina={pagina}&tamanhoPagina={tamanho}

    Parâmetros:
        ano: Ano de referência do PCA.
        pagina: Número da página para paginação.
        tamanho: Itens por página (máximo 100).

    Retorna:
        Lista de dicionários com os dados dos PCAs.
    """
    params = {
        "ano": ano,
        "pagina": pagina,
        "tamanhoPagina": min(tamanho, 100),
    }

    try:
        response = requests.get(
            f"{BASE_URL}pca",
            params=params,
            timeout=60,
        )
        response.raise_for_status()
        data = response.json()

        # A API retorna um wrapper paginado { content: [], totalPages: ..., ... }
        if isinstance(data, dict):
            resultados = data.get("content", data.get("data", []))
            logger.info(
                f"PNCP PCA ano={ano}: página={pagina}, "
                f"total_paginas={data.get('totalPages', '?')}, "
                f"recebidos={len(resultados)}"
            )
        else:
            resultados = data if isinstance(data, list) else []
            logger.info(
                f"PNCP PCA ano={ano}: página={pagina}, recebidos={len(resultados)}"
            )

        return resultados

    except requests.RequestException as e:
        logger.error(f"Erro ao buscar PCA no PNCP (ano={ano}, página={pagina}): {e}")
        return []


def buscar_pca_por_uf(
    uf: str,
    ano: int,
    pagina: int = 1,
    tamanho: int = 50,
) -> list[dict]:
    """
    Consulta PCA por UF e ano.
    """
    params = {
        "uf": uf.upper(),
        "ano": ano,
        "pagina": pagina,
        "tamanhoPagina": min(tamanho, 100),
    }

    try:
        response = requests.get(
            f"{BASE_URL}pca",
            params=params,
            timeout=60,
        )
        response.raise_for_status()
        data = response.json()
        return data.get("content", []) if isinstance(data, dict) else []

    except requests.RequestException as e:
        logger.error(f"Erro ao buscar PCA UF={uf} ano={ano}: {e}")
        return []


# ─── Mapeamento e Persistência ─────────────────────────────────────────────


def _mapear_pca(item: dict) -> dict:
    """
    Mapeia os campos retornados pela API do PNCP para a estrutura
    do modelo PCAPNCP (snake_case).
    """
    orgao_raw = item.get("orgaoEntidade") or item.get("orgao") or {}
    orgao = OrgaoEntidadePCA(
        cnpj=orgao_raw.get("cnpj"),
        nome=orgao_raw.get("nome"),
        sigla=orgao_raw.get("sigla"),
        uf=orgao_raw.get("uf"),
        municipio=orgao_raw.get("municipio"),
    )

    # Mapeia itens do PCA
    itens_raw = item.get("itens", []) or []
    itens = []
    for i in itens_raw:
        descricao = i.get("descricao", i.get("descricaoItem", ""))
        itens.append(
            ItemPCA(
                codigo_classificacao=i.get("codigoClassificacao")
                or i.get("codigo_classificacao"),
                descricao_item=descricao,
                quantidade_estimada=i.get("quantidadeEstimada")
                or i.get("quantidade_estimada"),
                valor_estimado=i.get("valorEstimado") or i.get("valor_estimado"),
                valor_total_estimado=i.get("valorTotalEstimado")
                or i.get("valor_total_estimado"),
                unidade_medida=i.get("unidadeMedida") or i.get("unidade_medida"),
                cnae_identificado=_extrair_cnae(descricao),
            )
        )

    return {
        "id_pca": str(item.get("idPca", item.get("id_pca", ""))),
        "ano": int(item.get("ano", datetime.now().year)),
        "orgao_entidade": orgao.model_dump() if orgao else None,
        "itens": [i.model_dump() for i in itens],
        "data_publicacao": _parse_data(
            item.get("dataPublicacao") or item.get("data_publicacao")
        ),
        "data_importacao": datetime.utcnow(),
        "origem": "api_pncp",
        "versao": "1.0",
    }


def salvar_pca_mongodb(pcas: list[dict]) -> int:
    """
    Salva PCAs no MongoDB (coleção 'pca_pncp'),
    evitando duplicatas pelo id_pca via upsert.

    Retorna o número de documentos inseridos ou atualizados.
    """
    collection = mongo_db["pca_pncp"]
    processados = 0

    for item in pcas:
        dados = _mapear_pca(item)
        if not dados["id_pca"]:
            continue

        try:
            result = collection.replace_one(
                {"id_pca": dados["id_pca"]},
                dados,
                upsert=True,
            )
            if result.upserted_id or result.modified_count > 0:
                processados += 1
        except Exception as e:
            logger.debug(f"Erro ao salvar PCA {dados['id_pca']}: {e}")
            continue

    if processados > 0:
        logger.info(f"{processados} PCAs salvos/atualizados no MongoDB")

    return processados


# ─── Scraping Completo ─────────────────────────────────────────────────────


def scraper_pca_completo(
    anos: Optional[list[int]] = None,
    ufs: Optional[list[str]] = None,
) -> dict:
    """
    Executa o scraping completo de PCAs para os anos e UFs especificados.

    Itera sobre todas as páginas de cada ano/UF.

    Parâmetros:
        anos: Lista de anos para buscar. Se None, busca ano atual e anterior.
        ufs: Lista de UFs para filtrar. Se None, busca todas.

    Retorna:
        Dicionário com total processado e anos processados.
    """
    anos_alvo = anos or [datetime.now().year, datetime.now().year - 1]
    total_global = 0

    for ano in anos_alvo:
        pagina = 1
        while True:
            dados = buscar_pca_por_ano(ano=ano, pagina=pagina)
            if not dados:
                break

            processados = salvar_pca_mongodb(dados)
            total_global += processados

            # Se retornou menos itens que o tamanho da página, é a última
            if len(dados) < 50:
                break

            pagina += 1

    logger.info(
        f"Scraping PCA concluído. Total processado: {total_global}, "
        f"anos: {anos_alvo}"
    )
    return {"total": total_global, "anos": anos_alvo}