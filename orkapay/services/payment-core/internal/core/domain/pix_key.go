package domain

import (
	"regexp"
	"strings"
)

// PixKeyType define os tipos válidos de chave Pix.
type PixKeyType string

const (
	PixKeyCPF   PixKeyType = "CPF"
	PixKeyCNPJ  PixKeyType = "CNPJ"
	PixKeyEmail PixKeyType = "EMAIL"
	PixKeyPhone PixKeyType = "PHONE"
	PixKeyEVP   PixKeyType = "EVP" // Chave aleatória (EvP)
)

// PixKey representa uma chave Pix validada.
// Value Object imutável.
type PixKey struct {
	value   string
	keyType PixKeyType
}

// NewPixKey cria uma PixKey a partir de uma string.
// Detecta automaticamente o tipo e valida o formato.
func NewPixKey(value string) (PixKey, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return PixKey{}, ErrEmptyPixKey
	}

	keyType := detectPixKeyType(value)
	if keyType == "" {
		return PixKey{}, ErrUnsupportedPixKey
	}

	if !validatePixKeyFormat(value, keyType) {
		return PixKey{}, ErrInvalidPixKey
	}

	return PixKey{
		value:   value,
		keyType: keyType,
	}, nil
}

// MustNewPixKey cria uma PixKey ou panic se inválida.
func MustNewPixKey(value string) PixKey {
	pk, err := NewPixKey(value)
	if err != nil {
		panic(err)
	}
	return pk
}

// Value retorna o valor da chave Pix.
func (p PixKey) Value() string {
	return p.value
}

// Type retorna o tipo da chave Pix.
func (p PixKey) Type() PixKeyType {
	return p.keyType
}

// Equals compara igualdade com outra PixKey.
func (p PixKey) Equals(other PixKey) bool {
	return p.value == other.value && p.keyType == other.keyType
}

// String retorna a representação da chave Pix.
func (p PixKey) String() string {
	return p.value
}

// detectPixKeyType detecta o tipo de chave Pix pelo formato.
func detectPixKeyType(value string) PixKeyType {
	// EVP (UUID v4)
	if matched, _ := regexp.MatchString(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`, value); matched {
		return PixKeyEVP
	}

	// CPF: 11 dígitos
	if matched, _ := regexp.MatchString(`^\d{11}$`, value); matched {
		return PixKeyCPF
	}

	// CNPJ: 14 dígitos
	if matched, _ := regexp.MatchString(`^\d{14}$`, value); matched {
		return PixKeyCNPJ
	}

	// Email
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, value); matched {
		return PixKeyEmail
	}

	// Phone: +55 ou 55 seguido de DDD + 9 dígitos
	if matched, _ := regexp.MatchString(`^(\+55|55)\d{11}$`, value); matched {
		return PixKeyPhone
	}

	return ""
}

// validatePixKeyFormat valida o formato específico de cada tipo.
func validatePixKeyFormat(value string, keyType PixKeyType) bool {
	switch keyType {
	case PixKeyCPF:
		return validateCPF(value)
	case PixKeyCNPJ:
		return validateCNPJ(value)
	case PixKeyEmail:
		return len(value) <= 77 // Limite do BACEN
	case PixKeyPhone:
		return len(value) >= 13 && len(value) <= 14
	case PixKeyEVP:
		return true // Já validado pelo regex
	default:
		return false
	}
}

// validateCPF valida dígitos verificadores do CPF.
func validateCPF(cpf string) bool {
	if len(cpf) != 11 {
		return false
	}

	// CPFs com todos dígitos iguais são inválidos
	allSame := true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// Valida primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(cpf[i]-'0') * (10 - i)
	}
	digit1 := 11 - (sum % 11)
	if digit1 >= 10 {
		digit1 = 0
	}
	if int(cpf[9]-'0') != digit1 {
		return false
	}

	// Valida segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(cpf[i]-'0') * (11 - i)
	}
	digit2 := 11 - (sum % 11)
	if digit2 >= 10 {
		digit2 = 0
	}
	if int(cpf[10]-'0') != digit2 {
		return false
	}

	return true
}

// validateCNPJ valida dígitos verificadores do CNPJ.
func validateCNPJ(cnpj string) bool {
	if len(cnpj) != 14 {
		return false
	}

	// CNPJs com todos dígitos iguais são inválidos
	allSame := true
	for i := 1; i < len(cnpj); i++ {
		if cnpj[i] != cnpj[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(cnpj[i]-'0') * weights1[i]
	}
	digit1 := 11 - (sum % 11)
	if digit1 >= 10 {
		digit1 = 0
	}
	if int(cnpj[12]-'0') != digit1 {
		return false
	}

	sum = 0
	for i := 0; i < 13; i++ {
		sum += int(cnpj[i]-'0') * weights2[i]
	}
	digit2 := 11 - (sum % 11)
	if digit2 >= 10 {
		digit2 = 0
	}
	if int(cnpj[13]-'0') != digit2 {
		return false
	}

	return true
}
