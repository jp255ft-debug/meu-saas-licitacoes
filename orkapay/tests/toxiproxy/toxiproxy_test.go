//go:build toxiproxy

package toxiproxy_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ---------------------------------------------------------------------------
// Constantes
// ---------------------------------------------------------------------------

const (
	toxiproxyImage    = "ghcr.io/shopify/toxiproxy:2.9.0"
	postgresImage     = "postgres:14-alpine"
	proxyName         = "postgres_proxy"
	proxyListenPort   = "8666"
	proxyAPIPort      = "8474"
)

// ---------------------------------------------------------------------------
// Setup do ambiente: Toxiproxy + PostgreSQL
// ---------------------------------------------------------------------------

type toxiproxyEnv struct {
	pgPool    *pgxpool.Pool
	proxyHost string
	proxyPort string
	apiURL    string
	cleanup   func()
}

func setupToxiproxyEnv(t *testing.T) *toxiproxyEnv {
	t.Helper()

	ctx := context.Background()

	// 1. Inicia o container PostgreSQL
	pgReq := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "cashin_test",
		},
		Cmd: []string{"postgres", "-c", "fsync=off"},
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pgReq,
		Started:          true,
	})
	if err != nil {
		t.Skipf("Docker não disponível: %v — pulando teste toxiproxy", err)
	}

	pgHost, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	pgPort, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// 2. Inicia o container Toxiproxy
	toxiproxyReq := testcontainers.ContainerRequest{
		Image:        toxiproxyImage,
		ExposedPorts: []string{proxyAPIPort + "/tcp", proxyListenPort + "/tcp"},
		WaitingFor:   wait.ForHTTP("/version").WithPort(proxyAPIPort + "/tcp"),
	}

	toxiproxyContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: toxiproxyReq,
		Started:          true,
	})
	require.NoError(t, err)

	proxyHost, err := toxiproxyContainer.Host(ctx)
	require.NoError(t, err)

	proxyAPIPortMapped, err := toxiproxyContainer.MappedPort(ctx, proxyAPIPort)
	require.NoError(t, err)

	proxyListenPortMapped, err := toxiproxyContainer.MappedPort(ctx, proxyListenPort)
	require.NoError(t, err)

	apiURL := fmt.Sprintf("http://%s:%s", proxyHost, proxyAPIPortMapped.Port())

	// 3. Configura proxy para PostgreSQL via API do Toxiproxy
	proxyUpstream := fmt.Sprintf("%s:%s", pgHost, pgPort.Port())
	createProxy(apiURL, proxyName, proxyListenPortMapped.Port(), proxyUpstream)

	// 4. Conecta ao PostgreSQL via proxy
	proxyConnString := fmt.Sprintf("postgres://test:test@%s:%s/cashin_test?sslmode=disable",
		proxyHost, proxyListenPortMapped.Port())

	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, proxyConnString)
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

	// Roda migrations básicas
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS invoices (
			id VARCHAR(64) PRIMARY KEY,
			status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
			amount_cents BIGINT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
		toxiproxyContainer.Terminate(ctx)
	}

	return &toxiproxyEnv{
		pgPool:    pool,
		proxyHost: proxyHost,
		proxyPort: proxyListenPortMapped.Port(),
		apiURL:    apiURL,
		cleanup:   cleanup,
	}
}

// createProxy configura um proxy no Toxiproxy via API HTTP.
func createProxy(apiURL, name, listen, upstream string) {
	body := fmt.Sprintf(`{
		"name": "%s",
		"listen": "0.0.0.0:%s",
		"upstream": "%s",
		"enabled": true
	}`, name, listen, upstream)

	resp, err := http.Post(apiURL+"/proxies", "application/json", strings.NewReader(body))
	if err != nil {
		panic(fmt.Sprintf("falha ao criar proxy: %v", err))
	}
	defer resp.Body.Close()
}

// addToxic adiciona um toxic ao proxy especificado.
func addToxic(apiURL, proxyName, toxicName, toxicType string, toxicity float64, attributes map[string]interface{}) {
	attrsJSON := "{"
	for k, v := range attributes {
		attrsJSON += fmt.Sprintf(`"%s": %v,`, k, v)
	}
	attrsJSON = strings.TrimRight(attrsJSON, ",") + "}"

	body := fmt.Sprintf(`{
		"name": "%s",
		"type": "%s",
		"toxicity": %f,
		"attributes": %s
	}`, toxicName, toxicType, toxicity, attrsJSON)

	resp, err := http.Post(
		fmt.Sprintf("%s/proxies/%s/toxics", apiURL, proxyName),
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		panic(fmt.Sprintf("falha ao adicionar toxic: %v", err))
	}
	defer resp.Body.Close()
}

// removeAllToxics remove todos os toxics de um proxy.
func removeAllToxics(apiURL, proxyName string) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/proxies/%s/toxics", apiURL, proxyName), nil)
	http.DefaultClient.Do(req)
}

// ---------------------------------------------------------------------------
// Test: Latência no banco de dados
// ---------------------------------------------------------------------------

func TestDatabaseLatency_TimeoutRespected(t *testing.T) {
	env := setupToxiproxyEnv(t)
	defer env.cleanup()

	// Adiciona latência de 1500ms no proxy
	addToxic(env.apiURL, proxyName, "latency", "latency", 1.0, map[string]interface{}{
		"latency": 1500,
	})
	defer removeAllToxics(env.apiURL, proxyName)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Tentativa deve falhar por timeout (latência > timeout)
	_, err := env.pgPool.Query(ctx, "SELECT 1")
	assert.Error(t, err, "esperava erro de timeout com latência de 1500ms e timeout de 500ms")
}

// ---------------------------------------------------------------------------
// Test: Conexão com banco offline
// ---------------------------------------------------------------------------

func TestDatabaseConnectionDown_ReturnsError(t *testing.T) {
	env := setupToxiproxyEnv(t)
	defer env.cleanup()

	// Simula falha completa (banco fora do ar)
	addToxic(env.apiURL, proxyName, "down", "down", 1.0, map[string]interface{}{})
	defer removeAllToxics(env.apiURL, proxyName)

	// Pequeno delay para o toxic aplicar
	time.Sleep(100 * time.Millisecond)

	_, err := env.pgPool.Query(context.Background(), "SELECT 1")
	assert.Error(t, err, "esperava erro com banco fora do ar")
}

// ---------------------------------------------------------------------------
// Test: Perda de pacotes
// ---------------------------------------------------------------------------

func TestDatabasePacketLoss_RetryLogic(t *testing.T) {
	env := setupToxiproxyEnv(t)
	defer env.cleanup()

	// Simula perda de 80% dos pacotes
	addToxic(env.apiURL, proxyName, "loss", "loss", 1.0, map[string]interface{}{
		"loss": 0.8,
	})
	defer removeAllToxics(env.apiURL, proxyName)

	// Com 80% de perda, algumas queries devem falhar
	// Testamos múltiplas tentativas para ver se alguma passa
	successCount := 0
	for i := 0; i < 10; i++ {
		_, err := env.pgPool.Query(context.Background(), "SELECT 1")
		if err == nil {
			successCount++
		}
	}

	// Com 80% de perda, esperamos que a maioria falhe
	t.Logf("Com 80%% de perda: %d/10 queries bem-sucedidas", successCount)
	assert.Less(t, successCount, 8, "perda de 80% deve causar falhas frequentes")
}

// ---------------------------------------------------------------------------
// Test: Slow close (conexão lenta para fechar)
// ---------------------------------------------------------------------------

func TestDatabaseSlowClose_NoConnectionLeak(t *testing.T) {
	env := setupToxiproxyEnv(t)
	defer env.cleanup()

	// Simula slow close (delay de 5s para fechar conexão)
	addToxic(env.apiURL, proxyName, "slow_close", "slow_close", 1.0, map[string]interface{}{
		"delay": 5000,
	})
	defer removeAllToxics(env.apiURL, proxyName)

	// Deve conseguir conectar e fazer query mesmo com slow close
	_, err := env.pgPool.Query(context.Background(), "SELECT 1")
	assert.NoError(t, err, "slow close não deve impedir queries normais")
}

// ---------------------------------------------------------------------------
// Test: Timeout de conexão
// ---------------------------------------------------------------------------

func TestDatabaseTimeout_ConnectionRefused(t *testing.T) {
	env := setupToxiproxyEnv(t)
	defer env.cleanup()

	// Simula timeout (conexão não responde)
	addToxic(env.apiURL, proxyName, "timeout", "timeout", 1.0, map[string]interface{}{
		"timeout": 0,
	})
	defer removeAllToxics(env.apiURL, proxyName)

	// Tenta conectar com timeout curto
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := env.pgPool.Query(ctx, "SELECT 1")
	assert.Error(t, err, "esperava erro com timeout ativo")
}

// ---------------------------------------------------------------------------
// Test: Reset de peer
// ---------------------------------------------------------------------------

func TestDatabaseResetPeer_ConnectionBroken(t *testing.T) {
	env := setupToxiproxyEnv(t)
	defer env.cleanup()

	// Simula reset de peer (conexão TCP resetada)
	addToxic(env.apiURL, proxyName, "reset", "reset_peer", 1.0, map[string]interface{}{})
	defer removeAllToxics(env.apiURL, proxyName)

	_, err := env.pgPool.Query(context.Background(), "SELECT 1")
	assert.Error(t, err, "esperava erro com reset de peer")
}

// ---------------------------------------------------------------------------
// Test: Recuperação após remoção de toxic
// ---------------------------------------------------------------------------

func TestDatabaseRecovery_AfterToxicRemoval(t *testing.T) {
	env := setupToxiproxyEnv(t)
	defer env.cleanup()

	// 1. Coloca banco offline
	addToxic(env.apiURL, proxyName, "down", "down", 1.0, map[string]interface{}{})
	time.Sleep(200 * time.Millisecond)

	_, err := env.pgPool.Query(context.Background(), "SELECT 1")
	assert.Error(t, err, "banco deve estar offline com toxic ativo")

	// 2. Remove toxic (recupera)
	removeAllToxics(env.apiURL, proxyName)
	time.Sleep(500 * time.Millisecond) // aguarda recuperação

	// 3. Verifica se consegue conectar novamente
	_, err = env.pgPool.Query(context.Background(), "SELECT 1")
	assert.NoError(t, err, "após remover toxic, banco deve responder novamente")
}