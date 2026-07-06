package main

import "os"

// Config contém as configurações da aplicação.
type Config struct {
	DatabaseURL string
	ServerPort  string
	APIKeys     string
}

// LoadConfig carrega as configurações das variáveis de ambiente.
func LoadConfig() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		APIKeys:     getEnv("API_KEYS", "demo-key:ClienteDemo"),
	}
}

// getEnv retorna o valor da variável de ambiente ou um valor padrão.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}