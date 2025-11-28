package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config groups every environment-based flag required by the backend.
type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	JWTSecret      string
	APIPort        string
	FrontendOrigin string
}

// Load reads .env (either in backend/ or repo root) and exposes the configuration.
func Load() (*Config, error) {
	loadEnvFile(".env")
	loadEnvFile(filepath.Clean("../.env"))

	cfg := &Config{
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		DBSSLMode:      fallback(os.Getenv("DB_SSLMODE"), "require"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		APIPort:        fallback(os.Getenv("API_PORT"), "8080"),
		FrontendOrigin: os.Getenv("FRONTEND_ORIGIN"),
	}

	return cfg, nil
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func loadEnvFile(path string) {
	_ = godotenv.Load(path)
}
