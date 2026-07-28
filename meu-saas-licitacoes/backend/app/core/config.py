from pydantic_settings import BaseSettings, SettingsConfigDict
from pathlib import Path


class Settings(BaseSettings):
    """
    Configurações carregadas do arquivo .env.
    Usa pydantic-settings para validação automática.
    SECRET_KEY é **obrigatório** via .env — sem ele a aplicação não inicia.
    """
    model_config = SettingsConfigDict(
        env_file=str(Path(__file__).resolve().parent.parent.parent / ".env"),
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    # Ambiente (development | production)
    ENVIRONMENT: str = "development"

    # PostgreSQL (via PgBouncer)
    DATABASE_URL: str = (
        "postgresql://licitacao_user:licitacao_pass@pgbouncer:6432/licitacao_db"
    )

    # MongoDB
    MONGODB_URL: str = (
        "mongodb://admin:admin_pass@mongodb:27017/licitacao_db?authSource=admin"
    )

    # Redis
    REDIS_URL: str = "redis://:redis_pass@redis:6379/0"

    # JWT — SECRET_KEY é obrigatório (sem valor default, falha se não configurado)
    SECRET_KEY: str  # type: ignore[assignment]
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 60 * 24 * 7  # 7 dias

    # CORS — origens permitidas (frontend)
    CORS_ORIGINS: list[str] = [
        "http://localhost:3000",
        "http://127.0.0.1:3000",
    ]

    # Stripe
    STRIPE_SECRET_KEY: str = ""
    STRIPE_PUBLISHABLE_KEY: str = ""
    STRIPE_WEBHOOK_SECRET: str = ""

    # WhatsApp (Evolution API)
    WHATSAPP_API_URL: str = ""
    WHATSAPP_API_KEY: str = ""
    WHATSAPP_INSTANCE: str = "licitacao"

    # E-mail (SendGrid)
    SENDGRID_API_KEY: str = ""
    EMAIL_FROM: str = "alertas@meusaas.com"

    # ─── LLM / IA — Provedor configurável ────────────────────────────────────
    # Valores: "openai" | "deepseek" | "gemini"
    LLM_PROVIDER: str = "openai"

    # OpenAI
    OPENAI_API_KEY: str = ""
    OPENAI_MODEL: str = "gpt-4o-mini"

    # DeepSeek (API compatível com OpenAI, ~5x mais barato)
    DEEPSEEK_API_KEY: str = ""
    DEEPSEEK_MODEL: str = "deepseek-chat"

    # Google Gemini
    GEMINI_API_KEY: str = ""
    GEMINI_MODEL: str = "gemini-1.5-flash"


settings = Settings()