package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/port/outbound"
)

type Service struct {
	sender outbound.BSPSender
}

func NewService(sender outbound.BSPSender) *Service {
	return &Service{sender: sender}
}

func (s *Service) SendPaymentConfirmation(ctx context.Context, toNumber, invoiceID string, amount int64) error {
	msg := fmt.Sprintf(
		"✅ *Pagamento confirmado!*\n\n"+
			"📄 *Fatura:* %s\n"+
			"💰 *Valor:* R$ %.2f\n"+
			"📅 *Data:* %s\n\n"+
			"O pagamento foi recebido e creditado em sua conta.\n\n"+
			"Obrigado por usar o OrkaPay! 🚀",
		invoiceID,
		float64(amount)/100.0,
		time.Now().Format("02/01/2006 às 15:04"),
	)

	slog.InfoContext(ctx, "sending payment confirmation",
		"to", toNumber,
		"invoice", invoiceID,
		"amount", amount,
	)

	return s.sender.SendTextMessage(ctx, toNumber, msg)
}