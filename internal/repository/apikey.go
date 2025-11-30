package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key record from the database
type APIKey struct {
	ID                uuid.UUID
	KeyHash           string
	Name              string
	RateLimit         int
	RateWindowSeconds int
	IsActive          bool
	CreatedAt         time.Time
	LastUsedAt        sql.NullTime
}

// APIKeyRepository defines the interface for API key operations
type APIKeyRepository interface {
	GetByKeyHash(keyHash string) (*APIKey, error)
	Create(name string, keyHash string, rateLimit int, rateWindowSeconds int) (*APIKey, error)
	List() ([]APIKey, error)
	Delete(id uuid.UUID) error
	UpdateLastUsed(id uuid.UUID) error
	IncrementRequestCount(apiKeyID uuid.UUID, windowStart time.Time) (int, error)
	GetRequestCount(apiKeyID uuid.UUID, windowStart time.Time) (int, error)
}

// PostgresAPIKeyRepository implements APIKeyRepository using PostgreSQL
type PostgresAPIKeyRepository struct {
	database *sql.DB
}

// NewPostgresAPIKeyRepository creates a new PostgreSQL-backed API key repository
func NewPostgresAPIKeyRepository(database *sql.DB) *PostgresAPIKeyRepository {
	return &PostgresAPIKeyRepository{
		database: database,
	}
}

// HashAPIKey hashes an API key using SHA-256
func HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// GetByKeyHash retrieves an API key by its hash
func (repository *PostgresAPIKeyRepository) GetByKeyHash(keyHash string) (*APIKey, error) {
	query := `
		SELECT id, key_hash, name, rate_limit, rate_window_seconds, is_active, created_at, last_used_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true
	`

	apiKey := &APIKey{}
	err := repository.database.QueryRow(query, keyHash).Scan(
		&apiKey.ID,
		&apiKey.KeyHash,
		&apiKey.Name,
		&apiKey.RateLimit,
		&apiKey.RateWindowSeconds,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.LastUsedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

// Create creates a new API key record
func (repository *PostgresAPIKeyRepository) Create(name string, keyHash string, rateLimit int, rateWindowSeconds int) (*APIKey, error) {
	query := `
		INSERT INTO api_keys (key_hash, name, rate_limit, rate_window_seconds)
		VALUES ($1, $2, $3, $4)
		RETURNING id, key_hash, name, rate_limit, rate_window_seconds, is_active, created_at, last_used_at
	`

	apiKey := &APIKey{}
	err := repository.database.QueryRow(query, keyHash, name, rateLimit, rateWindowSeconds).Scan(
		&apiKey.ID,
		&apiKey.KeyHash,
		&apiKey.Name,
		&apiKey.RateLimit,
		&apiKey.RateWindowSeconds,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.LastUsedAt,
	)

	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

// List retrieves all API keys
func (repository *PostgresAPIKeyRepository) List() ([]APIKey, error) {
	query := `
		SELECT id, key_hash, name, rate_limit, rate_window_seconds, is_active, created_at, last_used_at
		FROM api_keys
		ORDER BY created_at DESC
	`

	rows, err := repository.database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []APIKey
	for rows.Next() {
		var apiKey APIKey
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.KeyHash,
			&apiKey.Name,
			&apiKey.RateLimit,
			&apiKey.RateWindowSeconds,
			&apiKey.IsActive,
			&apiKey.CreatedAt,
			&apiKey.LastUsedAt,
		)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, rows.Err()
}

// Delete soft-deletes an API key by setting is_active to false
func (repository *PostgresAPIKeyRepository) Delete(id uuid.UUID) error {
	query := `UPDATE api_keys SET is_active = false WHERE id = $1`
	_, err := repository.database.Exec(query, id)
	return err
}

// UpdateLastUsed updates the last_used_at timestamp for an API key
func (repository *PostgresAPIKeyRepository) UpdateLastUsed(id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := repository.database.Exec(query, id)
	return err
}

// IncrementRequestCount increments the request count for a given time window
// Uses upsert to handle both new and existing windows atomically
func (repository *PostgresAPIKeyRepository) IncrementRequestCount(apiKeyID uuid.UUID, windowStart time.Time) (int, error) {
	query := `
		INSERT INTO rate_limit_records (api_key_id, window_start, request_count)
		VALUES ($1, $2, 1)
		ON CONFLICT (api_key_id, window_start)
		DO UPDATE SET request_count = rate_limit_records.request_count + 1
		RETURNING request_count
	`

	var requestCount int
	err := repository.database.QueryRow(query, apiKeyID, windowStart).Scan(&requestCount)
	if err != nil {
		return 0, err
	}

	return requestCount, nil
}

// GetRequestCount retrieves the current request count for a given time window
func (repository *PostgresAPIKeyRepository) GetRequestCount(apiKeyID uuid.UUID, windowStart time.Time) (int, error) {
	query := `
		SELECT request_count
		FROM rate_limit_records
		WHERE api_key_id = $1 AND window_start = $2
	`

	var requestCount int
	err := repository.database.QueryRow(query, apiKeyID, windowStart).Scan(&requestCount)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return requestCount, nil
}
