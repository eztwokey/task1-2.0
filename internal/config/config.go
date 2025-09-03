package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName        string
	HTTPPort       string
	LogLevel       string

	PostgresURL    string
	PGMaxConns     int

	KafkaBrokers   []string
	KafkaTopic     string
	KafkaGroupID   string

	CacheTTL       time.Duration
	CacheJanitor   time.Duration
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}

func atoi(def int, s string) int {
	if s == "" { return def }
	if n, err := strconv.Atoi(s); err == nil { return n }
	return def
}

func dur(def time.Duration, s string) time.Duration {
	if s == "" { return def }
	d, err := time.ParseDuration(s)
	if err != nil { return def }
	return d
}

func Load() *Config {
	return &Config{
		AppName:      getenv("APP_NAME", "wb-order-service"),
		HTTPPort:     getenv("HTTP_PORT", "8081"),
		LogLevel:     getenv("LOG_LEVEL", "INFO"),

		PostgresURL:  getenv("POSTGRES_URL", "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"),
		PGMaxConns:   atoi(10, os.Getenv("PG_MAX_CONNS")),

		KafkaBrokers: strings.Split(getenv("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaTopic:   getenv("KAFKA_ORDERS_TOPIC", "orders"),
		KafkaGroupID: getenv("KAFKA_GROUP_ID", "wb-order-consumer"),

		CacheTTL:     dur(5*time.Minute, os.Getenv("CACHE_TTL")),
		CacheJanitor: dur(1*time.Minute, os.Getenv("CACHE_JANITOR_INTERVAL")),
	}
}
