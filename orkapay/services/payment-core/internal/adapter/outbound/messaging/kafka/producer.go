package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/jp255ft-debug/payment-core/internal/core/domain/event"
	"github.com/jp255ft-debug/payment-core/internal/core/port"
)

// Config contem as configuracoes do producer Kafka.
type Config struct {
	Brokers      []string
	ClientID     string
	BatchSize    int
	BatchBytes   int
	BatchTimeout time.Duration
	Compression  string
	RequiredAcks string
	RetryMax     int
	RetryBackoff time.Duration
}

// DefaultConfig retorna configuracao padrao otimizada para sistemas financeiros.
func DefaultConfig() Config {
	return Config{
		Brokers:      []string{"localhost:9092"},
		ClientID:     "orkapay-payment-core",
		BatchSize:    100,
		BatchBytes:   1048576,
		BatchTimeout: 10 * time.Millisecond,
		Compression:  "snappy",
		RequiredAcks: "all",
		RetryMax:     3,
		RetryBackoff: 500 * time.Millisecond,
	}
}

// Topics define os topicos de mensageria do sistema.
type Topics struct {
	PaymentCreated   string
	PaymentApproved  string
	PaymentSettled   string
	PaymentFailed    string
	PaymentCancelled string
	PaymentRefunded  string
}

// DefaultTopics retorna os topicos padrao.
func DefaultTopics() Topics {
	return Topics{
		PaymentCreated:   "payment.created",
		PaymentApproved:  "payment.approved",
		PaymentSettled:   "payment.settled",
		PaymentFailed:    "payment.failed",
		PaymentCancelled: "payment.cancelled",
		PaymentRefunded:  "payment.refunded",
	}
}

// Producer implementa port.EventPublisher usando Sarama (Kafka).
type Producer struct {
	producer sarama.AsyncProducer
	topics   Topics
	cfg      Config
}

// NewProducer cria um novo producer Kafka.
func NewProducer(cfg Config, topics Topics) (*Producer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.ClientID = cfg.ClientID
	saramaCfg.Producer.RequiredAcks = parseAcks(cfg.RequiredAcks)
	saramaCfg.Producer.Compression = parseCompression(cfg.Compression)
	saramaCfg.Producer.Flush.MaxMessages = cfg.BatchSize
	saramaCfg.Producer.Flush.Bytes = cfg.BatchBytes
	saramaCfg.Producer.Flush.Frequency = cfg.BatchTimeout
	saramaCfg.Producer.Retry.Max = cfg.RetryMax
	saramaCfg.Producer.Retry.Backoff = cfg.RetryBackoff
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true
	saramaCfg.Net.MaxOpenRequests = 5

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar producer Kafka: %w", err)
	}

	p := &Producer{producer: producer, topics: topics}

	go func() {
		for e := range producer.Errors() {
			fmt.Printf("[KAFKA] Erro: %v\n", e)
		}
	}()

	return p, nil
}

func parseAcks(acks string) sarama.RequiredAcks {
	switch acks {
	case "none":
		return sarama.NoResponse
	case "one":
		return sarama.WaitForLocal
	default:
		return sarama.WaitForAll
	}
}

func parseCompression(comp string) sarama.CompressionCodec {
	switch comp {
	case "gzip":
		return sarama.CompressionGZIP
	default:
		return sarama.CompressionSnappy
	}
}

// PublishEvents implementa port.EventPublisher.
func (p *Producer) PublishEvents(ctx context.Context, events []event.DomainEvent) error {
	for _, e := range events {
		msg, err := p.buildMessage(e)
		if err != nil {
			return fmt.Errorf("erro ao construir mensagem: %w", err)
		}
		select {
		case p.producer.Input() <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (p *Producer) buildMessage(e event.DomainEvent) (*sarama.ProducerMessage, error) {
	topic := p.resolveTopic(e.EventName())
	if topic == "" {
		return nil, fmt.Errorf("topico nao encontrado: %s", e.EventName())
	}
	payload, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar: %w", err)
	}
	return &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(e.EventName()),
		Value: sarama.ByteEncoder(payload),
		Timestamp: time.Now().UTC(),
	}, nil
}

func (p *Producer) resolveTopic(eventName string) string {
	switch eventName {
	case "payment.created":
		return p.topics.PaymentCreated
	case "payment.approved":
		return p.topics.PaymentApproved
	case "payment.settled":
		return p.topics.PaymentSettled
	case "payment.failed":
		return p.topics.PaymentFailed
	case "payment.cancelled":
		return p.topics.PaymentCancelled
	case "payment.refunded":
		return p.topics.PaymentRefunded
	default:
		return ""
	}
}

// Close fecha o producer.
func (p *Producer) Close() error {
	return p.producer.Close()
}