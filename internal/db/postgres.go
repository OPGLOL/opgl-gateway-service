package db

import (
	"database/sql"
	"fmt"

	// PostgreSQL driver import for database/sql
	_ "github.com/lib/pq"
)

// Database wraps the sql.DB connection for PostgreSQL operations
type Database struct {
	*sql.DB
}

// NewPostgresConnection creates a new PostgreSQL database connection
func NewPostgresConnection(host string, port string, user string, password string, dbname string) (*Database, error) {
	// Build PostgreSQL connection string
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	// Open database connection
	sqlDB, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Verify connection is working
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{DB: sqlDB}, nil
}

// Close closes the database connection
func (database *Database) Close() error {
	if database.DB != nil {
		return database.DB.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (database *Database) Ping() error {
	return database.DB.Ping()
}
