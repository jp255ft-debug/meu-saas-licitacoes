from datetime import datetime, timedelta, timezone
from typing import Optional

from jose import JWTError, jwt
from passlib.context import CryptContext

from backend.app.core.config import settings

# ─── Hashing de senhas (bcrypt) ─────────────────────────────────────────────
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


def hash_password(password: str) -> str:
    """
    Gera o hash bcrypt da senha com 12 rounds de salt.
    """
    return pwd_context.hash(password)


def verify_password(plain_password: str, hashed_password: str) -> bool:
    """
    Compara senha em texto plano com o hash armazenado.
    """
    return pwd_context.verify(plain_password, hashed_password)


# ─── JWT ────────────────────────────────────────────────────────────────────


def create_access_token(
    subject: str,
    expires_delta: Optional[timedelta] = None,
) -> str:
    """
    Cria um token JWT com o UUID do usuário como subject ('sub').
    Por padrão, expira em ACCESS_TOKEN_EXPIRE_MINUTES (config).
    """
    if expires_delta is not None:
        expire = datetime.now(timezone.utc) + expires_delta
    else:
        expire = datetime.now(timezone.utc) + timedelta(
            minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES
        )

    to_encode = {
        "sub": subject,
        "exp": expire,
        "iat": datetime.now(timezone.utc),
    }
    encoded_jwt = jwt.encode(
        to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM
    )
    return encoded_jwt


def decode_access_token(token: str) -> Optional[str]:
    """
    Decodifica o token JWT e retorna o subject (UUID do usuário).
    Retorna None se o token for inválido ou expirado.
    """
    try:
        payload = jwt.decode(
            token, settings.SECRET_KEY, algorithms=[settings.ALGORITHM]
        )
        subject: str = payload.get("sub")
        if subject is None:
            return None
        return subject
    except JWTError:
        return None