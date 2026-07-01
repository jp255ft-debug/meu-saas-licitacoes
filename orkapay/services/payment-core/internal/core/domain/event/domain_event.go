package event

import "time"

// DomainEvent é a interface base para todos os eventos de domínio.
// Event Sourcing: eventos são imutáveis e representam fatos ocorridos.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// EventName é um tipo base para nomes de eventos.
type EventName string

func (e EventName) String() string { return string(e) }
