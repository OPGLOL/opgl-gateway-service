package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User represents a user record from the database
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(email string, passwordHash string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByID(id uuid.UUID) (*User, error)
}

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	database *sql.DB
}

// NewPostgresUserRepository creates a new PostgreSQL-backed user repository
func NewPostgresUserRepository(database *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		database: database,
	}
}

// Create creates a new user record
func (repository *PostgresUserRepository) Create(email string, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at, updated_at
	`

	user := &User{}
	err := repository.database.QueryRow(query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email address
func (repository *PostgresUserRepository) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}
	err := repository.database.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (repository *PostgresUserRepository) GetByID(id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := repository.database.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
