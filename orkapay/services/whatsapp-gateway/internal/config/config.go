package config

import "os"

type Config struct {
	Port         string
	BSPProvider  string
	BSPAPIKey    string
	BSPBaseURL   string
	BSPPhoneID   string
	CashinURL    string
	CashinAPIKey string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	return &Config{
		Port:         port,
		BSPProvider:  os.Getenv("BSP_PROVIDER"),
		BSPAPIKey:    os.Getenv("BSP_API_KEY"),
		BSPBaseURL:   os.Getenv("BSP_BASE_URL"),
		BSPPhoneID:   os.Getenv("BSP_PHONE_ID"),
		CashinURL:    os.Getenv("CASHIN_BASE_URL"),
		CashinAPIKey: os.Getenv("CASHIN_API_KEY"),
	}
}