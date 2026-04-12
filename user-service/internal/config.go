package internal

import (
	"log"
	"os"
)

type config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	SeedAdminEmail    string
	SeedAdminPassword string
}

func LoadConfig() config {
	return config{
		Port:              getEnv("PORT", "8081"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable"),
		JWTSecret:         mustGetEnv("JWT_SECRET"),
		SeedAdminEmail:    getEnv("SEED_ADMIN_EMAIL", "admin@mail.com"),
		SeedAdminPassword: getEnv("SEED_ADMIN_PASSWORD", "admin1234"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
