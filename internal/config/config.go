package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 24 * time.Hour
)

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:         getEnv("SERVER_ADDR", ":9900"),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			Database: getEnv("POSTGRES_DB", "chatx"),
			SSLMode:  getEnv("POSTGRES_SSL", "disable"),
		},
		AuthToken: AuthTokenConfig{
			Secret:          getEnv("AUTH_TOKEN_SECRET", "secret"),
			AccessTokenTTL:  getEnvDuration("AUTH_TOKEN_ACCESS_TOKEN_TTL", defaultAccessTokenTTL),
			RefreshTokenTTL: getEnvDuration("AUTH_TOKEN_REFRESH_TOKEN_TTL", defaultRefreshTokenTTL),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			Bucket:          getEnv("MINIO_BUCKET", "chatx"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnvBool("MINIO_USE_SSL", false),
		},
		Kafka: KafkaConfig{
			Brokers:      getEnv("KAFKA_BROKERS", "localhost:9092"),
			SaslUsername: getEnv("KAFKA_SASL_USERNAME", ""),
			SaslPassword: getEnv("KAFKA_SASL_PASSWORD", ""),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@chatx.code19m.uz"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}
}

type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	AuthToken AuthTokenConfig
	MinIO     MinIOConfig
	Kafka     KafkaConfig
	SMTP      SMTPConfig
	Redis     RedisConfig
}

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

func (pc PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pc.Host,
		pc.Port,
		pc.User,
		pc.Password,
		pc.Database,
		pc.SSLMode,
	)
}

type AuthTokenConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type MinIOConfig struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

type KafkaConfig struct {
	Brokers      string
	SaslUsername string
	SaslPassword string
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// func getEnvSlice(key string, defaultValue []string) []string {
// 	if value := os.Getenv(key); value != "" {
// 		parts := strings.Split(value, ",")
// 		result := make([]string, 0, len(parts))
// 		for _, part := range parts {
// 			if trimmed := strings.TrimSpace(part); trimmed != "" {
// 				result = append(result, trimmed)
// 			}
// 		}
// 		return result
// 	}
// 	return defaultValue
// }
