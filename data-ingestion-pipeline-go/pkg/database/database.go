package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/anurag/data-ingestion-pipeline-go/internal/config"
	"github.com/anurag/data-ingestion-pipeline-go/pkg/logger"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DB interface defines database operations
// Go Concept: Interface for database abstraction and testing
type DB interface {
	// Raw database access
	GetDB() *sql.DB

	// Health check
	Ping(ctx context.Context) error

	// Transaction management
	BeginTx(ctx context.Context) (*sql.Tx, error)

	// Query operations
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Close connection
	Close() error
}

// database implements DB interface
// Go Concept: Struct implementing interface with embedded type
type database struct {
	*sql.DB // Embedded type - gets all methods of sql.DB
	logger  logger.Logger
	config  config.DatabaseConfig
}

// New creates a new database connection
// Go Concept: Constructor function with error handling
func New(cfg config.DatabaseConfig, log logger.Logger) (DB, error) {
	// Build connection string for SQLite
	// SQLite uses file path as connection string
	connStr := cfg.Database
	if connStr == "" {
		connStr = "./data.db" // Default SQLite file
	}

	// Open database connection
	// Go Concept: Using blank import for driver registration
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	// Go Concept: Connection pool configuration
	db.SetMaxOpenConns(cfg.MaxOpenConns)   // Maximum open connections
	db.SetMaxIdleConns(cfg.MaxIdleConns)   // Maximum idle connections
	db.SetConnMaxLifetime(cfg.ConnMaxLife) // Maximum connection lifetime

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("SQLite database connection established",
		logger.String("database_file", connStr),
		logger.Int("max_open_conns", cfg.MaxOpenConns),
		logger.Int("max_idle_conns", cfg.MaxIdleConns),
	)

	return &database{
		DB:     db,
		logger: log,
		config: cfg,
	}, nil
}

// GetDB returns the underlying sql.DB
// Go Concept: Getter method for accessing embedded type
func (d *database) GetDB() *sql.DB {
	return d.DB
}

// Ping checks database connectivity
// Go Concept: Method with context for timeout control
func (d *database) Ping(ctx context.Context) error {
	if err := d.DB.PingContext(ctx); err != nil {
		d.logger.Error("Database ping failed", logger.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// BeginTx starts a database transaction
// Go Concept: Transaction management with context
func (d *database) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		d.logger.Error("Failed to begin transaction", logger.Error(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	d.logger.Debug("Database transaction started")
	return tx, nil
}

// QueryContext executes a query with context
// Go Concept: Wrapping methods with logging and error handling
func (d *database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Query failed",
			logger.String("query", query),
			logger.Any("args", args),
			logger.Duration("duration", duration),
			logger.Error(err),
		)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	d.logger.Debug("Query executed",
		logger.String("query", query),
		logger.Duration("duration", duration),
	)

	return rows, nil
}

// QueryRowContext executes a query that returns a single row
func (d *database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := d.DB.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	d.logger.Debug("Single row query executed",
		logger.String("query", query),
		logger.Duration("duration", duration),
	)

	return row
}

// ExecContext executes a command (INSERT, UPDATE, DELETE)
func (d *database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := d.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Command execution failed",
			logger.String("query", query),
			logger.Any("args", args),
			logger.Duration("duration", duration),
			logger.Error(err),
		)
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	// Log affected rows if available
	if result != nil {
		if rowsAffected, err := result.RowsAffected(); err == nil {
			d.logger.Debug("Command executed",
				logger.String("query", query),
				logger.Duration("duration", duration),
				logger.Int64("rows_affected", rowsAffected),
			)
		}
	}

	return result, nil
}

// Close closes the database connection
func (d *database) Close() error {
	d.logger.Info("Closing database connection")
	return d.DB.Close()
}

// Health check utilities

// HealthCheck performs a comprehensive database health check
// Go Concept: Function that returns structured health information
func (d *database) HealthCheck(ctx context.Context) (*HealthInfo, error) {
	start := time.Now()

	// Ping test
	if err := d.Ping(ctx); err != nil {
		return &HealthInfo{
			Status:    "unhealthy",
			Error:     err.Error(),
			CheckedAt: start,
			Duration:  time.Since(start),
		}, err
	}

	// Get connection stats
	stats := d.DB.Stats()

	// Simple query test
	var result int
	err := d.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return &HealthInfo{
			Status:    "unhealthy",
			Error:     fmt.Sprintf("query test failed: %v", err),
			CheckedAt: start,
			Duration:  time.Since(start),
		}, err
	}

	return &HealthInfo{
		Status:          "healthy",
		CheckedAt:       start,
		Duration:        time.Since(start),
		OpenConnections: stats.OpenConnections,
		IdleConnections: stats.Idle,
		MaxOpenConns:    stats.MaxOpenConnections,
	}, nil
}

// HealthInfo contains database health information
// Go Concept: Struct for health check results
type HealthInfo struct {
	Status          string        `json:"status"`
	Error           string        `json:"error,omitempty"`
	CheckedAt       time.Time     `json:"checked_at"`
	Duration        time.Duration `json:"duration"`
	OpenConnections int           `json:"open_connections"`
	IdleConnections int           `json:"idle_connections"`
	MaxOpenConns    int           `json:"max_open_connections"`
}

// Migration utilities

// Migrate runs database migrations
// Go Concept: Simple migration system
func (d *database) Migrate(ctx context.Context) error {
	d.logger.Info("Starting database migration")

	// Create migrations table if it doesn't exist
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	if _, err := d.ExecContext(ctx, createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define migrations
	migrations := []Migration{
		{
			Version: "001_create_posts_table",
			SQL: `
				CREATE TABLE IF NOT EXISTS posts (
					id INTEGER PRIMARY KEY,
					user_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					body TEXT NOT NULL,
					ingested_at DATETIME NOT NULL,
					source TEXT NOT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);
				
				CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
				CREATE INDEX IF NOT EXISTS idx_posts_ingested_at ON posts(ingested_at);
				CREATE INDEX IF NOT EXISTS idx_posts_source ON posts(source);
			`,
		},
	}

	// Apply migrations
	for _, migration := range migrations {
		if err := d.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
	}

	d.logger.Info("Database migration completed")
	return nil
}

// Migration represents a database migration
type Migration struct {
	Version string
	SQL     string
}

// applyMigration applies a single migration
func (d *database) applyMigration(ctx context.Context, migration Migration) error {
	// Check if migration already applied
	var count int
	err := d.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM schema_migrations WHERE version = $1",
		migration.Version,
	).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		d.logger.Debug("Migration already applied", logger.String("version", migration.Version))
		return nil
	}

	// Begin transaction for migration
	tx, err := d.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback if not committed

	// Execute migration SQL
	_, err = tx.ExecContext(ctx, migration.SQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	_, err = tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		migration.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	d.logger.Info("Migration applied", logger.String("version", migration.Version))
	return nil
}
