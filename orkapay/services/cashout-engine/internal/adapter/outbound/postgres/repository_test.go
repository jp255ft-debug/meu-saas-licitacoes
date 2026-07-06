//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// ---------------------------------------------------------------------------
// Setup / Teardown
// ---------------------------------------------------------------------------

// startPostgresContainer inicia um container PostgreSQL via Testcontainers.
func startPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("SKIP_INTEGRATION está definido — pulando teste de integração")
	}

	pool, cleanup, err := startTestContainer(t)
	if err != nil {
		t.Skipf("Docker não disponível ou Testcontainers falhou: %v — pulando teste de integração", err)
	}

	return pool, cleanup
}

// startTestContainer tenta iniciar o container PostgreSQL via Testcontainers.
func startTestContainer(t *testing.T) (*pgxpool.Pool, func(), error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "cashout_test",
		},
		Cmd: []string{"postgres", "-c", "fsync=off"},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao iniciar container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("falha ao obter host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("falha ao obter porta: %w", err)
	}

	connString := fmt.Sprintf("postgres://test:test@%s:%s/cashout_test?sslmode=disable", host, port.Port())

	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, connString)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				break
			}
			pool.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("falha ao conectar ao PostgreSQL: %w", err)
	}

	if err := runMigrations(ctx, pool); err != nil {
		pool.Close()
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("falha ao rodar migrations: %w", err)
	}

	cleanup := func() {
		pool.Close()
		container.Terminate(ctx)
	}

	return pool, cleanup, nil
}

// runMigrations executa o script SQL de criação da tabela.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migration := `
	CREATE TABLE IF NOT EXISTS transfers (
		id VARCHAR(64) PRIMARY KEY,
		external_id VARCHAR(128),
		status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
		amount_cents BIGINT NOT NULL,
		payee_name VARCHAR(255) NOT NULL,
		payee_tax_id VARCHAR(32) NOT NULL,
		payee_pix_key VARCHAR(128) NOT NULL,
		payee_pix_type VARCHAR(16),
		payee_bank_code VARCHAR(16),
		payee_branch VARCHAR(16),
		payee_account VARCHAR(32),
		payee_account_type VARCHAR(16),
		provider VARCHAR(32) NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		metadata JSONB DEFAULT '{}',
		processed_at TIMESTAMPTZ,
		settled_at TIMESTAMPTZ,
		failed_at TIMESTAMPTZ,
		cancelled_at TIMESTAMPTZ,
		failure_reason TEXT,
		version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
	CREATE INDEX IF NOT EXISTS idx_transfers_external_id ON transfers(external_id);
	CREATE INDEX IF NOT EXISTS idx_transfers_payee_tax_id ON transfers(payee_tax_id);
	`

	_, err := pool.Exec(ctx, migration)
	return err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func createTestTransfer(t *testing.T) *entity.Transfer {
	t.Helper()

	payee := entity.TransferPayee{
		Name:   "Jane Doe",
		TaxID:  "98765432101",
		PixKey: "jane@example.com",
		PixType: "EMAIL",
	}

	tr, err := entity.NewTransfer(
		50000,
		payee,
		entity.CashOutProviderStarkBank,
		"Test transfer",
		map[string]string{"order": "456"},
	)
	require.NoError(t, err)
	return tr
}

// ---------------------------------------------------------------------------
// Tests: Save
// ---------------------------------------------------------------------------

func TestTransferRepository_Save(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)
	transfer := createTestTransfer(t)

	err := repo.Save(context.Background(), transfer)
	require.NoError(t, err)

	saved, err := repo.FindByID(context.Background(), transfer.ID())
	require.NoError(t, err)
	assert.Equal(t, transfer.ID(), saved.ID())
	assert.Equal(t, transfer.Status(), saved.Status())
}

func TestTransferRepository_Save_Duplicate(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)
	transfer := createTestTransfer(t)

	err := repo.Save(context.Background(), transfer)
	require.NoError(t, err)

	err = repo.Save(context.Background(), transfer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao inserir transferência")
}

// ---------------------------------------------------------------------------
// Tests: FindByID
// ---------------------------------------------------------------------------

func TestTransferRepository_FindByID_Success(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)
	transfer := createTestTransfer(t)

	err := repo.Save(context.Background(), transfer)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), transfer.ID())
	require.NoError(t, err)
	assert.Equal(t, transfer.ID(), found.ID())
	assert.Equal(t, transfer.Status(), found.Status())
	assert.Equal(t, transfer.Payee().Name, found.Payee().Name)
}

func TestTransferRepository_FindByID_NotFound(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)

	_, err := repo.FindByID(context.Background(), "non-existent-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTransferNotFound)
}

// ---------------------------------------------------------------------------
// Tests: FindByExternalID
// ---------------------------------------------------------------------------

func TestTransferRepository_FindByExternalID_Success(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)
	transfer := createTestTransfer(t)

	err := repo.Save(context.Background(), transfer)
	require.NoError(t, err)

	externalID := "ext_tr_12345"
	err = transfer.Approve(externalID)
	require.NoError(t, err)
	err = repo.Update(context.Background(), transfer)
	require.NoError(t, err)

	found, err := repo.FindByExternalID(context.Background(), externalID)
	require.NoError(t, err)
	assert.Equal(t, transfer.ID(), found.ID())
	assert.Equal(t, externalID, found.ExternalID())
}

func TestTransferRepository_FindByExternalID_NotFound(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)

	_, err := repo.FindByExternalID(context.Background(), "non-existent-ext-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTransferNotFound)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestTransferRepository_Update_Success(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)
	transfer := createTestTransfer(t)

	err := repo.Save(context.Background(), transfer)
	require.NoError(t, err)

	err = transfer.Approve("ext_tr_update")
	require.NoError(t, err)
	err = repo.Update(context.Background(), transfer)
	require.NoError(t, err)

	updated, err := repo.FindByID(context.Background(), transfer.ID())
	require.NoError(t, err)
	assert.Equal(t, entity.TransferStatusApproved, updated.Status())
	assert.Equal(t, 2, updated.Version())
}

func TestTransferRepository_Update_OptimisticLock(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)
	transfer := createTestTransfer(t)

	err := repo.Save(context.Background(), transfer)
	require.NoError(t, err)

	transfer1, err := repo.FindByID(context.Background(), transfer.ID())
	require.NoError(t, err)

	transfer2, err := repo.FindByID(context.Background(), transfer.ID())
	require.NoError(t, err)

	err = transfer1.Approve("ext_1")
	require.NoError(t, err)
	err = repo.Update(context.Background(), transfer1)
	require.NoError(t, err)

	err = transfer2.Approve("ext_2")
	require.NoError(t, err)
	err = repo.Update(context.Background(), transfer2)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: ListByStatus
// ---------------------------------------------------------------------------

func TestTransferRepository_ListByStatus(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)

	for i := 0; i < 3; i++ {
		tr := createTestTransfer(t)
		err := repo.Save(context.Background(), tr)
		require.NoError(t, err)
	}

	transfers, err := repo.ListByStatus(context.Background(), entity.TransferStatusPending, 10, 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 3)

	transfers, err = repo.ListByStatus(context.Background(), entity.TransferStatusPending, 2, 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 2)

	transfers, err = repo.ListByStatus(context.Background(), entity.TransferStatusPending, 10, 10)
	require.NoError(t, err)
	assert.Len(t, transfers, 0)

	transfers, err = repo.ListByStatus(context.Background(), entity.TransferStatusApproved, 10, 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 0)
}

// ---------------------------------------------------------------------------
// Tests: ListByPayeeTaxID
// ---------------------------------------------------------------------------

func TestTransferRepository_ListByPayeeTaxID(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)

	for i := 0; i < 2; i++ {
		tr := createTestTransfer(t)
		err := repo.Save(context.Background(), tr)
		require.NoError(t, err)
	}

	transfers, err := repo.ListByPayeeTaxID(context.Background(), "98765432101", 10, 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 2)

	transfers, err = repo.ListByPayeeTaxID(context.Background(), "00000000000", 10, 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 0)
}

// ---------------------------------------------------------------------------
// Tests: ListExpired
// ---------------------------------------------------------------------------

func TestTransferRepository_ListExpired(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)

	transfer := createTestTransfer(t)
	err := repo.Save(context.Background(), transfer)
	require.NoError(t, err)

	expired, err := repo.ListExpired(context.Background(), interface{}(time.Now()), 10)
	require.NoError(t, err)
	assert.Len(t, expired, 0)

	future := time.Now().Add(72 * time.Hour)
	expired, err = repo.ListExpired(context.Background(), interface{}(future), 10)
	require.NoError(t, err)
	assert.Len(t, expired, 1)
}

// ---------------------------------------------------------------------------
// Tests: CountByStatus
// ---------------------------------------------------------------------------

func TestTransferRepository_CountByStatus(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)

	for i := 0; i < 3; i++ {
		tr := createTestTransfer(t)
		err := repo.Save(context.Background(), tr)
		require.NoError(t, err)
	}

	count, err := repo.CountByStatus(context.Background(), entity.TransferStatusPending)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = repo.CountByStatus(context.Background(), entity.TransferStatusApproved)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ---------------------------------------------------------------------------
// Tests: Full CRUD flow
// ---------------------------------------------------------------------------

func TestTransferRepository_FullCRUD(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewTransferRepository(pool)

	transfer := createTestTransfer(t)
	err := repo.Save(context.Background(), transfer)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), transfer.ID())
	require.NoError(t, err)
	assert.Equal(t, entity.TransferStatusPending, found.Status())

	err = found.Approve("ext_full_crud")
	require.NoError(t, err)
	err = repo.Update(context.Background(), found)
	require.NoError(t, err)

	approved, err := repo.FindByID(context.Background(), transfer.ID())
	require.NoError(t, err)
	assert.Equal(t, entity.TransferStatusApproved, approved.Status())
	assert.Equal(t, 2, approved.Version())

	transfers, err := repo.ListByStatus(context.Background(), entity.TransferStatusApproved, 10, 0)
	require.NoError(t, err)
	assert.Len(t, transfers, 1)

	count, err := repo.CountByStatus(context.Background(), entity.TransferStatusApproved)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}