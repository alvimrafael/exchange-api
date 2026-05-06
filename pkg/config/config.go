package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	RedisURL            string
	CacheTTL            int // segundos
	ExchangeAPIKey      string
	RateLimitRPS        float64
	RateLimitBurst      int
	WebhookIntervalSecs int
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		RedisURL:            getEnv("REDIS_URL", ""),
		CacheTTL:            getEnvInt("CACHE_TTL_SECONDS", 300),
		ExchangeAPIKey:      getEnv("EXCHANGE_API_KEY", ""),
		RateLimitRPS:        getEnvFloat("RATE_LIMIT_RPS", 5),
		RateLimitBurst:      getEnvInt("RATE_LIMIT_BURST", 10),
		WebhookIntervalSecs: getEnvInt("WEBHOOK_INTERVAL_SECONDS", 600),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
