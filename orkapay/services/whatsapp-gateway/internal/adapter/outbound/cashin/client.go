package cashin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type CreateInvoiceRequest struct {
	AmountCents int64  `json:"amount_cents"`
	DueDate     string `json:"due_date"`
	PayerName   string `json:"payer_name"`
	PayerTaxID  string `json:"payer_tax_id"`
	Provider    string `json:"provider"`
}

type InvoiceResponse struct {
	ID        string `json:"ID"`
	Status    string `json:"Status"`
	PixQRCode string `json:"PixQRCode"`
	PixKey    string `json:"PixKey"`
	InvoiceURL string `json:"InvoiceURL"`
}

type IssueResponse struct {
	ID        string `json:"ID"`
	Status    string `json:"Status"`
	PixQRCode string `json:"PixQRCode"`
	PixKey    string `json:"PixKey"`
	InvoiceURL string `json:"InvoiceURL"`
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) CreateAndIssueInvoice(ctx context.Context, amountCents int64, payerName string) (*IssueResponse, error) {
	dueDate := time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")

	reqBody := CreateInvoiceRequest{
		AmountCents: amountCents,
		DueDate:     dueDate,
		PayerName:   payerName,
		PayerTaxID:  "00000000000",
		Provider:    "ASAAS",
	}

	body, _ := json.Marshal(reqBody)

	// 1. Criar invoice
	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/invoices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("cashin returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var invoice InvoiceResponse
	if err := json.Unmarshal(respBody, &invoice); err != nil {
		return nil, fmt.Errorf("failed to decode invoice response: %w", err)
	}

	// 2. Issue invoice (chama Asaas para gerar PIX/boleto)
	issueReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/invoices/"+invoice.ID+"/issue", nil)
	issueReq.Header.Set("X-API-Key", c.apiKey)

	issueResp, err := c.httpClient.Do(issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to issue invoice: %w", err)
	}
	defer issueResp.Body.Close()

	issueBody, _ := io.ReadAll(issueResp.Body)

	if issueResp.StatusCode >= 400 {
		return nil, fmt.Errorf("cashin issue returned status %d: %s", issueResp.StatusCode, string(issueBody))
	}

	var issued IssueResponse
	if err := json.Unmarshal(issueBody, &issued); err != nil {
		return nil, fmt.Errorf("failed to decode issue response: %w", err)
	}

	return &issued, nil
}