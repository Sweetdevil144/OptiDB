package db

import (
	"database/sql"
	"fmt"
	"os"

	"cli/internal/logger"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

func NewConfig() *Config {
	return &Config{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		Database: getEnv("POSTGRES_DB", "optidb"),
		Username: getEnv("POSTGRES_USER", "postgres"),
		Password: getEnv("POSTGRES_PASSWORD", "postgres"),
	}
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Database)
}

func Connect(config *Config) (*sql.DB, error) {
	logger.LogInfof("Attempting to connect to database: %s:%s/%s as user %s",
		config.Host, config.Port, config.Database, config.Username)

	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		logger.LogErrorf("Failed to open database connection: %v", err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		logger.LogErrorf("Failed to ping database: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.LogInfo("Database connection established successfully")
	return db, nil
}

func ConnectAsProfiler() (*sql.DB, error) {
	logger.LogInfo("Connecting as profiler_ro user")
	config := &Config{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		Database: getEnv("POSTGRES_DB", "optidb"),
		Username: "profiler_ro",
		Password: "profiler_ro_pass",
	}
	return Connect(config)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
