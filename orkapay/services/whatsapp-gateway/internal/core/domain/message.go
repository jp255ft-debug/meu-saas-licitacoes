package domain

import "time"

type WhatsAppMessage struct {
	ID          string
	FromNumber  string
	ToNumber    string
	Body        string
	Timestamp   time.Time
	MessageType string
}