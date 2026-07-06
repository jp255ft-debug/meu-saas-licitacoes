package asaas

// CreatePaymentRequest representa a requisicao para criar um pagamento no Asaas.
type CreatePaymentRequest struct {
	Customer          string `json:"customer"`
	BillingType       string `json:"billingType"`
	Value             int64  `json:"value"`
	DueDate           string `json:"dueDate"`
	Description       string `json:"description,omitempty"`
	ExternalReference string `json:"externalReference,omitempty"`
}

// CreatePaymentResponse representa a resposta da criacao de pagamento.
type CreatePaymentResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Status    string `json:"status"`
	DueDate   string `json:"dueDate"`
	Value     int64  `json:"value"`
	PixQRCode string `json:"pixQRCode,omitempty"`
}

// GetPaymentResponse representa a resposta de consulta de pagamento.
type GetPaymentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Value  int64  `json:"value"`
}

// WebhookEvent representa um evento recebido do Asaas.
type WebhookEvent struct {
	Event   string         `json:"event"`
	Payment WebhookPayment `json:"payment"`
}

// WebhookPayment representa os dados de pagamento no webhook.
type WebhookPayment struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Value  int64  `json:"value"`
}

// ErrorResponse representa uma resposta de erro da API Asaas.
type ErrorResponse struct {
	Errors []ErrorDetail `json:"errors"`
}

// ErrorDetail representa um detalhe de erro da API Asaas.
type ErrorDetail struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}