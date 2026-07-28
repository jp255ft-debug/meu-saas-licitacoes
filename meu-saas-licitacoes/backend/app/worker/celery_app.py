"""
Configuração do Celery para processamento de tarefas assíncronas.
Broker: Redis.
"""
from celery import Celery
from celery.schedules import crontab

from backend.app.core.config import settings

celery_app = Celery(
    "licitacao_saas",
    broker=settings.REDIS_URL,
    backend=settings.REDIS_URL,
    include=[
        "backend.app.worker.tarefas_cron",
    ],
)

# Configurações opcionais do Celery
celery_app.conf.update(
    task_serializer="json",
    accept_content=["json"],
    result_serializer="json",
    timezone="America/Sao_Paulo",
    enable_utc=True,
    task_track_started=True,
    task_soft_time_limit=300,  # 5 minutos
    task_time_limit=600,        # 10 minutos
)

# ─── Agendamento (Celery Beat) ──────────────────────────────────────────────
celery_app.conf.beat_schedule = {
    "scraping-diario": {
        "task": "scraping_diario",
        "schedule": crontab(hour=6, minute=0),  # 6h da manhã
    },
    "processar-editais-pendentes": {
        "task": "processar_editais_pendentes",
        "schedule": crontab(hour=7, minute=30),  # 7h30
    },
    "verificar-certidoes-diario": {
        "task": "verificar_certidoes_vencendo",
        "schedule": crontab(hour=8, minute=0),  # 8h da manhã
    },
    "scraping-contratacoes-diario": {
        "task": "scraping_contratacoes_diario",
        "schedule": crontab(hour=6, minute=30),  # 6h30 (após scraper de editais)
    },
    "verificar-novas-contratacoes": {
        "task": "verificar_novas_contratacoes",
        "schedule": crontab(minute="*/30"),  # A cada 30 minutos
    },
    "scraping-pca-diario": {
        "task": "scraping_pca_diario",
        "schedule": crontab(hour=5, minute=0),  # 5h da manhã
    },
}
