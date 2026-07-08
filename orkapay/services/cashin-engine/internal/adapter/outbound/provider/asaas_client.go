package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AsaasClient struct {
	baseURL    string
	apiKey     string
	walletID   string
	httpClient *http.Client
}

func NewAsaasClient(baseURL, apiKey, walletID string) *AsaasClient {
	return &AsaasClient{
		baseURL:  baseURL,
		apiKey:   apiKey,
		walletID: walletID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *AsaasClient) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := c.baseURL + path
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && len(errResp.Errors) > 0 {
			return nil, fmt.Errorf("asaas error [%d]: %s", resp.StatusCode, errResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("asaas error [%d]: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// CreateCustomer cria um novo cliente no Asaas.
func (c *AsaasClient) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CustomerResponse, error) {
	respBody, err := c.doRequest(ctx, http.MethodPost, "/customers", req)
	if err != nil {
		return nil, err
	}
	var result CustomerResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse customer response: %w", err)
	}
	return &result, nil
}

// FindCustomerByTaxID busca um cliente pelo CPF/CNPJ no Asaas.
func (c *AsaasClient) FindCustomerByTaxID(ctx context.Context, taxID string) (*CustomerResponse, error) {
	path := "/customers?cpfCnpj=" + taxID
	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var list CustomerListResponse
	if err := json.Unmarshal(respBody, &list); err != nil {
		return nil, fmt.Errorf("failed to parse customer list: %w", err)
	}
	if len(list.Data) == 0 {
		return nil, nil
	}
	return &list.Data[0], nil
}

func (c *AsaasClient) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	// Split desabilitado temporariamente: Asaas rejeita split para a própria carteira.
	// TODO: reativar com subcontas (wallets diferentes) quando implementado.
	// if len(req.Split) == 0 && c.walletID != "" {
	// 	req.Split = []Split{
	// 		{
	// 			WalletId:   c.walletID,
	// 			FixedValue: 0.50,
	// 		},
	// 	}
	// }

	respBody, err := c.doRequest(ctx, http.MethodPost, "/payments", req)
	if err != nil {
		return nil, err
	}

	var result CreateInvoiceResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func (c *AsaasClient) GetInvoice(ctx context.Context, asaasID string) (*GetInvoiceResponse, error) {
	path := "/payments/" + asaasID
	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var result GetInvoiceResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

func (c *AsaasClient) CancelInvoice(ctx context.Context, asaasID string) error {
	path := "/payments/" + asaasID
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}