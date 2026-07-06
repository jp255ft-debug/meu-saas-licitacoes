package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OutboxEvent representa um evento a ser publicado no barramento.
type OutboxEvent struct {
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       interface{}
}

// OutboxRepository gerencia a tabela outbox_events.
type OutboxRepository struct {
	pool *pgxpool.Pool
}

// NewOutboxRepository cria uma nova instância do repositório.
func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

// SaveEvent salva um evento na tabela outbox dentro de uma transação existente.
func (r *OutboxRepository) SaveEvent(ctx context.Context, tx pgx.Tx, event OutboxEvent) error {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("falha ao serializar payload do outbox: %w", err)
	}

	query := `
		INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload, status)
		VALUES ($1, $2, $3, $4, 'PENDING')
	`

	_, err = tx.Exec(ctx, query,
		event.AggregateType,
		event.AggregateID,
		event.EventType,
		payloadJSON,
	)
	if err != nil {
		return fmt.Errorf("falha ao inserir evento outbox: %w", err)
	}

	return nil
}

// FetchPending busca eventos PENDING para processamento, limitando por lote.
func (r *OutboxRepository) FetchPending(ctx context.Context, limit int) ([]OutboxEventRow, error) {
	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar eventos outbox pendentes: %w", err)
	}
	defer rows.Close()

	var events []OutboxEventRow
	for rows.Next() {
		var e OutboxEventRow
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("falha ao escanear evento outbox: %w", err)
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// MarkAsProcessed marca um evento como PROCESSED.
func (r *OutboxRepository) MarkAsProcessed(ctx context.Context, id string) error {
	query := `
		UPDATE outbox_events
		SET status = 'PROCESSED', processed_at = $1
		WHERE id = $2 AND status = 'PENDING'
	`

	_, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("falha ao marcar outbox como PROCESSED: %w", err)
	}

	return nil
}

// MarkAsFailed marca um evento como FAILED após tentativas esgotadas.
func (r *OutboxRepository) MarkAsFailed(ctx context.Context, id string) error {
	query := `
		UPDATE outbox_events
		SET status = 'FAILED', processed_at = $1
		WHERE id = $2 AND status = 'PENDING'
	`

	_, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("falha ao marcar outbox como FAILED: %w", err)
	}

	return nil
}

// OutboxEventRow representa uma linha da tabela outbox_events.
type OutboxEventRow struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	CreatedAt     time.Time
}