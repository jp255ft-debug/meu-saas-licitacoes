package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Money representa um valor monetário em centavos (inteiro).
// Usa int64 para evitar overflow e garantir precisão absoluta.
// Exemplo: R$ 1.234,56 é representado como 123456 (centavos).
type Money struct {
	cents int64
}

// NewMoney cria um Money a partir de centavos.
// Validação: não permite valores negativos.
func NewMoney(cents int64) (Money, error) {
	if cents < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{cents: cents}, nil
}

// MustNewMoney cria um Money a partir de centavos.
// Panica se o valor for negativo. Use apenas em inicializações
// onde o valor é garantidamente válido (ex: constantes).
func MustNewMoney(cents int64) Money {
	m, err := NewMoney(cents)
	if err != nil {
		panic(err)
	}
	return m
}

// Zero retorna Money com valor zero.
func Zero() Money {
	return Money{cents: 0}
}

// Cents retorna o valor em centavos.
func (m Money) Cents() int64 {
	return m.cents
}

// Add soma dois valores monetários.
func (m Money) Add(other Money) Money {
	return Money{cents: m.cents + other.cents}
}

// Sub subtrai dois valores monetários.
// Atenção: pode resultar em valor negativo. Valide antes se necessário.
func (m Money) Sub(other Money) Money {
	return Money{cents: m.cents - other.cents}
}

// Multiply multiplica o valor por um fator inteiro.
func (m Money) Multiply(factor int64) Money {
	return Money{cents: m.cents * factor}
}

// Divide divide o valor por um divisor inteiro.
// Atenção: perda de precisão por truncamento. Use com cuidado.
func (m Money) Divide(divisor int64) (Money, error) {
	if divisor == 0 {
		return Money{}, errors.New("divisão por zero")
	}
	return Money{cents: m.cents / divisor}, nil
}

// Percent calcula uma porcentagem do valor.
// Exemplo: m.Percent(10) = 10% de m
func (m Money) Percent(percent int64) Money {
	return Money{cents: (m.cents * percent) / 100}
}

// IsZero retorna true se o valor é zero.
func (m Money) IsZero() bool {
	return m.cents == 0
}

// IsPositive retorna true se o valor é maior que zero.
func (m Money) IsPositive() bool {
	return m.cents > 0
}

// Equals compara igualdade com outro Money.
func (m Money) Equals(other Money) bool {
	return m.cents == other.cents
}

// GreaterThan retorna true se m > other.
func (m Money) GreaterThan(other Money) bool {
	return m.cents > other.cents
}

// LessThan retorna true se m < other.
func (m Money) LessThan(other Money) bool {
	return m.cents < other.cents
}

// ToReais retorna o valor formatado em reais como string.
// Exemplo: 123456 -> "1234.56"
// Usa string para evitar perda de precisão do float64.
func (m Money) ToReais() string {
	reais := m.cents / 100
	centavos := m.cents % 100
	if centavos < 0 {
		centavos = -centavos
	}
	return fmt.Sprintf("%d.%02d", reais, centavos)
}

// String retorna o valor formatado para exibição: "R$ 1.234,56"
func (m Money) String() string {
	return formatBRL(m.cents)
}

// formatBRL formata centavos para o padrão brasileiro: "R$ 1.234,56"
func formatBRL(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}

	reais := cents / 100
	centavos := cents % 100

	// Formata os reais com separador de milhar
	reaisStr := strconv.FormatInt(reais, 10)
	var parts []string
	for i := len(reaisStr); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{reaisStr[start:i]}, parts...)
	}
	formattedReais := strings.Join(parts, ".")

	result := fmt.Sprintf("R$ %s,%02d", formattedReais, centavos)
	if negative {
		result = "-" + result
	}
	return result
}
