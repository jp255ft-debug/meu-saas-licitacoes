package domain

import (
	"testing"
)

func TestNewPixKey_EVP(t *testing.T) {
	pk, err := NewPixKey("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	if err != nil {
		t.Fatalf("esperado sem erro para EVP válida, got: %v", err)
	}
	if pk.Type() != PixKeyEVP {
		t.Errorf("esperado tipo EVP, got: %s", pk.Type())
	}
	if pk.Value() != "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" {
		t.Errorf("valor não corresponde")
	}
}

func TestNewPixKey_CPF_Valid(t *testing.T) {
	// CPF válido: 529.982.247-25
	pk, err := NewPixKey("52998224725")
	if err != nil {
		t.Fatalf("esperado sem erro para CPF válido, got: %v", err)
	}
	if pk.Type() != PixKeyCPF {
		t.Errorf("esperado tipo CPF, got: %s", pk.Type())
	}
}

func TestNewPixKey_CPF_Invalid(t *testing.T) {
	// CPF inválido
	_, err := NewPixKey("11111111111")
	if err != ErrInvalidPixKey {
		t.Fatalf("esperado ErrInvalidPixKey para CPF inválido, got: %v", err)
	}
}

func TestNewPixKey_CNPJ_Valid(t *testing.T) {
	// CNPJ válido: 11.222.333/0001-81
	pk, err := NewPixKey("11222333000181")
	if err != nil {
		t.Fatalf("esperado sem erro para CNPJ válido, got: %v", err)
	}
	if pk.Type() != PixKeyCNPJ {
		t.Errorf("esperado tipo CNPJ, got: %s", pk.Type())
	}
}

func TestNewPixKey_CNPJ_Invalid(t *testing.T) {
	_, err := NewPixKey("00000000000000")
	if err != ErrInvalidPixKey {
		t.Fatalf("esperado ErrInvalidPixKey para CNPJ inválido, got: %v", err)
	}
}

func TestNewPixKey_Email(t *testing.T) {
	pk, err := NewPixKey("teste@orkapay.com.br")
	if err != nil {
		t.Fatalf("esperado sem erro para email válido, got: %v", err)
	}
	if pk.Type() != PixKeyEmail {
		t.Errorf("esperado tipo EMAIL, got: %s", pk.Type())
	}
}

func TestNewPixKey_Phone(t *testing.T) {
	pk, err := NewPixKey("+5511999999999")
	if err != nil {
		t.Fatalf("esperado sem erro para telefone válido, got: %v", err)
	}
	if pk.Type() != PixKeyPhone {
		t.Errorf("esperado tipo PHONE, got: %s", pk.Type())
	}
}

func TestNewPixKey_Empty(t *testing.T) {
	_, err := NewPixKey("")
	if err != ErrEmptyPixKey {
		t.Fatalf("esperado ErrEmptyPixKey, got: %v", err)
	}
}

func TestNewPixKey_Unsupported(t *testing.T) {
	_, err := NewPixKey("chave_qualquer_123")
	if err != ErrUnsupportedPixKey {
		t.Fatalf("esperado ErrUnsupportedPixKey, got: %v", err)
	}
}

func TestPixKey_Equals(t *testing.T) {
	a := MustNewPixKey("52998224725")
	b := MustNewPixKey("52998224725")
	c := MustNewPixKey("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

	if !a.Equals(b) {
		t.Error("esperado a.Equals(b) = true")
	}
	if a.Equals(c) {
		t.Error("esperado a.Equals(c) = false")
	}
}

func TestPixKey_String(t *testing.T) {
	pk := MustNewPixKey("teste@orkapay.com")
	if pk.String() != "teste@orkapay.com" {
		t.Errorf("esperado 'teste@orkapay.com', got: %s", pk.String())
	}
}

func TestMustNewPixKey_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("esperado panic para chave inválida")
		}
	}()
	MustNewPixKey("")
}
