"""
Serviço de scraping do Portal Nacional de Compras Públicas (PNCP).
Responsável por buscar editais publicados e armazená-los no MongoDB.
"""
import logging
from datetime import datetime
from typing import Optional

import requests
from pymongo.collection import Collection

from backend.app.core.database import mongo_db
from backend.app.models.nosql.edital_pncp import EditalPNCP

logger = logging.getLogger(__name__)

BASE_URL = "https://pncp.gov.br/api/consulta/v1/"


def buscar_editais(
    uf: Optional[str] = None,
    palavra_chave: Optional[str] = None,
    pagina: int = 1,
    tamanho: int = 50,
) -> list[dict]:
    """
    Consulta a API pública do PNCP e retorna lista de editais.
    Parâmetros:
        uf: Sigla da UF (ex: "SP", "RJ"). None para todas.
        palavra_chave: Termo para busca textual.
        pagina: Número da página para paginação.
        tamanho: Itens por página (máximo 100).
    """
    params = {
        "pagina": pagina,
        "tamanhoPagina": min(tamanho, 100),
    }
    if uf:
        params["uf"] = uf.upper()
    if palavra_chave:
        params["palavraChave"] = palavra_chave

    try:
        response = requests.get(
            f"{BASE_URL}editais",
            params=params,
            timeout=30,
        )
        response.raise_for_status()
        data = response.json()
        logger.info(
            f"PNCP: {len(data)} editais encontrados (página {pagina})"
        )
        return data if isinstance(data, list) else data.get("data", [])
    except requests.RequestException as e:
        logger.error(f"Erro ao buscar editais no PNCP: {e}")
        return []


def salvar_editais_mongodb(editais: list[dict]) -> int:
    """
    Salva editais no MongoDB, evitando duplicatas pelo pncp_id.
    Retorna o número de novos documentos inseridos.
    """
    collection: Collection = mongo_db["editais_pncp"]
    inseridos = 0

    for edital_data in editais:
        pncp_id = edital_data.get("id") or edital_data.get("pncpId")
        if not pncp_id:
            continue

        # Evita duplicata
        if collection.find_one({"pncp_id": str(pncp_id)}):
            continue

        documento = EditalPNCP(
            pncp_id=str(pncp_id),
            orgao_nome=edital_data.get("orgaoNome", ""),
            orgao_cnpj=edital_data.get("orgaoCnpj", ""),
            uf=edital_data.get("uf", ""),
            cidade=edital_data.get("cidade"),
            modalidade=edital_data.get("modalidade", ""),
            numero=edital_data.get("numero", ""),
            ano=int(edital_data.get("ano", datetime.now().year)),
            objeto=edital_data.get("objeto", ""),
            valor_estimado=edital_data.get("valorEstimado"),
            data_publicacao=_parse_data(edital_data.get("dataPublicacao")),
            data_abertura=_parse_data(edital_data.get("dataAbertura")),
            url_pncp=edital_data.get("url"),
            documento_url=edital_data.get("documentoUrl"),
            dados_brutos=edital_data,
        )

        collection.insert_one(documento.model_dump(exclude={"id"}))
        inseridos += 1

    if inseridos > 0:
        logger.info(f"{inseridos} novos editais salvos no MongoDB")

    return inseridos


def _parse_data(data_str) -> Optional[datetime]:
    """Tenta converter string ISO para datetime."""
    if not data_str:
        return None
    try:
        return datetime.fromisoformat(data_str.replace("Z", "+00:00"))
    except (ValueError, TypeError):
        return None