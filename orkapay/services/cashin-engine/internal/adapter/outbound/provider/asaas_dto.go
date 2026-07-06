package provider

import "time"

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