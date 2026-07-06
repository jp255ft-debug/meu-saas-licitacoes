package invoice

import (
	"context"
	"time"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
	"github.com/jp255ft-debug/cashin-engine/internal/service/webhook"
)

// ServiceImpl implementa inbound.InvoiceService orquestrando os use cases.
type ServiceImpl struct {
	createUseCase  *CreateInvoiceUseCase
	issueUseCase   *IssueInvoiceUseCase
	payUseCase     *PayInvoiceUseCase
	cancelUseCase  *CancelInvoiceUseCase
	expireUseCase  *ExpireInvoiceUseCase
	getByIDUseCase *GetInvoiceByIDUseCase
	listByStatusUC *ListInvoicesByStatusUseCase
	webhookSvc     *webhook.WebhookService
}

// NewService cria uma nova instância do serviço de fatura.
func NewService(repo outbound.InvoiceRepository, provider outbound.CashInProviderGateway, webhookSvc *webhook.WebhookService) *ServiceImpl {
	return &ServiceImpl{
		createUseCase:  NewCreateInvoiceUseCase(repo, provider),
		issueUseCase:   NewIssueInvoiceUseCase(repo, provider),
		payUseCase:     NewPayInvoiceUseCase(repo),
		cancelUseCase:  NewCancelInvoiceUseCase(repo, provider),
		expireUseCase:  NewExpireInvoiceUseCase(repo),
		getByIDUseCase: NewGetInvoiceByIDUseCase(repo),
		listByStatusUC: NewListInvoicesByStatusUseCase(repo),
		webhookSvc:     webhookSvc,
	}
}

// Create cria uma nova fatura.
func (s *ServiceImpl) Create(ctx context.Context, input inbound.CreateInvoiceInput) (*inbound.InvoiceOutput, error) {
	return s.createUseCase.Execute(ctx, input)
}

// Issue emite uma fatura pendente.
func (s *ServiceImpl) Issue(ctx context.Context, id string) (*inbound.InvoiceOutput, error) {
	return s.issueUseCase.Execute(ctx, id)
}

// Pay registra o pagamento e dispara webhook.
func (s *ServiceImpl) Pay(ctx context.Context, id string, paidAmountCents int64, paymentMethod string) (*inbound.InvoiceOutput, error) {
	result, err := s.payUseCase.Execute(ctx, id, paidAmountCents, paymentMethod)
	if err == nil && s.webhookSvc != nil {
		s.webhookSvc.Dispatch(ctx, "", domain.EventInvoicePaid, map[string]interface{}{
			"invoice_id":   id,
			"amount_cents": paidAmountCents,
			"paid_at":      time.Now().UTC().Format(time.RFC3339),
		})
	}
	return result, err
}

// Cancel cancela uma fatura e dispara webhook.
func (s *ServiceImpl) Cancel(ctx context.Context, id, reason string) (*inbound.InvoiceOutput, error) {
	result, err := s.cancelUseCase.Execute(ctx, id, reason)
	if err == nil && s.webhookSvc != nil {
		s.webhookSvc.Dispatch(ctx, "", domain.EventInvoiceCanceled, map[string]interface{}{
			"invoice_id":  id,
			"reason":      reason,
			"canceled_at": time.Now().UTC().Format(time.RFC3339),
		})
	}
	return result, err
}

// Expire marca uma fatura como expirada e dispara webhook.
func (s *ServiceImpl) Expire(ctx context.Context, id string) (*inbound.InvoiceOutput, error) {
	result, err := s.expireUseCase.Execute(ctx, id)
	if err == nil && s.webhookSvc != nil {
		s.webhookSvc.Dispatch(ctx, "", domain.EventInvoiceExpired, map[string]interface{}{
			"invoice_id": id,
			"expired_at": time.Now().UTC().Format(time.RFC3339),
		})
	}
	return result, err
}

// GetByID busca uma fatura pelo ID.
func (s *ServiceImpl) GetByID(ctx context.Context, id string) (*entity.Invoice, error) {
	return s.getByIDUseCase.Execute(ctx, id)
}

// ListByStatus lista faturas por status.
func (s *ServiceImpl) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entity.Invoice, error) {
	return s.listByStatusUC.Execute(ctx, status, limit, offset)
}

// Compile-time check
var _ inbound.InvoiceService = (*ServiceImpl)(nil)