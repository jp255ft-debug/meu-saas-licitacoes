package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Rate     RateConfig
	Log      LogConfig
	TLS      TLSConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	URL      string
	MaxConns int
	MinConns int
}

// AuthConfig holds API key authentication configuration.
type AuthConfig struct {
	APIKeys map[string]string // key -> clientID
}

// RateConfig holds rate limiting configuration.
type RateConfig struct {
	MaxReqs int
	Window  time.Duration
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string
	JSON  bool
}

// TLSConfig holds TLS configuration.
type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getEnvInt returns the integer value of an environment variable or a default.
func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvDuration returns the duration value of an environment variable.
func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

// parseAPIKeys parses API_KEYS env var in format "key1:client1,key2:client2".
func parseAPIKeys(envVal string) map[string]string {
	result := make(map[string]string)
	if envVal == "" {
		return result
	}
	pairs := strings.Split(envVal, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", "postgres://localhost:5432/cashin?sslmode=disable"),
			MaxConns: getEnvInt("DATABASE_MAX_CONNS", 20),
			MinConns: getEnvInt("DATABASE_MIN_CONNS", 5),
		},
		Auth: AuthConfig{
			APIKeys: parseAPIKeys(getEnv("API_KEYS", "demo-key:ClienteDemo")),
		},
		Rate: RateConfig{
			MaxReqs: getEnvInt("RATE_LIMIT_MAX", 100),
			Window:  getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			JSON:  strings.ToLower(getEnv("LOG_JSON", "true")) == "true",
		},
		TLS: TLSConfig{
			Enabled:  strings.ToLower(getEnv("TLS_ENABLED", "false")) == "true",
			CertFile: getEnv("TLS_CERT_FILE", "/certs/cert.pem"),
			KeyFile:  getEnv("TLS_KEY_FILE", "/certs/key.pem"),
		},
	}
}