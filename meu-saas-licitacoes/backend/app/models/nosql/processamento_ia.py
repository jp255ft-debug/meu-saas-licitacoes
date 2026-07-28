"""
Schema/documento para a coleção 'processamento_ia' no MongoDB.
Armazena resultados de processamento de IA sobre editais (resumos, scores, embeddings).
"""
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class ProcessamentoIA(BaseModel):
    """
    Resultado do processamento de IA para um edital específico.
    Referencia o edital via edital_mongodb_id.
    """
    id: Optional[str] = Field(None, alias="_id")
    edital_mongodb_id: str = Field(
        ..., description="ID do edital no MongoDB (referência à coleção editais_pncp)"
    )
    resumo: str = Field(..., description="Resumo do edital em 5 tópicos gerado pela IA")
    score_relevancia: float = Field(
        ..., ge=0.0, le=1.0, description="Score de relevância para o perfil do usuário (0 a 1)"
    )
    categorias: list[str] = Field(
        default_factory=list, description="Categorias identificadas pela IA"
    )
    palavras_chave: list[str] = Field(
        default_factory=list, description="Palavras-chave extraídas"
    )
    embedding: Optional[list[float]] = Field(
        None, description="Vetor de embedding gerado pela OpenAI"
    )
    modelo_ia: str = Field(
        default="gpt-4o-mini", description="Modelo de IA utilizado no processamento"
    )
    data_processamento: datetime = Field(default_factory=datetime.utcnow)
    tokens_utilizados: Optional[int] = Field(None, description="Total de tokens consumidos")

    class Config:
        populate_by_name = True
        json_encoders = {
            datetime: lambda v: v.isoformat(),
        }