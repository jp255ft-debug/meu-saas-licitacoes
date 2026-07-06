//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain"
	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
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
			"POSTGRES_DB":       "cashin_test",
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

	connString := fmt.Sprintf("postgres://test:test@%s:%s/cashin_test?sslmode=disable", host, port.Port())

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
	CREATE TABLE IF NOT EXISTS invoices (
		id VARCHAR(64) PRIMARY KEY,
		external_id VARCHAR(128),
		invoice_number VARCHAR(128) NOT NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
		amount_cents BIGINT NOT NULL,
		due_date TIMESTAMPTZ NOT NULL,
		payer_name VARCHAR(255) NOT NULL,
		payer_tax_id VARCHAR(32) NOT NULL,
		payer_email VARCHAR(255),
		payer_phone VARCHAR(32),
		payer_address_line1 VARCHAR(255),
		payer_address_line2 VARCHAR(255),
		payer_city VARCHAR(128),
		payer_state VARCHAR(32),
		payer_zip_code VARCHAR(16),
		provider VARCHAR(32) NOT NULL,
		pix_qr_code TEXT,
		pix_key VARCHAR(128),
		boleto_barcode VARCHAR(64),
		boleto_digitable VARCHAR(64),
		metadata JSONB DEFAULT '{}',
		paid_amount_cents BIGINT NOT NULL DEFAULT 0,
		paid_at TIMESTAMPTZ,
		payment_method VARCHAR(32),
		cancelled_at TIMESTAMPTZ,
		failure_reason TEXT,
		version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
	CREATE INDEX IF NOT EXISTS idx_invoices_external_id ON invoices(external_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_payer_tax_id ON invoices(payer_tax_id);
	`

	_, err := pool.Exec(ctx, migration)
	return err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func createTestInvoice(t *testing.T) *entity.Invoice {
	t.Helper()

	payer := entity.InvoicePayer{
		Name:  "John Doe",
		TaxID: "12345678901",
		Email: "john@example.com",
	}

	inv, err := entity.NewInvoice(
		10000,
		time.Now().Add(72*time.Hour),
		payer,
		entity.CashInProviderAsaas,
		map[string]string{"order": "123"},
	)
	require.NoError(t, err)
	return inv
}

// ---------------------------------------------------------------------------
// Tests: Save
// ---------------------------------------------------------------------------

func TestInvoiceRepository_Save(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)
	invoice := createTestInvoice(t)

	err := repo.Save(context.Background(), invoice)
	require.NoError(t, err)

	saved, err := repo.FindByID(context.Background(), invoice.ID())
	require.NoError(t, err)
	assert.Equal(t, invoice.ID(), saved.ID())
	assert.Equal(t, invoice.Status(), saved.Status())
}

func TestInvoiceRepository_Save_Duplicate(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)
	invoice := createTestInvoice(t)

	err := repo.Save(context.Background(), invoice)
	require.NoError(t, err)

	err = repo.Save(context.Background(), invoice)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao inserir fatura")
}

// ---------------------------------------------------------------------------
// Tests: FindByID
// ---------------------------------------------------------------------------

func TestInvoiceRepository_FindByID_Success(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)
	invoice := createTestInvoice(t)

	err := repo.Save(context.Background(), invoice)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), invoice.ID())
	require.NoError(t, err)
	assert.Equal(t, invoice.ID(), found.ID())
	assert.Equal(t, invoice.Status(), found.Status())
	assert.Equal(t, invoice.Payer().Name, found.Payer().Name)
}

func TestInvoiceRepository_FindByID_NotFound(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)

	_, err := repo.FindByID(context.Background(), "non-existent-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvoiceNotFound)
}

// ---------------------------------------------------------------------------
// Tests: FindByExternalID
// ---------------------------------------------------------------------------

func TestInvoiceRepository_FindByExternalID_Success(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)
	invoice := createTestInvoice(t)

	err := repo.Save(context.Background(), invoice)
	require.NoError(t, err)

	externalID := "ext_inv_12345"
	err = invoice.Issue(externalID, "pixqrcode:test", "pixkey:test")
	require.NoError(t, err)
	err = repo.Update(context.Background(), invoice)
	require.NoError(t, err)

	found, err := repo.FindByExternalID(context.Background(), externalID)
	require.NoError(t, err)
	assert.Equal(t, invoice.ID(), found.ID())
	assert.Equal(t, externalID, found.ExternalID())
}

func TestInvoiceRepository_FindByExternalID_NotFound(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)

	_, err := repo.FindByExternalID(context.Background(), "non-existent-ext-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvoiceNotFound)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestInvoiceRepository_Update_Success(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)
	invoice := createTestInvoice(t)

	err := repo.Save(context.Background(), invoice)
	require.NoError(t, err)

	err = invoice.Issue("ext_inv_update", "qrcode", "pixkey")
	require.NoError(t, err)
	err = repo.Update(context.Background(), invoice)
	require.NoError(t, err)

	updated, err := repo.FindByID(context.Background(), invoice.ID())
	require.NoError(t, err)
	assert.Equal(t, entity.InvoiceStatusIssued, updated.Status())
	assert.Equal(t, 2, updated.Version())
}

func TestInvoiceRepository_Update_OptimisticLock(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)
	invoice := createTestInvoice(t)

	err := repo.Save(context.Background(), invoice)
	require.NoError(t, err)

	invoice1, err := repo.FindByID(context.Background(), invoice.ID())
	require.NoError(t, err)

	invoice2, err := repo.FindByID(context.Background(), invoice.ID())
	require.NoError(t, err)

	err = invoice1.Issue("ext_1", "qrcode_1", "pixkey_1")
	require.NoError(t, err)
	err = repo.Update(context.Background(), invoice1)
	require.NoError(t, err)

	err = invoice2.Issue("ext_2", "qrcode_2", "pixkey_2")
	require.NoError(t, err)
	err = repo.Update(context.Background(), invoice2)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: ListByStatus
// ---------------------------------------------------------------------------

func TestInvoiceRepository_ListByStatus(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)

	for i := 0; i < 3; i++ {
		inv := createTestInvoice(t)
		err := repo.Save(context.Background(), inv)
		require.NoError(t, err)
	}

	invoices, err := repo.ListByStatus(context.Background(), entity.InvoiceStatusPending, 10, 0)
	require.NoError(t, err)
	assert.Len(t, invoices, 3)

	invoices, err = repo.ListByStatus(context.Background(), entity.InvoiceStatusPending, 2, 0)
	require.NoError(t, err)
	assert.Len(t, invoices, 2)

	invoices, err = repo.ListByStatus(context.Background(), entity.InvoiceStatusPending, 10, 10)
	require.NoError(t, err)
	assert.Len(t, invoices, 0)

	invoices, err = repo.ListByStatus(context.Background(), entity.InvoiceStatusIssued, 10, 0)
	require.NoError(t, err)
	assert.Len(t, invoices, 0)
}

// ---------------------------------------------------------------------------
// Tests: ListByPayerTaxID
// ---------------------------------------------------------------------------

func TestInvoiceRepository_ListByPayerTaxID(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)

	for i := 0; i < 2; i++ {
		inv := createTestInvoice(t)
		err := repo.Save(context.Background(), inv)
		require.NoError(t, err)
	}

	invoices, err := repo.ListByPayerTaxID(context.Background(), "12345678901", 10, 0)
	require.NoError(t, err)
	assert.Len(t, invoices, 2)

	invoices, err = repo.ListByPayerTaxID(context.Background(), "00000000000", 10, 0)
	require.NoError(t, err)
	assert.Len(t, invoices, 0)
}

// ---------------------------------------------------------------------------
// Tests: ListExpired
// ---------------------------------------------------------------------------

func TestInvoiceRepository_ListExpired(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)

	invoice := createTestInvoice(t)
	err := repo.Save(context.Background(), invoice)
	require.NoError(t, err)

	expired, err := repo.ListExpired(context.Background(), interface{}(time.Now()), 10)
	require.NoError(t, err)
	assert.Len(t, expired, 0)

	future := time.Now().Add(96 * time.Hour)
	expired, err = repo.ListExpired(context.Background(), interface{}(future), 10)
	require.NoError(t, err)
	assert.Len(t, expired, 1)
}

// ---------------------------------------------------------------------------
// Tests: CountByStatus
// ---------------------------------------------------------------------------

func TestInvoiceRepository_CountByStatus(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)

	for i := 0; i < 3; i++ {
		inv := createTestInvoice(t)
		err := repo.Save(context.Background(), inv)
		require.NoError(t, err)
	}

	count, err := repo.CountByStatus(context.Background(), entity.InvoiceStatusPending)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = repo.CountByStatus(context.Background(), entity.InvoiceStatusIssued)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ---------------------------------------------------------------------------
// Tests: Full CRUD flow
// ---------------------------------------------------------------------------

func TestInvoiceRepository_FullCRUD(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewInvoiceRepository(pool)

	invoice := createTestInvoice(t)
	err := repo.Save(context.Background(), invoice)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), invoice.ID())
	require.NoError(t, err)
	assert.Equal(t, entity.InvoiceStatusPending, found.Status())

	err = found.Issue("ext_full_crud", "pixqrcode:full", "pixkey:full")
	require.NoError(t, err)
	err = repo.Update(context.Background(), found)
	require.NoError(t, err)

	issued, err := repo.FindByID(context.Background(), invoice.ID())
	require.NoError(t, err)
	assert.Equal(t, entity.InvoiceStatusIssued, issued.Status())
	assert.Equal(t, 2, issued.Version())

	invoices, err := repo.ListByStatus(context.Background(), entity.InvoiceStatusIssued, 10, 0)
	require.NoError(t, err)
	assert.Len(t, invoices, 1)

	count, err := repo.CountByStatus(context.Background(), entity.InvoiceStatusIssued)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}