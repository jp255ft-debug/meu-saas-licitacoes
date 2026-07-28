from fastapi import Depends, HTTPException, Request, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from sqlalchemy.orm import Session

from backend.app.core.database import get_db, get_mongo_db
from backend.app.core.security import decode_access_token
from backend.app.models.sql.assinatura import Assinatura
from backend.app.models.sql.usuario import Usuario

security_scheme = HTTPBearer(auto_error=False)


def get_current_user(
    request: Request,
    credentials: HTTPAuthorizationCredentials = Depends(security_scheme),
    db: Session = Depends(get_db),
) -> Usuario:
    """
    Valida o token JWT e retorna o usuário autenticado.

    Estratégia de leitura do token (fallback):
    1. Tenta extrair do header `Authorization: Bearer <token>` (compatível com Swagger)
    2. Se não encontrado, tenta ler do cookie `access_token` (usado pelo frontend)

    Lança HTTPException 401 se inválido ou ausente.
    """
    token = None

    # 1. Tenta do header Authorization (Bearer)
    if credentials is not None:
        token = credentials.credentials

    # 2. Fallback: tenta do cookie access_token
    if token is None:
        token = request.cookies.get("access_token")

    if token is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token de acesso ausente",
            headers={"WWW-Authenticate": "Bearer"},
        )

    subject = decode_access_token(token)
    if subject is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token inválido ou expirado",
            headers={"WWW-Authenticate": "Bearer"},
        )

    usuario = db.query(Usuario).filter(Usuario.uuid == subject).first()
    if usuario is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Usuário não encontrado",
        )

    if not usuario.ativo:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Usuário inativo",
        )

    return usuario


def get_usuario_assinatura_ativa(
    usuario: Usuario = Depends(get_current_user),
    db: Session = Depends(get_db),
) -> Assinatura:
    """
    Dependency que verifica se o usuário tem uma assinatura ativa.
    Útil para proteger rotas exclusivas de planos pagos.
    Retorna HTTP 402 se não houver assinatura ativa.
    """
    assinatura = (
        db.query(Assinatura)
        .filter(
            Assinatura.usuario_id == usuario.id,
            Assinatura.status == "ativa",
        )
        .first()
    )

    if not assinatura:
        raise HTTPException(
            status_code=status.HTTP_402_PAYMENT_REQUIRED,
            detail="Esta funcionalidade requer um plano ativo",
        )
    return assinatura