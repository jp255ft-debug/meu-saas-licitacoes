from pydantic import BaseModel, EmailStr


class LoginRequest(BaseModel):
    """Schema para login com email/CPF e senha via JSON body."""
    username: str
    password: str


class TokenResponse(BaseModel):
    """Schema de resposta para autenticação (login)."""
    access_token: str
    token_type: str = "bearer"


class TokenPayload(BaseModel):
    """Schema extraído do payload do JWT decodificado."""
    sub: str  # UUID do usuário
    exp: int  # timestamp de expiração