//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/payment-core/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// ---------------------------------------------------------------------------
// Setup / Teardown
// ---------------------------------------------------------------------------

// startPaymentPostgresContainer inicia um container PostgreSQL via Testcontainers.
func startPaymentPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("SKIP_INTEGRATION está definido — pulando teste de integração")
	}

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "payment_test",
		},
		Cmd: []string{"postgres", "-c", "fsync=off"},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("Docker não disponível ou Testcontainers falhou: %v — pulando teste de integração", err)
	}

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connString := fmt.Sprintf("postgres://test:test@%s:%s/payment_test?sslmode=disable", host, port.Port())

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
	require.NoError(t, err)

	if err := runPaymentMigrations(ctx, pool); err != nil {
		pool.Close()
		container.Terminate(ctx)
		t.Fatalf("falha ao rodar migrations: %v", err)
	}

	cleanup := func() {
		pool.Close()
		container.Terminate(ctx)
	}

	return pool, cleanup
}

// runPaymentMigrations executa o script SQL de criação da tabela payments.
func runPaymentMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migration := `
	CREATE TABLE IF NOT EXISTS payments (
		id VARCHAR(64) PRIMARY KEY,
		external_id VARCHAR(128),
		status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
		payment_type VARCHAR(16) NOT NULL,
		method VARCHAR(16) NOT NULL,
		amount_cents BIGINT NOT NULL,
		fee_cents BIGINT NOT NULL DEFAULT 0,
		net_amount_cents BIGINT NOT NULL DEFAULT 0,
		payer_name VARCHAR(255) NOT NULL,
		payer_tax_id VARCHAR(32) NOT NULL,
		payer_email VARCHAR(255),
		payer_pix_key VARCHAR(128),
		payer_pix_type VARCHAR(16),
		payee_name VARCHAR(255) NOT NULL,
		payee_tax_id VARCHAR(32) NOT NULL,
		payee_pix_key VARCHAR(128),
		payee_pix_type VARCHAR(16),
		payee_bank_code VARCHAR(16),
		payee_branch VARCHAR(16),
		payee_account VARCHAR(32),
		payee_account_type VARCHAR(16),
		description TEXT NOT NULL DEFAULT '',
		expires_at TIMESTAMPTZ,
		processed_at TIMESTAMPTZ,
		settled_at TIMESTAMPTZ,
		failed_at TIMESTAMPTZ,
		cancelled_at TIMESTAMPTZ,
		failure_reason TEXT,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		version INT NOT NULL DEFAULT 1
	);

	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
	CREATE INDEX IF NOT EXISTS idx_payments_external_id ON payments(external_id);
	CREATE INDEX IF NOT EXISTS idx_payments_payer_tax_id ON payments(payer_tax_id);
	`

	_, err := pool.Exec(ctx, migration)
	return err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func createTestPayment(t *testing.T) *domain.Payment {
	t.Helper()

	amount, err := domain.NewMoney(10000)
	require.NoError(t, err)

	payer := domain.Payer{
		Name:   "João Silva",
		TaxID:  "52998224725",
		Email:  "joao@email.com",
		PixKey: domain.MustNewPixKey("joao@email.com"),
	}

	payee := domain.Payee{
		Name:   "OrkaPay LTDA",
		TaxID:  "11222333000181",
		PixKey: domain.MustNewPixKey("11222333000181"),
	}

	payment, err := domain.NewPayment(
		domain.PaymentTypePix,
		domain.PaymentMethodPixQRCode,
		amount,
		payer,
		payee,
		"Pagamento teste integração",
		map[string]string{"test": "true"},
	)
	require.NoError(t, err)
	return payment
}

// ---------------------------------------------------------------------------
// Tests: Save
// ---------------------------------------------------------------------------

func TestPaymentRepository_Save(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	saved, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, payment.ID(), saved.ID())
	assert.Equal(t, payment.Status(), saved.Status())
	assert.Equal(t, payment.Amount().Cents(), saved.Amount().Cents())
}

func TestPaymentRepository_Save_Duplicate(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	err = repo.Save(context.Background(), payment)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao inserir pagamento")
}

// ---------------------------------------------------------------------------
// Tests: FindByID
// ---------------------------------------------------------------------------

func TestPaymentRepository_FindByID_Success(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, payment.ID(), found.ID())
	assert.Equal(t, payment.Status(), found.Status())
	assert.Equal(t, payment.Payer().Name, found.Payer().Name)
}

func TestPaymentRepository_FindByID_NotFound(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	_, err := repo.FindByID(context.Background(), "non-existent-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

// ---------------------------------------------------------------------------
// Tests: FindByExternalID
// ---------------------------------------------------------------------------

func TestPaymentRepository_FindByExternalID_Success(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	extID := "ext_pay_12345"
	err = payment.Approve(extID)
	require.NoError(t, err)
	err = repo.Update(context.Background(), payment)
	require.NoError(t, err)

	found, err := repo.FindByExternalID(context.Background(), extID)
	require.NoError(t, err)
	assert.Equal(t, payment.ID(), found.ID())
	assert.Equal(t, extID, found.ExternalID())
}

func TestPaymentRepository_FindByExternalID_NotFound(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	_, err := repo.FindByExternalID(context.Background(), "non-existent-ext-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

// ---------------------------------------------------------------------------
// Tests: Update
// ---------------------------------------------------------------------------

func TestPaymentRepository_Update_Success(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	err = payment.Approve("ext_pay_update")
	require.NoError(t, err)
	err = repo.Update(context.Background(), payment)
	require.NoError(t, err)

	updated, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusApproved, updated.Status())
	assert.Equal(t, "ext_pay_update", updated.ExternalID())
	assert.Equal(t, 2, updated.Version())
}

func TestPaymentRepository_Update_OptimisticLock(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	payment1, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)

	payment2, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)

	err = payment1.Approve("ext_lock_1")
	require.NoError(t, err)
	err = repo.Update(context.Background(), payment1)
	require.NoError(t, err)

	err = payment2.Approve("ext_lock_2")
	require.NoError(t, err)
	err = repo.Update(context.Background(), payment2)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrPaymentAlreadyProcessed)
}

// ---------------------------------------------------------------------------
// Tests: ListByStatus
// ---------------------------------------------------------------------------

func TestPaymentRepository_ListByStatus(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	for i := 0; i < 3; i++ {
		p := createTestPayment(t)
		err := repo.Save(context.Background(), p)
		require.NoError(t, err)
	}

	payments, err := repo.ListByStatus(context.Background(), domain.PaymentStatusPending, 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 3)

	payments, err = repo.ListByStatus(context.Background(), domain.PaymentStatusPending, 2, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 2)

	payments, err = repo.ListByStatus(context.Background(), domain.PaymentStatusPending, 10, 10)
	require.NoError(t, err)
	assert.Len(t, payments, 0)

	payments, err = repo.ListByStatus(context.Background(), domain.PaymentStatusApproved, 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 0)
}

// ---------------------------------------------------------------------------
// Tests: ListByPayerTaxID
// ---------------------------------------------------------------------------

func TestPaymentRepository_ListByPayerTaxID(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	for i := 0; i < 2; i++ {
		p := createTestPayment(t)
		err := repo.Save(context.Background(), p)
		require.NoError(t, err)
	}

	payments, err := repo.ListByPayerTaxID(context.Background(), "52998224725", 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 2)

	payments, err = repo.ListByPayerTaxID(context.Background(), "00000000000", 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 0)
}

// ---------------------------------------------------------------------------
// Tests: ListExpired
// ---------------------------------------------------------------------------

func TestPaymentRepository_ListExpired(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	payment := createTestPayment(t)
	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	expired, err := repo.ListExpired(context.Background(), time.Now(), 10)
	require.NoError(t, err)
	assert.Len(t, expired, 0)
}

// ---------------------------------------------------------------------------
// Tests: CountByStatus
// ---------------------------------------------------------------------------

func TestPaymentRepository_CountByStatus(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	for i := 0; i < 3; i++ {
		p := createTestPayment(t)
		err := repo.Save(context.Background(), p)
		require.NoError(t, err)
	}

	count, err := repo.CountByStatus(context.Background(), domain.PaymentStatusPending)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = repo.CountByStatus(context.Background(), domain.PaymentStatusApproved)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ---------------------------------------------------------------------------
// Tests: Full CRUD flow
// ---------------------------------------------------------------------------

func TestPaymentRepository_FullCRUD(t *testing.T) {
	pool, cleanup := startPaymentPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	payment := createTestPayment(t)
	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusPending, found.Status())

	err = found.Approve("ext_full_crud")
	require.NoError(t, err)
	err = repo.Update(context.Background(), found)
	require.NoError(t, err)

	approved, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusApproved, approved.Status())
	assert.Equal(t, "ext_full_crud", approved.ExternalID())
	assert.Equal(t, 2, approved.Version())

	payments, err := repo.ListByStatus(context.Background(), domain.PaymentStatusApproved, 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 1)

	count, err := repo.CountByStatus(context.Background(), domain.PaymentStatusApproved)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}