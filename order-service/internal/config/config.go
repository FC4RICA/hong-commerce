package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DBPort         string
	DBHost         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	RabbitMQURL    string
	TimeoutMinutes int
}

func LoadConfig() *Config {
	timeoutStr := getEnv("ORDER_TIMEOUT_MINUTES", "15")
	timeout, _ := strconv.Atoi(timeoutStr)

	return &Config{
		Port:           getEnv("PORT", "8083"),
		DBPort:         getEnv("ORDER_DB_PORT", "5434"),
		DBHost:         getEnv("ORDER_DB_HOST", "localhost"),
		DBUser:         getEnv("ORDER_DB_USER", "postgres"),
		DBPassword:     getEnv("ORDER_DB_PASSWORD", "postgres"),
		DBName:         getEnv("ORDER_DB_NAME", "orderdb"),
		DBSSLMode:      getEnv("ORDER_DB_SSLMODE", "disable"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		TimeoutMinutes: timeout,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
