package provider

import "time"

// CreateCustomerRequest representa os dados para criar um cliente no Asaas.
type CreateCustomerRequest struct {
	Name        string `json:"name"`
	CpfCnpj     string `json:"cpfCnpj"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	MobilePhone string `json:"mobilePhone,omitempty"`
	Address     string `json:"address,omitempty"`
	AddressNumber string `json:"addressNumber,omitempty"`
	Province    string `json:"province,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
}

// CustomerResponse representa a resposta da API de customer do Asaas.
type CustomerResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CpfCnpj     string `json:"cpfCnpj"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	MobilePhone string `json:"mobilePhone"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// CustomerListResponse representa a resposta de listagem de customers.
type CustomerListResponse struct {
	Data       []CustomerResponse `json:"data"`
	TotalCount int                `json:"totalCount"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

type CreateInvoiceRequest struct {
	Customer      string  `json:"customer"`
	BillingType   string  `json:"billingType"`
	DueDate       string  `json:"dueDate"`
	Value         float64 `json:"value"`
	Description   string  `json:"description,omitempty"`
	ExternalID    string  `json:"externalReference,omitempty"`
	Split         []Split `json:"split,omitempty"`
}

type Split struct {
	WalletId        string  `json:"walletId"`
	FixedValue      float64 `json:"fixedValue,omitempty"`
	PercentualValue float64 `json:"percentualValue,omitempty"`
}

type CreateInvoiceResponse struct {
	ID          string    `json:"id"`
	Customer    string    `json:"customer"`
	BillingType string    `json:"billingType"`
	DueDate     string    `json:"dueDate"`
	Value       float64   `json:"value"`
	Status      string    `json:"status"`
	PixQRCode   string    `json:"pixQrCode"`
	PixKey      string    `json:"pixKey"`
	Barcode     string    `json:"barcode"`
	InvoiceURL  string    `json:"invoiceUrl"`
	ExternalID  string    `json:"externalReference"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GetInvoiceResponse struct {
	ID          string  `json:"id"`
	Customer    string  `json:"customer"`
	BillingType string  `json:"billingType"`
	DueDate     string  `json:"dueDate"`
	Value       float64 `json:"value"`
	Status      string  `json:"status"`
	PixQRCode   string  `json:"pixQrCode"`
	PixKey      string  `json:"pixKey"`
	Barcode     string  `json:"barcode"`
	InvoiceURL  string  `json:"invoiceUrl"`
	ExternalID  string  `json:"externalReference"`
}

type CancelInvoiceResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ErrorResponse struct {
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}