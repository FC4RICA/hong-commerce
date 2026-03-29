package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                string
	UserServiceURL      string
	CatalogServiceURL   string
	OrderServiceURL     string
	InventoryServiceURL string
	PaymentServiceURL   string
	JWTSecret           string

	// Server timeouts
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration

	// Proxy transport
	ProxyMaxIdleConns        int
	ProxyMaxIdleConnsPerHost int
	ProxyIdleConnTimeout     time.Duration
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		UserServiceURL:      getEnv("USER_SERVICE_URL", "http://user-service:8081"),
		CatalogServiceURL:   getEnv("CATALOG_SERVICE_URL", "http://catalog-service:8082"),
		OrderServiceURL:     getEnv("ORDER_SERVICE_URL", "http://order-service:8083"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://inventory-service:8084"),
		PaymentServiceURL:   getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8085"),
		JWTSecret:           getEnv("JWT_SECRET", ""),

		ServerReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		ServerWriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		ServerIdleTimeout:  getDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),

		ProxyMaxIdleConns:        getInt("PROXY_MAX_IDLE_CONNS", 100),
		ProxyMaxIdleConnsPerHost: getInt("PROXY_MAX_IDLE_CONNS_PER_HOST", 10),
		ProxyIdleConnTimeout:     getDuration("PROXY_IDLE_CONN_TIMEOUT", 90*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	// Accept seconds as plain integer (e.g. "30") or Go duration string (e.g. "30s")
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	return fallback
}

func getInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	return fallback
}
