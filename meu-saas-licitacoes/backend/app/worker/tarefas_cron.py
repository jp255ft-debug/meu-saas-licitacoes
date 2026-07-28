"""
Tarefas agendadas do Celery para execução periódica.
Inclui scraping de editais, processamento de IA e verificação de certidões.
"""
import logging
from datetime import date, datetime, timedelta, timezone

from sqlalchemy.orm import Session

from backend.app.core.database import SessionLocal, mongo_db
from backend.app.models.sql.certidao import Certidao
from backend.app.services.scraper_pncp import buscar_editais, salvar_editais_mongodb
from backend.app.services.scraper_contratacoes import (
    buscar_contratacoes_propostas_abertas,
    buscar_novas_contratacoes,
    salvar_contratacoes_mongodb,
    scraper_contratacoes_completo,
)
from backend.app.services.integracao_openai import resumir_edital
from backend.app.services.integracao_whatsapp import (
    enviar_alerta_whatsapp_contratacao,
    formatar_mensagem_contratacao,
)
from backend.app.services.smart_match import usuarios_compatíveis_com_contratacao
from backend.app.services.email import (
    enviar_email_alerta,
    formatar_email_contratacao,
    formatar_assunto_contratacao,
)
from backend.app.services.scraper_pca import scraper_pca_completo
from backend.app.services.smart_match import buscar_pca_para_usuario
from backend.app.models.sql.notificacao import (
    Notificacao,
    TipoNotificacaoEnum,
    StatusNotificacaoEnum,
)
from backend.app.worker.celery_app import celery_app

logger = logging.getLogger(__name__)


@celery_app.task(name="scraping_diario")
def scraping_diario():
    """
    Tarefa diária: raspa editais do PNCP e salva no MongoDB.
    Executa para todas as UFs ou para UFs específicas.
    """
    logger.info("Iniciando scraping diário de editais...")
    total_inseridos = 0

    # Lista de UFs prioritárias (Sudeste + Sul)
    ufs = ["SP", "RJ", "MG", "ES", "PR", "SC", "RS"]

    for uf in ufs:
        editais = buscar_editais(uf=uf, pagina=1, tamanho=50)
        if editais:
            inseridos = salvar_editais_mongodb(editais)
            total_inseridos += inseridos
            logger.info(f"UF {uf}: {inseridos} novos editais")

    # Busca editais sem filtro de UF (nacionais)
    editais_nacionais = buscar_editais(pagina=1, tamanho=100)
    if editais_nacionais:
        inseridos = salvar_editais_mongodb(editais_nacionais)
        total_inseridos += inseridos

    logger.info(
        f"Scraping diário concluído. Total de novos editais: {total_inseridos}"
    )
    return total_inseridos


@celery_app.task(name="processar_editais_pendentes")
def processar_editais_pendentes():
    """
    Processa editais novos que ainda não têm resumo de IA.
    Busca no MongoDB documentos sem processamento_ia correspondente.
    """
    editais_collection = mongo_db["editais_pncp"]
    processamento_collection = mongo_db["processamento_ia"]

    # Encontra editais sem processamento
    editais_sem_resumo = editais_collection.aggregate([
        {
            "$lookup": {
                "from": "processamento_ia",
                "localField": "_id",
                "foreignField": "edital_mongodb_id",
                "as": "ia",
            }
        },
        {"$match": {"ia": {"$size": 0}}},
        {"$limit": 10},  # processa 10 por vez para não sobrecarregar
    ])

    processados = 0
    for edital in editais_sem_resumo:
        edital_id = str(edital["_id"])
        resultado = resumir_edital(edital_id)
        if resultado:
            processados += 1
            logger.info(f"Edital {edital_id} processado com IA")

    logger.info(f"Editais processados com IA: {processados}")
    return processados


@celery_app.task(name="verificar_certidoes_vencendo")
def verificar_certidoes_vencendo():
    """
    Tarefa diária: verifica certidões próximas do vencimento e dispara alertas.
    - ≤ 30 dias: alerta de 30 dias (uma vez)
    - ≤ 7 dias: alerta de 7 dias (uma vez)
    - Vencida: atualiza status automaticamente
    Usa flags booleanas (alerta_enviado_7d, alerta_enviado_30d) para evitar
    reenvio de alertas a cada execução.
    """
    db: Session = SessionLocal()
    try:
        hoje = date.today()
        limite_30d = hoje + timedelta(days=30)
        limite_7d = hoje + timedelta(days=7)

        # Busca certidões que vencem em até 30 dias e não estão vencidas
        certidoes = (
            db.query(Certidao)
            .filter(
                Certidao.data_validade <= limite_30d,
                Certidao.data_validade >= hoje,
            )
            .all()
        )

        logger.info(
            f"Verificando {len(certidoes)} certidões próximas do vencimento..."
        )

        alertas_disparados = 0
        for cert in certidoes:
            dias_restantes = (cert.data_validade - hoje).days
            status_alterado = False

            # Alerta de 30 dias
            if dias_restantes <= 30 and not cert.alerta_enviado_30d:
                logger.info(
                    f"Alerta 30d para certidão {cert.uuid} "
                    f"(tipo: {cert.tipo}, vence em {dias_restantes} dias)"
                )
                # TODO: Integrar com WhatsApp service
                # enviar_alerta_whatsapp.delay(...)
                cert.alerta_enviado_30d = True
                alertas_disparados += 1

            # Alerta de 7 dias
            if dias_restantes <= 7 and not cert.alerta_enviado_7d:
                logger.info(
                    f"Alerta 7d para certidão {cert.uuid} "
                    f"(tipo: {cert.tipo}, vence em {dias_restantes} dias)"
                )
                # TODO: Integrar com WhatsApp service
                # enviar_alerta_whatsapp.delay(...)
                cert.alerta_enviado_7d = True
                alertas_disparados += 1

            # Atualiza status
            if dias_restantes <= 30 and cert.status == "VALIDA":
                cert.status = "VENCENDO"
                status_alterado = True

        # Atualiza certidões que já venceram
        vencidas = (
            db.query(Certidao)
            .filter(
                Certidao.data_validade < hoje,
                Certidao.status.in_(["VALIDA", "VENCENDO"]),
            )
            .update({"status": "VENCIDA"}, synchronize_session=False)
        )

        db.commit()

        logger.info(
            f"Radar concluído. Alertas disparados: {alertas_disparados}. "
            f"Status alterados: {status_alterado}. Certidões vencidas: {vencidas}"
        )
        return {
            "alertas_disparados": alertas_disparados,
            "certidoes_vencidas": vencidas,
        }
    finally:
        db.close()


@celery_app.task(name="scraping_contratacoes_diario")
def scraping_contratacoes_diario():
    """
    Tarefa diária: raspa contratações com propostas abertas do PNCP
    e salva no MongoDB (coleção 'contratacoes_pncp').
    Executa para todas as UFs brasileiras.
    """
    logger.info("Iniciando scraping diário de contratações...")
    total = scraper_contratacoes_completo()
    logger.info(f"Scraping de contratações concluído. Total: {total}")
    return total


@celery_app.task(name="verificar_novas_contratacoes")
def verificar_novas_contratacoes():
    """
    Tarefa periódica (a cada 30 min): verifica NOVAS contratações com
    propostas abertas e dispara alertas para usuários compatíveis.

    Fluxo completo:
    1. Busca contratações novas que ainda não estão no banco
    2. Para cada nova contratação, encontra usuários compatíveis (smart match)
    3. Envia WhatsApp e/ou e-mail para cada usuário
    4. Registra as notificações na tabela 'notificacoes'
    """
    logger.info("Iniciando verificação de novas contratações...")
    db: Session = SessionLocal()
    ufs = ["SP", "RJ", "MG", "ES", "PR", "SC", "RS"]

    try:
        # 1. Buscar novas contratações via PNCP
        novas_contratacoes = buscar_novas_contratacoes(ufs=ufs)
        if not novas_contratacoes:
            logger.info("Nenhuma nova contratação encontrada.")
            return 0

        logger.info(f"Encontradas {len(novas_contratacoes)} novas contratações.")
        total_alertas = 0

        for contratacao in novas_contratacoes:
            contratacao_id = contratacao.get("_id", "")
            if not contratacao_id:
                continue

            # 2. Buscar usuários compatíveis
            usuarios = usuarios_compatíveis_com_contratacao(db, contratacao)
            logger.info(
                f"Contratação {contratacao_id}: {len(usuarios)} usuários compatíveis"
            )

            for usuario in usuarios:
                # Verificar se já foi notificado para esta contratação
                ja_notificado = (
                    db.query(Notificacao)
                    .filter(
                        Notificacao.usuario_id == usuario.id,
                        Notificacao.contratacao_id == contratacao_id,
                    )
                    .first()
                )
                if ja_notificado:
                    continue

                # 3. Enviar WhatsApp
                if usuario.alertas_whatsapp and usuario.telefone_whatsapp:
                    try:
                        sucesso_whats = enviar_alerta_whatsapp_contratacao(
                            usuario.telefone_whatsapp, contratacao
                        )
                        status_whats = (
                            StatusNotificacaoEnum.ENVIADO
                            if sucesso_whats
                            else StatusNotificacaoEnum.FALHA
                        )
                        erro_whats = None if sucesso_whats else "Falha no envio WhatsApp"
                    except Exception as e:
                        status_whats = StatusNotificacaoEnum.FALHA
                        erro_whats = str(e)
                        logger.error(f"Erro WhatsApp para {usuario.id}: {e}")

                    # Registrar notificação WhatsApp
                    import uuid

                    notif = Notificacao(
                        uuid=str(uuid.uuid4()),
                        usuario_id=usuario.id,
                        contratacao_id=contratacao_id,
                        tipo=TipoNotificacaoEnum.WHATSAPP,
                        status=status_whats,
                        mensagem=formatar_mensagem_contratacao(contratacao),
                        data_envio=datetime.utcnow()
                        if status_whats == StatusNotificacaoEnum.ENVIADO
                        else None,
                        erro=erro_whats,
                    )
                    db.add(notif)
                    total_alertas += 1

                # 4. Enviar E-mail
                if usuario.alertas_email and usuario.email:
                    assunto = formatar_assunto_contratacao(contratacao)
                    corpo = formatar_email_contratacao(contratacao)
                    sucesso_email = enviar_email_alerta(usuario.email, assunto, corpo)
                    status_email = (
                        StatusNotificacaoEnum.ENVIADO
                        if sucesso_email
                        else StatusNotificacaoEnum.FALHA
                    )

                    import uuid

                    notif = Notificacao(
                        uuid=str(uuid.uuid4()),
                        usuario_id=usuario.id,
                        contratacao_id=contratacao_id,
                        tipo=TipoNotificacaoEnum.EMAIL,
                        status=status_email,
                        mensagem=corpo,
                        data_envio=datetime.utcnow()
                        if sucesso_email
                        else None,
                        erro=None if sucesso_email else "Falha no envio de e-mail",
                    )
                    db.add(notif)
                    total_alertas += 1

                db.commit()

        logger.info(
            f"Verificação concluída. {total_alertas} alertas disparados "
            f"para {len(novas_contratacoes)} novas contratações."
        )
        return total_alertas

    except Exception as e:
        logger.exception(f"Erro ao verificar novas contratações: {e}")
        db.rollback()
        return 0
    finally:
        db.close()


# ─── PCA (Plano de Contratações Anual) ──────────────────────────────────────


@celery_app.task(name="scraping_pca_diario")
def scraping_pca_diario():
    """
    Tarefa diária: raspa PCAs do PNCP e salva no MongoDB.
    Busca ano atual e ano anterior.
    """
    logger.info("Iniciando scraping diário de PCA...")
    resultado = scraper_pca_completo()
    logger.info(f"Scraping PCA concluído: {resultado}")
    return resultado
