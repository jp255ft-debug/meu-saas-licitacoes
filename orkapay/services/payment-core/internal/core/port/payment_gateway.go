package port

import (
	"context"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
)

// PaymentGateway define a interface para comunicaÃ§Ã£o com BaaS providers.
// Interface de outbound port â€” implementada por adaptadores especÃ­ficos
// (Asaas, Stark Bank, Celcoin, Fitbank).
// Segue o Adapter Pattern para plugabilidade de gateways.
type PaymentGateway interface {
	// CreatePayment solicita a criaÃ§Ã£o de um pagamento no gateway.
	// Retorna o ID externo da transaÃ§Ã£o e o QR Code (se aplicÃ¡vel).
	CreatePayment(ctx context.Context, payment *domain.Payment) (*GatewayResponse, error)

	// GetPaymentStatus consulta o status atual de um pagamento no gateway.
	GetPaymentStatus(ctx context.Context, externalID string) (*GatewayStatus, error)

	// CancelPayment solicita o cancelamento de um pagamento pendente.
	CancelPayment(ctx context.Context, externalID string, reason string) error

	// RefundPayment solicita o estorno de um pagamento jÃ¡ liquidado.
	RefundPayment(ctx context.Context, externalID string, amount domain.Money, reason string) (*GatewayResponse, error)

	// HealthCheck verifica se o gateway estÃ¡ operacional.
	HealthCheck(ctx context.Context) error

	// Name retorna o identificador do gateway (ex: "asaas", "starkbank").
	Name() string
}

// GatewayResponse encapsula a resposta de uma operaÃ§Ã£o no gateway.
type GatewayResponse struct {
	ExternalID    string
	QRCode        string // Para Pix QR Code
	CopyPaste     string // Para Pix Copia e Cola
	BoletoURL     string // Para boleto
	BoletoPDF     string // URL do PDF do boleto
	DigitableLine string // Linha digitÃ¡vel do boleto
	ExpiresAt     *time.Time
}

// GatewayStatus representa o status de um pagamento no gateway externo.
type GatewayStatus struct {
	ExternalID    string
	Status        string               // Status raw do gateway
	MappedStatus  domain.PaymentStatus // Status mapeado para o domÃ­nio
	Amount        domain.Money
	PaidAt        *time.Time
	FailureReason string
}
