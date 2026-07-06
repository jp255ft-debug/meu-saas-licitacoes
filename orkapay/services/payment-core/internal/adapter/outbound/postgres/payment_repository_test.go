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

// startPostgresContainer inicia um container PostgreSQL via Testcontainers.
// Retorna o pool de conexão e uma função de cleanup.
// Se o Docker não estiver disponível, pula o teste.
func startPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	// Verifica se o Docker está disponível
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("SKIP_INTEGRATION está definido — pulando teste de integração")
	}

	// Tenta usar Testcontainers
	pool, cleanup, err := startTestContainer(t)
	if err != nil {
		t.Skipf("Docker não disponível ou Testcontainers falhou: %v — pulando teste de integração", err)
	}

	return pool, cleanup
}

// startTestContainer tenta iniciar o container PostgreSQL via Testcontainers.
func startTestContainer(t *testing.T) (*pgxpool.Pool, func(), error) {
	ctx := context.Background()

	// Requisitar imagem postgres:14-alpine
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "paymentcore_test",
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

	// Obter connection string
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

	connString := fmt.Sprintf("postgres://test:test@%s:%s/paymentcore_test?sslmode=disable", host, port.Port())

	// Conectar com retry
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

	// Rodar migrations
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
	CREATE TABLE IF NOT EXISTS payments (
		id              VARCHAR(64) PRIMARY KEY,
		external_id     VARCHAR(128),
		status          VARCHAR(32) NOT NULL,
		payment_type    VARCHAR(32) NOT NULL,
		method          VARCHAR(32) NOT NULL,
		amount_cents    BIGINT NOT NULL,
		fee_cents       BIGINT NOT NULL DEFAULT 0,
		net_amount_cents BIGINT NOT NULL,
		payer_name      VARCHAR(255),
		payer_tax_id    VARCHAR(32) NOT NULL,
		payer_email     VARCHAR(255),
		payer_pix_key   VARCHAR(128),
		payer_pix_type  VARCHAR(16),
		payee_name      VARCHAR(255),
		payee_tax_id    VARCHAR(32) NOT NULL,
		payee_pix_key   VARCHAR(128),
		payee_pix_type  VARCHAR(16),
		payee_bank_code VARCHAR(16),
		payee_branch    VARCHAR(16),
		payee_account   VARCHAR(32),
		payee_account_type VARCHAR(16),
		description     TEXT,
		expires_at      TIMESTAMPTZ,
		processed_at    TIMESTAMPTZ,
		settled_at      TIMESTAMPTZ,
		failed_at       TIMESTAMPTZ,
		cancelled_at    TIMESTAMPTZ,
		failure_reason  TEXT,
		metadata        JSONB DEFAULT '{}',
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		version         INT NOT NULL DEFAULT 1
	);

	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
	CREATE INDEX IF NOT EXISTS idx_payments_external_id ON payments(external_id);
	CREATE INDEX IF NOT EXISTS idx_payments_payer_tax_id ON payments(payer_tax_id);
	CREATE INDEX IF NOT EXISTS idx_payments_expires_at ON payments(expires_at) WHERE status = 'PENDING';
	CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);
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
	payerPixKey := domain.MustNewPixKey("test@example.com")
	payeePixKey := domain.MustNewPixKey("payee@example.com")

	payment, err := domain.NewPayment(
		domain.PaymentTypePix,
		domain.PaymentMethodPixQRCode,
		amount,
		domain.Payer{
			Name:   "John Doe",
			TaxID:  "12345678901",
			Email:  "john@example.com",
			PixKey: payerPixKey,
		},
		domain.Payee{
			Name:   "Jane Doe",
			TaxID:  "98765432101",
			PixKey: payeePixKey,
		},
		"Test payment",
		nil,
	)
	require.NoError(t, err)
	return payment
}

func createExpiredPayment(t *testing.T) *domain.Payment {
	t.Helper()
	payment := createTestPayment(t)
	// Força expiração no passado — o domain.NewPayment define expires_at como
	// 24h a partir de agora, então precisamos de um pagamento que já expirou.
	// Como o domínio não permite criar pagamentos expirados, vamos usar um
	// pagamento normal e verificar se ele NÃO está na lista de expirados.
	return payment
}

// ---------------------------------------------------------------------------
// Tests: Save
// ---------------------------------------------------------------------------

func TestPaymentRepository_Save(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	// Verificar que foi salvo buscando por ID
	saved, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, payment.ID(), saved.ID())
	assert.Equal(t, payment.Status(), saved.Status())
	assert.Equal(t, payment.Amount().Cents(), saved.Amount().Cents())
}

func TestPaymentRepository_Save_Duplicate(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	// Salvar novamente com o mesmo ID deve falhar (PK violation)
	err = repo.Save(context.Background(), payment)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao inserir pagamento")
}

// ---------------------------------------------------------------------------
// Tests: FindByID
// ---------------------------------------------------------------------------

func TestPaymentRepository_FindByID_Success(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
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
	assert.Equal(t, payment.Payee().Name, found.Payee().Name)
}

func TestPaymentRepository_FindByID_NotFound(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
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
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	// O pagamento criado não tem external_id, então precisamos atualizar
	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	// Atualizar com external_id via Approve (que define externalID)
	externalID := "ext_123456"
	err = payment.Approve(externalID)
	require.NoError(t, err)
	err = repo.Update(context.Background(), payment)
	require.NoError(t, err)

	found, err := repo.FindByExternalID(context.Background(), externalID)
	require.NoError(t, err)
	assert.Equal(t, payment.ID(), found.ID())
	assert.Equal(t, externalID, found.ExternalID())
}

func TestPaymentRepository_FindByExternalID_NotFound(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
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
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	// Atualizar status
	payment.Cancel("customer request")
	err = repo.Update(context.Background(), payment)
	require.NoError(t, err)

	// Verificar
	updated, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusCancelled, updated.Status())
	assert.Equal(t, 2, updated.Version()) // versão incrementada
}

func TestPaymentRepository_Update_OptimisticLock(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)
	payment := createTestPayment(t)

	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	// Simular concorrência: carregar duas vezes, atualizar uma, tentar atualizar a outra
	payment1, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)

	payment2, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)

	// Primeira atualização funciona
	payment1.Cancel("reason 1")
	err = repo.Update(context.Background(), payment1)
	require.NoError(t, err)

	// Segunda atualização com versão desatualizada deve falhar
	payment2.Cancel("reason 2")
	err = repo.Update(context.Background(), payment2)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrPaymentAlreadyProcessed)
}

// ---------------------------------------------------------------------------
// Tests: ListByStatus
// ---------------------------------------------------------------------------

func TestPaymentRepository_ListByStatus(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	// Criar 3 pagamentos PENDING
	for i := 0; i < 3; i++ {
		p := createTestPayment(t)
		err := repo.Save(context.Background(), p)
		require.NoError(t, err)
	}

	// Listar PENDING
	payments, err := repo.ListByStatus(context.Background(), domain.PaymentStatusPending, 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 3)

	// Listar com limit 2
	payments, err = repo.ListByStatus(context.Background(), domain.PaymentStatusPending, 2, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 2)

	// Listar com offset
	payments, err = repo.ListByStatus(context.Background(), domain.PaymentStatusPending, 10, 10)
	require.NoError(t, err)
	assert.Len(t, payments, 0)

	// Listar status sem pagamentos
	payments, err = repo.ListByStatus(context.Background(), domain.PaymentStatusApproved, 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 0)
}

// ---------------------------------------------------------------------------
// Tests: ListByPayerTaxID
// ---------------------------------------------------------------------------

func TestPaymentRepository_ListByPayerTaxID(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	// Criar 2 pagamentos para o mesmo pagador
	taxID := "12345678901"
	for i := 0; i < 2; i++ {
		p := createTestPayment(t)
		err := repo.Save(context.Background(), p)
		require.NoError(t, err)
	}

	payments, err := repo.ListByPayerTaxID(context.Background(), taxID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 2)

	// TaxID sem pagamentos
	payments, err = repo.ListByPayerTaxID(context.Background(), "00000000000", 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 0)
}

// ---------------------------------------------------------------------------
// Tests: ListExpired
// ---------------------------------------------------------------------------

func TestPaymentRepository_ListExpired(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	// Criar um pagamento PENDING (expira em 24h)
	payment := createTestPayment(t)
	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	// Listar expirados antes de agora — não deve encontrar (expira em 24h)
	expired, err := repo.ListExpired(context.Background(), time.Now(), 10)
	require.NoError(t, err)
	assert.Len(t, expired, 0)

	// Listar expirados depois de 48h — deve encontrar
	future := time.Now().Add(48 * time.Hour)
	expired, err = repo.ListExpired(context.Background(), future, 10)
	require.NoError(t, err)
	assert.Len(t, expired, 1)
}

// ---------------------------------------------------------------------------
// Tests: CountByStatus
// ---------------------------------------------------------------------------

func TestPaymentRepository_CountByStatus(t *testing.T) {
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	// Criar 3 pagamentos PENDING
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
	pool, cleanup := startPostgresContainer(t)
	defer cleanup()

	repo := NewPaymentRepository(pool)

	// 1. Create
	payment := createTestPayment(t)
	err := repo.Save(context.Background(), payment)
	require.NoError(t, err)

	// 2. Read
	found, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusPending, found.Status())

	// 3. Update — aprovar pagamento (via Approve, que é o método real do domínio)
	externalID := "ext_999"
	err = found.Approve(externalID)
	require.NoError(t, err)
	err = repo.Update(context.Background(), found)
	require.NoError(t, err)

	// Verificar atualização
	approved, err := repo.FindByID(context.Background(), payment.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusApproved, approved.Status())
	assert.Equal(t, 2, approved.Version())

	// 4. List por status
	payments, err := repo.ListByStatus(context.Background(), domain.PaymentStatusApproved, 10, 0)
	require.NoError(t, err)
	assert.Len(t, payments, 1)

	// 5. Count
	count, err := repo.CountByStatus(context.Background(), domain.PaymentStatusApproved)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
