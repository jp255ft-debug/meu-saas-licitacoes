package kafka

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// MockEventPublisher Ã© uma implementaÃ§Ã£o em memÃ³ria de EventPublisher.
// Usado em testes e no MVP antes de conectar ao Kafka real.
type MockEventPublisher struct {
	events []domain.DomainEvent
}

// NewMockEventPublisher cria um novo mock.
func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		events: make([]domain.DomainEvent, 0),
	}
}

// Publish armazena o evento em memÃ³ria (simula publicaÃ§Ã£o).
func (m *MockEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	m.events = append(m.events, event)
	fmt.Printf("[MOCK-KAFKA] Evento publicado: %s\n", event.EventName())
	return nil
}

// Close nÃ£o faz nada no mock.
func (m *MockEventPublisher) Close() error {
	return nil
}

// Events retorna todos os eventos publicados (para verificaÃ§Ã£o em testes).
func (m *MockEventPublisher) Events() []domain.DomainEvent {
	return m.events
}

// Reset limpa os eventos armazenados.
func (m *MockEventPublisher) Reset() {
	m.events = m.events[:0]
}

// Compile-time check: MockEventPublisher implementa port.EventPublisher
var _ port.EventPublisher = (*MockEventPublisher)(nil)
