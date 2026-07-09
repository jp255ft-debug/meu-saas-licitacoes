package main

import "os"

// Config contém as configurações da aplicação.
type Config struct {
	DatabaseURL       string
	ServerPort        string
	APIKeys           string
	AsaasWebhookToken string
	AsaasWebhookSecret string
}

// LoadConfig carrega as configurações das variáveis de ambiente.
func LoadConfig() Config {
	return Config{
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		APIKeys:            getEnv("API_KEYS", "demo-key:ClienteDemo"),
		AsaasWebhookToken:  getEnv("ASAAS_WEBHOOK_TOKEN", ""),
		AsaasWebhookSecret: getEnv("ASAAS_WEBHOOK_SECRET", ""),
	}
}

// getEnv retorna o valor da variável de ambiente ou um valor padrão.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}