package whatsapp

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/domain"
	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/port/outbound"
)

type Service struct {
	sender outbound.BSPSender
}

func NewService(sender outbound.BSPSender) *Service {
	return &Service{sender: sender}
}

func (s *Service) ProcessMessage(ctx context.Context, msg *domain.WhatsAppMessage) error {
	slog.InfoContext(ctx, "processing message", "from", msg.FromNumber, "body", msg.Body)

	cmd, err := s.parseCommand(msg.Body)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse command", "error", err)
		return s.sender.SendTextMessage(ctx, msg.FromNumber, s.helpMessage())
	}

	switch cmd.CommandType {
	case "CREATE_INVOICE":
		return s.handleCreateInvoice(ctx, msg.FromNumber, cmd)
	case "HELP":
		return s.sender.SendTextMessage(ctx, msg.FromNumber, s.helpMessage())
	default:
		return s.sender.SendTextMessage(ctx, msg.FromNumber, "Comando não reconhecido. Digite /ajuda para ver os comandos disponíveis.")
	}
}

func (s *Service) parseCommand(raw string) (*domain.Command, error) {
	raw = strings.TrimSpace(raw)
	cmd := &domain.Command{
		RawMessage: raw,
		ParsedAt:   time.Now(),
	}

	re := regexp.MustCompile(`^/cobrar\s+([0-9]+[.,]?[0-9]{0,2})\s+(.*)$`)
	matches := re.FindStringSubmatch(raw)
	if len(matches) == 3 {
		amountStr := strings.Replace(matches[1], ",", ".", 1)
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, fmt.Errorf("valor inválido: %w", err)
		}
		cmd.CommandType = "CREATE_INVOICE"
		cmd.Amount = amount
		cmd.PayerName = strings.TrimSpace(matches[2])
		return cmd, nil
	}

	if raw == "/ajuda" || raw == "/help" {
		cmd.CommandType = "HELP"
		return cmd, nil
	}

	return nil, fmt.Errorf("comando não reconhecido: %s", raw)
}

func (s *Service) handleCreateInvoice(ctx context.Context, fromNumber string, cmd *domain.Command) error {
	slog.InfoContext(ctx, "creating invoice", "amount", cmd.Amount, "payer", cmd.PayerName)

	paymentLink := "https://orkapay-cashin.onrender.com/pay/simulated-uuid"

	message := fmt.Sprintf(
		"✅ Cobrança gerada com sucesso!\n\n"+
			"💳 *Valor:* R$ %.2f\n"+
			"👤 *Pagador:* %s\n"+
			"🔗 *Link de pagamento:* %s\n\n"+
			"Envie este link para o pagador ou use o QR Code.",
		cmd.Amount, cmd.PayerName, paymentLink,
	)
	return s.sender.SendTextMessage(ctx, fromNumber, message)
}

func (s *Service) helpMessage() string {
	return "📋 *Comandos disponíveis:*\n\n" +
		"1️⃣ */cobrar <valor> <nome>*\n" +
		"   Ex: `/cobrar 150.00 João Silva`\n" +
		"   → Gera um link de pagamento Pix/Boleto.\n\n" +
		"2️⃣ */ajuda* ou */help*\n" +
		"   → Mostra esta mensagem.\n\n" +
		"🔧 *Em breve:* status de pagamentos, extratos e muito mais!"
}