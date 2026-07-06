package outbound

import "context"

type BSPSender interface {
	SendTextMessage(ctx context.Context, toNumber, message string) error
	SendPaymentLink(ctx context.Context, toNumber, paymentLink string, amount float64) error
}