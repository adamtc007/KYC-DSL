package dataservice

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the global connection pool for the Data Service
var DB *pgxpool.Pool

// InitDB initializes the PostgreSQL connection pool
// Environment variables:
//   - DATABASE_URL: full connection string (default: localhost:5432/kyc_dsl)
//   - PGHOST: PostgreSQL host (default: localhost)
//   - PGPORT: PostgreSQL port (default: 5432)
//   - PGUSER: PostgreSQL user (default: postgres)
//   - PGPASSWORD: PostgreSQL password (default: postgres)
//   - PGDATABASE: PostgreSQL database (default: kyc_dsl)
func InitDB() error {
	dsn := getDatabaseURL()

	// Parse the connection string
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	cfg.MaxConns = 20                      // Maximum number of connections
	cfg.MinConns = 5                       // Minimum number of connections
	cfg.MaxConnLifetime = time.Hour        // Maximum connection lifetime
	cfg.MaxConnIdleTime = 30 * time.Minute // Maximum idle time
	cfg.HealthCheckPeriod = time.Minute    // Health check interval

	// Create the pool with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	DB, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	if err := DB.Ping(ctx); err != nil {
		DB.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	log.Printf("📊 Connection pool: max=%d, min=%d", cfg.MaxConns, cfg.MinConns)

	return nil
}

// CloseDB closes the database connection pool gracefully
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("🔒 Database connection pool closed")
	}
}

// getDatabaseURL constructs the database connection string from environment variables
func getDatabaseURL() string {
	// Check for full DATABASE_URL first
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	// Build from individual components
	host := getEnv("PGHOST", "localhost")
	port := getEnv("PGPORT", "5432")
	user := getEnv("PGUSER", "postgres")
	password := getEnv("PGPASSWORD", "postgres")
	database := getEnv("PGDATABASE", "kyc_dsl")
	sslmode := getEnv("PGSSLMODE", "disable")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, database, sslmode,
	)
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// HealthCheck verifies the database connection is alive
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database pool not initialized")
	}
	return DB.Ping(ctx)
}
