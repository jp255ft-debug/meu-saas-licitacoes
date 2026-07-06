package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
)

// WebhookRepository implementa port outbound.WebhookRepository usando PostgreSQL.
type WebhookRepository struct {
	pool *pgxpool.Pool
}

// NewWebhookRepository cria uma nova instância do repositório.
func NewWebhookRepository(pool *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{pool: pool}
}

// Save insere um novo webhook.
func (r *WebhookRepository) Save(ctx context.Context, wh *domain.Webhook) error {
	query := `
		INSERT INTO webhooks (id, client_id, url, events, active, secret, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, query,
		wh.ID, wh.ClientID, wh.URL, wh.Events, wh.Active, wh.Secret, now, now,
	)
	if err != nil {
		return fmt.Errorf("falha ao inserir webhook: %w", err)
	}
	wh.CreatedAt = now
	wh.UpdatedAt = now
	return nil
}

// FindByClientID busca todos os webhooks de um cliente.
func (r *WebhookRepository) FindByClientID(ctx context.Context, clientID string) ([]*domain.Webhook, error) {
	query := `SELECT id, client_id, url, events, active, secret, created_at, updated_at FROM webhooks WHERE client_id = $1`
	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar webhooks: %w", err)
	}
	defer rows.Close()

	return scanWebhooks(rows)
}

// FindByID busca um webhook pelo ID.
func (r *WebhookRepository) FindByID(ctx context.Context, id string) (*domain.Webhook, error) {
	query := `SELECT id, client_id, url, events, active, secret, created_at, updated_at FROM webhooks WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	wh, err := scanWebhook(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("webhook não encontrado")
		}
		return nil, fmt.Errorf("falha ao buscar webhook: %w", err)
	}
	return wh, nil
}

// Update atualiza um webhook existente.
func (r *WebhookRepository) Update(ctx context.Context, wh *domain.Webhook) error {
	query := `
		UPDATE webhooks SET url = $1, events = $2, active = $3, updated_at = $4
		WHERE id = $5
	`
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, query, wh.URL, wh.Events, wh.Active, now, wh.ID)
	if err != nil {
		return fmt.Errorf("falha ao atualizar webhook: %w", err)
	}
	wh.UpdatedAt = now
	return nil
}

// Delete remove um webhook.
func (r *WebhookRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM webhooks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("falha ao deletar webhook: %w", err)
	}
	return nil
}

// scanWebhook lê uma linha do pgx.Row.
func scanWebhook(row pgx.Row) (*domain.Webhook, error) {
	var wh domain.Webhook
	err := row.Scan(&wh.ID, &wh.ClientID, &wh.URL, &wh.Events, &wh.Active, &wh.Secret, &wh.CreatedAt, &wh.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

// scanWebhooks converte pgx.Rows em slice de Webhook.
func scanWebhooks(rows pgx.Rows) ([]*domain.Webhook, error) {
	var webhooks []*domain.Webhook
	for rows.Next() {
		wh, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, wh)
	}
	return webhooks, rows.Err()
}