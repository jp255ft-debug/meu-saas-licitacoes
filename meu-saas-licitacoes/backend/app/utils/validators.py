import re
from typing import Tuple


def validar_cpf(cpf: str) -> bool:
    """
    Valida CPF brasileiro (11 dígitos, com dígitos verificadores).
    """
    cpf = re.sub(r"\D", "", cpf)
    if len(cpf) != 11 or cpf == cpf[0] * 11:
        return False

    # Validação do 1º dígito verificador
    soma = sum(int(cpf[i]) * (10 - i) for i in range(9))
    resto = (soma * 10) % 11
    if resto == 10:
        resto = 0
    if resto != int(cpf[9]):
        return False

    # Validação do 2º dígito verificador
    soma = sum(int(cpf[i]) * (11 - i) for i in range(10))
    resto = (soma * 10) % 11
    if resto == 10:
        resto = 0
    if resto != int(cpf[10]):
        return False

    return True


def validar_cnpj(cnpj: str) -> bool:
    """
    Valida CNPJ brasileiro (14 dígitos, com dígitos verificadores).
    """
    cnpj = re.sub(r"\D", "", cnpj)
    if len(cnpj) != 14 or cnpj == cnpj[0] * 14:
        return False

    # Pesos para validação
    def _calcular_digito(cnpj_parcial: str, pesos: list[int]) -> int:
        soma = sum(int(cnpj_parcial[i]) * pesos[i] for i in range(len(pesos)))
        resto = soma % 11
        return 0 if resto < 2 else 11 - resto

    # 1º dígito
    pesos_1 = [5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2]
    digito_1 = _calcular_digito(cnpj[:12], pesos_1)
    if digito_1 != int(cnpj[12]):
        return False

    # 2º dígito
    pesos_2 = [6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2]
    digito_2 = _calcular_digito(cnpj[:13], pesos_2)
    if digito_2 != int(cnpj[13]):
        return False

    return True


def formatar_cpf_cnpj(valor: str) -> str:
    """
    Formata CPF (XXX.XXX.XXX-XX) ou CNPJ (XX.XXX.XXX/XXXX-XX).
    """
    cleaned = re.sub(r"\D", "", valor)
    if len(cleaned) == 11:
        return f"{cleaned[:3]}.{cleaned[3:6]}.{cleaned[6:9]}-{cleaned[9:]}"
    elif len(cleaned) == 14:
        return (
            f"{cleaned[:2]}.{cleaned[2:5]}.{cleaned[5:8]}/"
            f"{cleaned[8:12]}-{cleaned[12:]}"
        )
    return valor


def validar_telefone(telefone: str) -> Tuple[bool, str]:
    """
    Valida telefone brasileiro (DDD + número, 10 ou 11 dígitos).
    Retorna (válido, número limpo).
    """
    cleaned = re.sub(r"\D", "", telefone)
    if len(cleaned) in (10, 11):
        return True, cleaned
    return False, cleaned