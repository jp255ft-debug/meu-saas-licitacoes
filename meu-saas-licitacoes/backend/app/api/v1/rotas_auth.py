from datetime import timedelta

from fastapi import APIRouter, Depends, HTTPException, Response, status
from sqlalchemy.orm import Session

from backend.app.core.config import settings
from backend.app.core.database import get_db
from backend.app.core.security import verify_password, create_access_token
from backend.app.models.sql.usuario import Usuario
from backend.app.schemas.token import LoginRequest, TokenResponse

router = APIRouter(prefix="/auth", tags=["Autenticação"])


@router.post("/token", response_model=TokenResponse)
def login(
    data: LoginRequest,
    response: Response,
    db: Session = Depends(get_db),
):
    """
    Autentica o usuário com email/CPF e senha.
    Retorna um token JWT (Bearer) no body da resposta **e** define
    um cookie httpOnly seguro para uso automático pelo frontend.

    Aceita tanto email quanto CPF/CNPJ como 'username'.
    """
    username = data.username
    password = data.password

    # Busca por email ou CPF/CNPJ
    usuario = (
        db.query(Usuario)
        .filter(
            (Usuario.email == username) | (Usuario.cpf_cnpj == username)
        )
        .first()
    )

    if usuario is None or not verify_password(password, usuario.senha_hash):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Credenciais inválidas",
        )

    if not usuario.ativo:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Usuário inativo",
        )

    token = create_access_token(subject=usuario.uuid)

    # Define cookie httpOnly (seguro contra XSS)
    # O cookie é enviado automaticamente nas requisições seguintes
    response.set_cookie(
        key="access_token",
        value=token,
        httponly=True,
        secure=(settings.ENVIRONMENT == "production"),  # HTTPS apenas em produção
        samesite="lax",
        max_age=settings.ACCESS_TOKEN_EXPIRE_MINUTES * 60,
        path="/",
    )

    return TokenResponse(access_token=token)


@router.post("/logout")
def logout(response: Response):
    """
    Invalida o cookie de autenticação (logout).
    Remove o cookie access_token da resposta.
    """
    response.delete_cookie(
        key="access_token",
        path="/",
    )
    return {"message": "Logout realizado com sucesso"}