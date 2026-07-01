package domain

import (
	"testing"
)

func TestNewMoney_Valid(t *testing.T) {
	m, err := NewMoney(123456)
	if err != nil {
		t.Fatalf("esperado sem erro, got: %v", err)
	}
	if m.Cents() != 123456 {
		t.Errorf("esperado 123456 centavos, got: %d", m.Cents())
	}
}

func TestNewMoney_Negative(t *testing.T) {
	_, err := NewMoney(-100)
	if err != ErrNegativeAmount {
		t.Fatalf("esperado ErrNegativeAmount, got: %v", err)
	}
}

func TestMustNewMoney(t *testing.T) {
	m := MustNewMoney(500)
	if m.Cents() != 500 {
		t.Errorf("esperado 500 centavos, got: %d", m.Cents())
	}
}

func TestMustNewMoney_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("esperado panic para valor negativo")
		}
	}()
	MustNewMoney(-1)
}

func TestZero(t *testing.T) {
	m := Zero()
	if !m.IsZero() {
		t.Error("esperado IsZero() = true")
	}
	if m.IsPositive() {
		t.Error("esperado IsPositive() = false")
	}
}

func TestMoney_Add(t *testing.T) {
	a := MustNewMoney(100)
	b := MustNewMoney(200)
	result := a.Add(b)
	if result.Cents() != 300 {
		t.Errorf("esperado 300 centavos, got: %d", result.Cents())
	}
}

func TestMoney_Sub(t *testing.T) {
	a := MustNewMoney(500)
	b := MustNewMoney(200)
	result := a.Sub(b)
	if result.Cents() != 300 {
		t.Errorf("esperado 300 centavos, got: %d", result.Cents())
	}
}

func TestMoney_Subtract_NegativeResult(t *testing.T) {
	a := MustNewMoney(100)
	b := MustNewMoney(200)
	result := a.Sub(b)
	// Sub pode retornar negativo — é comportamento esperado
	if result.Cents() != -100 {
		t.Errorf("esperado -100 centavos, got: %d", result.Cents())
	}
}

func TestMoney_Multiply(t *testing.T) {
	m := MustNewMoney(150)
	result := m.Multiply(3)
	if result.Cents() != 450 {
		t.Errorf("esperado 450 centavos, got: %d", result.Cents())
	}
}

func TestMoney_Divide(t *testing.T) {
	m := MustNewMoney(1000)
	result, err := m.Divide(4)
	if err != nil {
		t.Fatalf("esperado sem erro, got: %v", err)
	}
	if result.Cents() != 250 {
		t.Errorf("esperado 250 centavos, got: %d", result.Cents())
	}
}

func TestMoney_Divide_ByZero(t *testing.T) {
	m := MustNewMoney(1000)
	_, err := m.Divide(0)
	if err == nil {
		t.Error("esperado erro de divisão por zero")
	}
}

func TestMoney_Percent(t *testing.T) {
	m := MustNewMoney(10000) // R$ 100,00
	result := m.Percent(10) // 10%
	if result.Cents() != 1000 {
		t.Errorf("esperado 1000 centavos (10%% de 10000), got: %d", result.Cents())
	}
}

func TestMoney_Equals(t *testing.T) {
	a := MustNewMoney(500)
	b := MustNewMoney(500)
	c := MustNewMoney(501)

	if !a.Equals(b) {
		t.Error("esperado a.Equals(b) = true")
	}
	if a.Equals(c) {
		t.Error("esperado a.Equals(c) = false")
	}
}

func TestMoney_GreaterThan(t *testing.T) {
	a := MustNewMoney(1000)
	b := MustNewMoney(500)
	if !a.GreaterThan(b) {
		t.Error("esperado a.GreaterThan(b) = true")
	}
	if b.GreaterThan(a) {
		t.Error("esperado b.GreaterThan(a) = false")
	}
}

func TestMoney_LessThan(t *testing.T) {
	a := MustNewMoney(100)
	b := MustNewMoney(200)
	if !a.LessThan(b) {
		t.Error("esperado a.LessThan(b) = true")
	}
	if b.LessThan(a) {
		t.Error("esperado b.LessThan(a) = false")
	}
}

func TestMoney_ToReais(t *testing.T) {
	tests := []struct {
		cents    int64
		expected string
	}{
		{0, "0.00"},
		{100, "1.00"},
		{123456, "1234.56"},
		{99, "0.99"},
		{100000, "1000.00"},
	}

	for _, tt := range tests {
		m := MustNewMoney(tt.cents)
		got := m.ToReais()
		if got != tt.expected {
			t.Errorf("ToReais(%d): esperado %s, got: %s", tt.cents, tt.expected, got)
		}
	}
}

func TestMoney_String(t *testing.T) {
	tests := []struct {
		cents    int64
		expected string
	}{
		{0, "R$ 0,00"},
		{100, "R$ 1,00"},
		{123456, "R$ 1.234,56"},
		{99, "R$ 0,99"},
		{100000000, "R$ 1.000.000,00"},
	}

	for _, tt := range tests {
		m := MustNewMoney(tt.cents)
		got := m.String()
		if got != tt.expected {
			t.Errorf("String(%d): esperado %s, got: %s", tt.cents, tt.expected, got)
		}
	}
}

func TestMoney_String_Negative(t *testing.T) {
	m := Money{cents: -100}
	got := m.String()
	expected := "-R$ 1,00"
	if got != expected {
		t.Errorf("String(-100): esperado %s, got: %s", expected, got)
	}
}
