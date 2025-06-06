package config

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	GrpcServerAddr string
	KafkaBrokers   []string
	KafkaTopic     string
	MaxConcurrency string
	DelaySeconds   string
	LogLevel       string
	LogFile        string
}

var (
	configInstance *Config
	once           sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		_ = godotenv.Load()

		configInstance = &Config{
			GrpcServerAddr: getEnv("GRPC_SERVER_ADDRESS", "localhost:50051"),
			KafkaBrokers:   []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			KafkaTopic:     getEnv("KAFKA_TOPIC", "container_topic"),
			MaxConcurrency: getEnv("MAX_CONCURRENCY", "20"),
			DelaySeconds:   getEnv("DELAY_SECONDS", "60"),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			LogFile:        getEnv("LOG_FILE", "../logs/healthcheck.log"),
		}
	})

	return configInstance
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
