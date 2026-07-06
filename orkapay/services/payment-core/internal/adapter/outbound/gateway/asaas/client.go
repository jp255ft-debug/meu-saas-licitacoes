package asaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
)

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	baseRetryDelay = 1 * time.Second
	maxRetryDelay  = 4 * time.Second
)

// Client e o cliente HTTP para a API do Asaas.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClient cria uma nova instancia do cliente Asaas.
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// doRequest executa uma requisicao HTTP com retry e backoff exponencial.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * baseRetryDelay
			if delay > maxRetryDelay {
				delay = maxRetryDelay
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		lastErr = c.doRequestOnce(ctx, method, path, body, result)
		if lastErr == nil {
			return nil
		}

		if isNonRetryable(lastErr) {
			return lastErr
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return lastErr
}

func (c *Client) doRequestOnce(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("erro ao serializar request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("erro ao criar requisicao: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", c.apiKey)
	req.Header.Set("User-Agent", "OrkaPay-PaymentCore/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisicao: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler resposta: %w", err)
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("erro ao desserializar resposta: %w", err)
			}
		}
		return nil

	case resp.StatusCode == http.StatusUnauthorized:
		return domain.ErrMissingGatewayConfig

	case resp.StatusCode == http.StatusNotFound:
		return domain.ErrPaymentNotFound

	case resp.StatusCode == http.StatusUnprocessableEntity:
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && len(errResp.Errors) > 0 {
			return fmt.Errorf("%w: %s", domain.ErrGatewayRejection, errResp.Errors[0].Description)
		}
		return domain.ErrGatewayRejection

	default:
		if resp.StatusCode >= 500 {
			return fmt.Errorf("erro interno Asaas (HTTP %d)", resp.StatusCode)
		}
		return fmt.Errorf("erro inesperado (HTTP %d)", resp.StatusCode)
	}
}

func isNonRetryable(err error) bool {
	return err == domain.ErrMissingGatewayConfig ||
		err == domain.ErrPaymentNotFound ||
		err == domain.ErrGatewayRejection
}