package inbound

import "github.com/jp255ft-debug/whatsapp-gateway/internal/core/domain"

type CommandParser interface {
	Parse(rawMessage string) (*domain.Command, error)
}