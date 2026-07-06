package domain

import "time"

// Webhook represents a webhook configuration for client notification.
type Webhook struct {
	ID        string    `json:"id"`
	ClientID  string    `json:"client_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Active    bool      `json:"active"`
	Secret    string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WebhookEvent represents a notification event sent to clients.
type WebhookEvent struct {
	Event     string      `json:"event"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ShouldDeliver checks if this webhook should be triggered for an event type.
func (w *Webhook) ShouldDeliver(eventType string) bool {
	if !w.Active {
		return false
	}
	for _, evt := range w.Events {
		if evt == eventType {
			return true
		}
	}
	return false
}

// Event types constants.
const (
	EventInvoicePaid     = "invoice.paid"
	EventInvoiceExpired  = "invoice.expired"
	EventInvoiceCanceled = "invoice.canceled"
	EventTransferSettled = "transfer.settled"
)