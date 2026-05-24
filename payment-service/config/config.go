package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	TimeZone   string
	Port       string
}

func GetEnv(key string, fallback ...string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	msg := fmt.Sprintf("Environment variable %s is required but not set", key)
	panic(msg)
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found. Using system environment variables instead.")
	}
	return &Config{
		DBHost:     GetEnv("DB_HOST"),
		DBPort:     GetEnv("DB_PORT"),
		DBUser:     GetEnv("DB_USER"),
		DBPassword: GetEnv("DB_PASSWORD"),
		DBName:     GetEnv("DB_NAME"),
		Port:       GetEnv("PORT", "8080"),
		TimeZone:   GetEnv("TIMEZONE", "Asia/Bangkok"),
	}
}
