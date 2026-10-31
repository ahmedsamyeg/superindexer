package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server configuration
	Server ServerConfig

	// Database configuration
	Database DatabaseConfig

	// Authentication configuration
	Auth AuthConfig

	// NNTP configuration
	NNTP NNTPConfig

	// Spotnet configuration
	Spotnet SpotnetConfig

	// Observability configuration
	Observability ObservabilityConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	DSN              string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	ConnMaxIdleTime  time.Duration
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	JWTSecret          string
	JWTExpiry          time.Duration
	SessionSecret      string
	SessionExpiry      time.Duration
	PasswordResetToken string
}

// NNTPConfig holds NNTP server settings.
type NNTPConfig struct {
	DefaultMaxConnections int
	DefaultTimeout        time.Duration
	DefaultRetryAttempts  int
	DefaultRetryDelay     time.Duration
}

// SpotnetConfig holds Spotnet source settings.
type SpotnetConfig struct {
	DefaultTimeout        time.Duration
	DefaultRetryAttempts  int
	DefaultRetryDelay     time.Duration
}

// ObservabilityConfig holds metrics and tracing settings.
type ObservabilityConfig struct {
	PrometheusEnabled bool
	OTLPEnabled       bool
	OTLPEndpoint      string
	LogLevel          string
	LogFormat         string
}

// Load reads all configuration from environment variables.
func Load() (*Config, error) {
	c := &Config{}

	c.Server = ServerConfig{
		Host:          getEnvString("SERVER_HOST", "0.0.0.0"),
		Port:          getEnvInt("SERVER_PORT", 8080),
		ReadTimeout:   getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:  getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
		ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
	}

	c.Database = DatabaseConfig{
		DSN:             getEnvString("DATABASE_DSN", "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 2*time.Minute),
	}

	c.Auth = AuthConfig{
		JWTSecret:          getEnvString("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:          getEnvDuration("JWT_EXPIRY", 24*time.Hour),
		SessionSecret:      getEnvString("SESSION_SECRET", "change-me-in-production"),
		SessionExpiry:      getEnvDuration("SESSION_EXPIRY", 7*24*time.Hour),
		PasswordResetToken: getEnvString("PASSWORD_RESET_TOKEN", ""),
	}

	c.NNTP = NNTPConfig{
		DefaultMaxConnections: getEnvInt("NNTP_DEFAULT_MAX_CONNECTIONS", 10),
		DefaultTimeout:        getEnvDuration("NNTP_DEFAULT_TIMEOUT", 30*time.Second),
		DefaultRetryAttempts:  getEnvInt("NNTP_DEFAULT_RETRY_ATTEMPTS", 3),
		DefaultRetryDelay:     getEnvDuration("NNTP_DEFAULT_RETRY_DELAY", 2*time.Second),
	}

	c.Spotnet = SpotnetConfig{
		DefaultTimeout:        getEnvDuration("SPOTNET_DEFAULT_TIMEOUT", 30*time.Second),
		DefaultRetryAttempts:  getEnvInt("SPOTNET_DEFAULT_RETRY_ATTEMPTS", 3),
		DefaultRetryDelay:     getEnvDuration("SPOTNET_DEFAULT_RETRY_DELAY", 2*time.Second),
	}

	c.Observability = ObservabilityConfig{
		PrometheusEnabled: getEnvBool("PROMETHEUS_ENABLED", true),
		OTLPEnabled:       getEnvBool("OTEL_ENABLED", false),
		OTLPEndpoint:      getEnvString("OTEL_ENDPOINT", "http://localhost:4317"),
		LogLevel:          getEnvString("LOG_LEVEL", "info"),
		LogFormat:         getEnvString("LOG_FORMAT", "json"),
	}

	return c, nil
}

// GetDSN returns the database connection string.
func (c *DatabaseConfig) GetDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnvString("DB_USER", "indexer"),
		getEnvString("DB_PASSWORD", "indexer"),
		getEnvString("DB_HOST", "localhost"),
		getEnvString("DB_PORT", "5432"),
		getEnvString("DB_NAME", "indexer"),
		getEnvString("DB_SSL_MODE", "disable"),
	)
}

// getEnvString returns the environment variable value or the default.
func getEnvString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// getEnvInt returns the environment variable as an integer or the default.
func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// getEnvDuration returns the environment variable as a duration or the default.
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

// getEnvBool returns the environment variable as a boolean or the default.
func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
