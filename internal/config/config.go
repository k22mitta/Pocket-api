package config

import "os"

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	Port              string
	PlaidClientID     string
	PlaidSecret       string
	PlaidEnv          string
	GeminiAPIKey      string
	CORSAllowedOrigin string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	plaidEnv := os.Getenv("PLAID_ENV")
	if plaidEnv == "" {
		plaidEnv = "sandbox"
	}

	corsAllowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsAllowedOrigin == "" {
		corsAllowedOrigin = "*"
	}

	return Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		Port:              port,
		PlaidClientID:     os.Getenv("PLAID_CLIENT_ID"),
		PlaidSecret:       os.Getenv("PLAID_SECRET"),
		PlaidEnv:          plaidEnv,
		GeminiAPIKey:      os.Getenv("GEMINI_API_KEY"),
		CORSAllowedOrigin: corsAllowedOrigin,
	}
}
