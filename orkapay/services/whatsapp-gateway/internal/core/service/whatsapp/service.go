package whatsapp

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jp255ft-debug/whatsapp-gateway/internal/adapter/outbound/cashin"
	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/domain"
	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/port/outbound"
)

type Service struct {
	sender   outbound.BSPSender
	cashin   *cashin.Client
}

func NewService(sender outbound.BSPSender, cashinClient *cashin.Client) *Service {
	return &Service{sender: sender, cashin: cashinClient}
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
	slog.InfoContext(ctx, "creating invoice via cashin-engine",
		"amount", cmd.Amount,
		"payer", cmd.PayerName,
		"from", fromNumber,
	)

	amountCents := int64(cmd.Amount * 100)

	if s.cashin != nil {
		issued, err := s.cashin.CreateAndIssueInvoice(ctx, amountCents, cmd.PayerName)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create invoice via cashin", "error", err)
			message := fmt.Sprintf(
				"❌ *Erro ao gerar cobrança*\n\n"+
					"Não foi possível criar a cobrança no momento. Tente novamente mais tarde.\n\n"+
					"Detalhes: %v",
				err,
			)
			return s.sender.SendTextMessage(ctx, fromNumber, message)
		}

		message := fmt.Sprintf(
			"✅ *Cobrança gerada com sucesso!*\n\n"+
				"💳 *Valor:* R$ %.2f\n"+
				"👤 *Pagador:* %s\n"+
				"📄 *ID:* %s\n"+
				"📌 *Status:* %s\n",
			cmd.Amount, cmd.PayerName, issued.ID, issued.Status,
		)

		if issued.PixQRCode != "" {
			message += fmt.Sprintf("📱 *Pix Copia e Cola:*\n`%s`\n\n", issued.PixQRCode)
		}
		if issued.PixKey != "" {
			message += fmt.Sprintf("🔑 *Chave Pix:* %s\n", issued.PixKey)
		}
		if issued.InvoiceURL != "" {
			message += fmt.Sprintf("🔗 *Link:* %s\n", issued.InvoiceURL)
		}

		message += "\nPague utilizando o Pix ou Boleto. O pagamento será confirmado automaticamente."

		return s.sender.SendTextMessage(ctx, fromNumber, message)
	}

	// Fallback: cashin não configurado
	paymentLink := "https://orkapay-cashin.onrender.com/pay/simulated-uuid"
	message := fmt.Sprintf(
		"✅ Cobrança gerada com sucesso!\n\n"+
			"💳 *Valor:* R$ %.2f\n"+
			"👤 *Pagador:* %s\n"+
			"🔗 *Link de pagamento:* %s\n\n"+
			"⚠️ Modo simulado — CASHIN_BASE_URL não configurada.",
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