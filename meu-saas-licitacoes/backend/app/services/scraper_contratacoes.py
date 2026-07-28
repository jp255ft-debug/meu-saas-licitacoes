"""
Serviço de raspagem de contratações do Portal Nacional de Compras Públicas (PNCP).
Consome o endpoint /contratacoes/proposta da API de consulta v1 do PNCP,
que retorna licitações com recebimento de propostas aberto.
"""
import logging
from datetime import datetime
from typing import Optional

import requests
from pymongo.collection import Collection
from pymongo.errors import DuplicateKeyError

from backend.app.core.database import mongo_db
from backend.app.models.nosql.contratacao_pncp import ContratacaoPNCP, OrgaoEntidade

logger = logging.getLogger(__name__)

# Base URL da API de consulta do PNCP (v1)
BASE_URL = "https://pncp.gov.br/api/consulta/v1/"

# Lista completa de UFs brasileiras
UFS = [
    "AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO",
    "MA", "MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI",
    "RJ", "RN", "RS", "RO", "RR", "SC", "SP", "SE", "TO",
]


def buscar_contratacoes_propostas_abertas(
    uf: Optional[str] = None,
    pagina: int = 1,
    tamanho: int = 50,
) -> list[dict]:
    """
    Consulta a API pública do PNCP e retorna contratações com
    recebimento de propostas aberto.

    Endpoint: GET /contratacoes/proposta

    Parâmetros:
        uf: Sigla da UF (ex: "SP", "RJ"). None para todas.
        pagina: Número da página para paginação.
        tamanho: Itens por página (máximo 100).

    Retorna:
        Lista de dicionários com os dados das contratações.
    """
    params = {
        "pagina": pagina,
        "tamanhoPagina": min(tamanho, 100),
    }
    if uf:
        params["uf"] = uf.upper()

    try:
        response = requests.get(
            f"{BASE_URL}contratacoes/proposta",
            params=params,
            timeout=30,
        )
        response.raise_for_status()
        data = response.json()

        # A API pode retornar um wrapper paginado { content: [], ... }
        # ou uma lista direta
        if isinstance(data, dict):
            resultados = data.get("content", data.get("data", []))
            logger.info(
                f"PNCP contratações UF={uf or 'TODAS'}: "
                f"página={pagina}, total_paginas={data.get('totalPages', '?')}, "
                f"recebidos={len(resultados)}"
            )
        else:
            resultados = data
            logger.info(
                f"PNCP contratações UF={uf or 'TODAS'}: página={pagina}, "
                f"recebidos={len(resultados)}"
            )

        return resultados

    except requests.RequestException as e:
        logger.error(f"Erro ao buscar contratações no PNCP: {e}")
        return []


def _parse_data(data_str) -> Optional[datetime]:
    """Tenta converter string ISO para datetime."""
    if not data_str:
        return None
    try:
        return datetime.fromisoformat(data_str.replace("Z", "+00:00"))
    except (ValueError, TypeError):
        return None


def _mapear_dados_contratacao(item: dict) -> dict:
    """
    Mapeia os campos retornados pela API do PNCP para a estrutura
    do modelo ContratacaoPNCP (snake_case).
    """
    orgao_raw = item.get("orgaoEntidade") or item.get("orgao") or {}

    orgao = OrgaoEntidade(
        cnpj=orgao_raw.get("cnpj"),
        nome=orgao_raw.get("nome"),
        sigla=orgao_raw.get("sigla"),
        uf=orgao_raw.get("uf"),
        municipio=orgao_raw.get("municipio"),
    )

    return {
        "numero_controle_pncp": str(item.get("numeroControlePNCP", "")),
        "ano_compra": int(item.get("anoCompra", datetime.now().year)),
        "sequencial_compra": item.get("sequencialCompra"),
        "orgao_entidade": orgao.model_dump() if orgao else None,
        "modalidade_nome": item.get("modalidadeNome", ""),
        "objeto_compra": item.get("objetoCompra", ""),
        "valor_total_estimado": item.get("valorTotalEstimado"),
        "data_publicacao_pncp": _parse_data(item.get("dataPublicacaoPncp")),
        "data_abertura_proposta": _parse_data(item.get("dataAberturaProposta")),
        "data_encerramento_proposta": _parse_data(item.get("dataEncerramentoProposta")),
        "link_edital": item.get("linkEdital") or item.get("linkSistemaOrigem"),
        "exclusividade_me_epp": item.get("exclusividadeMeEpp", False),
        "cnaes_identificados": item.get("cnaesIdentificados", []),
        "dados_brutos": item,
        "data_importacao": datetime.utcnow(),
        "status_proposta": "aberta",
    }


def salvar_contratacoes_mongodb(contratacoes: list[dict]) -> int:
    """
    Salva contratações no MongoDB (coleção 'contratacoes_pncp'),
    evitando duplicatas pelo numero_controle_pncp via upsert.

    Retorna o número de documentos inseridos ou atualizados.

    NOTA: O índice único "numero_controle_pncp" é criado na inicialização
    do app (via app/main.py ou database.py), não aqui — evita chamadas
    repetidas de create_index em cada scraping.
    """
    collection: Collection = mongo_db["contratacoes_pncp"]

    processados = 0
    for item in contratacoes:
        numero_controle = item.get("numeroControlePNCP") or item.get("numero_controle_pncp")
        if not numero_controle:
            continue

        dados = _mapear_dados_contratacao(item)

        try:
            result = collection.replace_one(
                {"numero_controle_pncp": dados["numero_controle_pncp"]},
                dados,
                upsert=True,
            )
            if result.upserted_id or result.modified_count > 0:
                processados += 1
        except DuplicateKeyError:
            # Se houver corrida concorrente, apenas ignora
            logger.debug(f"Duplicata ignorada: {numero_controle}")
            continue

    if processados > 0:
        logger.info(f"{processados} contratações salvas/atualizadas no MongoDB")

    return processados


def scraper_contratacoes_completo(ufs: Optional[list[str]] = None) -> int:
    """
    Executa o scraping completo de contratações com propostas abertas
    para todas as UFs (ou UFs especificadas).

    Itera sobre todas as páginas de cada UF.

    Parâmetros:
        ufs: Lista de UFs para buscar. Se None, busca todas.

    Retorna:
        Total de contratações processadas em toda a execução.
    """
    ufs_alvo = ufs or UFS
    total_global = 0

    for uf in ufs_alvo:
        pagina = 1
        while True:
            dados = buscar_contratacoes_propostas_abertas(uf=uf, pagina=pagina)
            if not dados:
                break

            processados = salvar_contratacoes_mongodb(dados)
            total_global += processados

            # Se retornou menos itens que o tamanho da página, é a última
            if len(dados) < 50:
                break

            pagina += 1

    logger.info(
        f"Scraping de contratações concluído. "
        f"Total processado: {total_global}"
    )
    return total_global


def buscar_novas_contratacoes(
    ufs: Optional[list[str]] = None,
    ultima_execucao: Optional[datetime] = None,
) -> int:
    """
    Busca APENAS as contratações novas (não existentes no banco)
    desde a última execução. Usa paginação para encontrar novidades.

    Parâmetros:
        ufs: Lista de UFs alvo.
        ultima_execucao: Timestamp da última execução bem-sucedida.

    Retorna:
        Número de contratações novas salvas.
    """
    ufs_alvo = ufs or UFS
    total_novas = 0

    for uf in ufs_alvo:
        dados = buscar_contratacoes_propostas_abertas(uf=uf, pagina=1, tamanho=100)
        if not dados:
            continue

        # Filtra apenas as que são realmente novas (não estão no banco)
        collection: Collection = mongo_db["contratacoes_pncp"]
        novas = []
        for item in dados:
            numero_controle = item.get("numeroControlePNCP")
            if numero_controle:
                existente = collection.find_one(
                    {"numero_controle_pncp": str(numero_controle)}
                )
                if not existente:
                    novas.append(item)

        if novas:
            processados = salvar_contratacoes_mongodb(novas)
            total_novas += processados
            logger.info(
                f"UF {uf}: {len(novas)} novas contratações de "
                f"{len(dados)} recebidas"
            )

    logger.info(f"Novas contratações encontradas: {total_novas}")
    return total_novas