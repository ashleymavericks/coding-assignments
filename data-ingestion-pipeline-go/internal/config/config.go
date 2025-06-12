package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
// Go Concept: Struct composition - embedding smaller structs into larger ones
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// External API configuration
	API APIConfig `json:"api"`

	// Ingestion worker configuration
	Ingestion IngestionConfig `json:"ingestion"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	User         string        `json:"user"`
	Password     string        `json:"password"`
	Database     string        `json:"database"`
	SSLMode      string        `json:"ssl_mode"`
	MaxOpenConns int           `json:"max_open_conns"`
	MaxIdleConns int           `json:"max_idle_conns"`
	ConnMaxLife  time.Duration `json:"conn_max_life"`
}

// APIConfig holds external API configuration
type APIConfig struct {
	BaseURL    string        `json:"base_url"`
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	RetryDelay time.Duration `json:"retry_delay"`
	RateLimit  int           `json:"rate_limit"` // requests per second
}

// IngestionConfig holds data ingestion worker configuration
type IngestionConfig struct {
	Interval     time.Duration `json:"interval"`      // How often to run ingestion
	BatchSize    int           `json:"batch_size"`    // How many posts to process at once
	WorkerCount  int           `json:"worker_count"`  // Number of concurrent workers
	QueueSize    int           `json:"queue_size"`    // Size of job queue
	EnableWorker bool          `json:"enable_worker"` // Whether to run background worker
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // json, text
}

// Load loads configuration from environment variables
// Go Concept: Function that returns struct and error
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:         getEnvString("SERVER_HOST", "localhost"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:         getEnvString("DB_HOST", ""),                             // Not used for SQLite
			Port:         getEnvInt("DB_PORT", 0),                                 // Not used for SQLite
			User:         getEnvString("DB_USER", ""),                             // Not used for SQLite
			Password:     getEnvString("DB_PASSWORD", ""),                         // Not used for SQLite
			Database:     getEnvString("DB_FILE", "./data/ingestion_pipeline.db"), // SQLite file path
			SSLMode:      getEnvString("DB_SSL_MODE", ""),                         // Not used for SQLite
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 1),                       // SQLite recommends 1 for writes
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 1),
			ConnMaxLife:  getEnvDuration("DB_CONN_MAX_LIFE", 5*time.Minute),
		},
		API: APIConfig{
			BaseURL:    getEnvString("API_BASE_URL", "https://jsonplaceholder.typicode.com"),
			Timeout:    getEnvDuration("API_TIMEOUT", 30*time.Second),
			RetryCount: getEnvInt("API_RETRY_COUNT", 3),
			RetryDelay: getEnvDuration("API_RETRY_DELAY", 1*time.Second),
			RateLimit:  getEnvInt("API_RATE_LIMIT", 10),
		},
		Ingestion: IngestionConfig{
			Interval:     getEnvDuration("INGESTION_INTERVAL", 5*time.Minute),
			BatchSize:    getEnvInt("INGESTION_BATCH_SIZE", 10),
			WorkerCount:  getEnvInt("INGESTION_WORKER_COUNT", 3),
			QueueSize:    getEnvInt("INGESTION_QUEUE_SIZE", 100),
			EnableWorker: getEnvBool("INGESTION_ENABLE_WORKER", true),
		},
		Logging: LoggingConfig{
			Level:  getEnvString("LOG_LEVEL", "info"),
			Format: getEnvString("LOG_FORMAT", "json"),
		},
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
// Go Concept: Method for validating struct data
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Validate database config (SQLite only needs file path)
	if c.Database.Database == "" {
		return fmt.Errorf("database file path is required")
	}

	// Validate API config
	if c.API.BaseURL == "" {
		return fmt.Errorf("API base URL is required")
	}

	// Validate ingestion config
	if c.Ingestion.WorkerCount < 1 {
		return fmt.Errorf("worker count must be at least 1")
	}
	if c.Ingestion.BatchSize < 1 {
		return fmt.Errorf("batch size must be at least 1")
	}

	return nil
}

// GetDatabaseURL returns the database connection URL
// Go Concept: Method that builds data from struct fields
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// Helper functions for parsing environment variables
// Go Concept: Package-level functions for common operations

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
