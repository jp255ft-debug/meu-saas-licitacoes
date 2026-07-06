package events

import "time"

type PaymentConfirmedEvent struct {
	PaymentID   string    `json:"payment_id"`
	ExternalID  string    `json:"external_id"`
	Amount      int64     `json:"amount"`
	Status      string    `json:"status"`
	PixQRCode   string    `json:"pix_qr_code,omitempty"`
	PixKey      string    `json:"pix_key,omitempty"`
	Barcode     string    `json:"barcode,omitempty"`
	InvoiceURL  string    `json:"invoice_url,omitempty"`
	OccurredAt  time.Time `json:"occurred_at"`
}