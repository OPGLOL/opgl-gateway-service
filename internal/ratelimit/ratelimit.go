package ratelimit

import (
	"time"

	"github.com/OPGLOL/opgl-gateway-service/internal/repository"
)

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed   bool
	Limit     int
	Remaining int
	ResetTime time.Time
}

// RateLimiter handles rate limiting logic using fixed time windows
type RateLimiter struct {
	apiKeyRepository repository.APIKeyRepository
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(apiKeyRepository repository.APIKeyRepository) *RateLimiter {
	return &RateLimiter{
		apiKeyRepository: apiKeyRepository,
	}
}

// CheckRateLimit verifies if a request is allowed based on the API key's rate limit
// Returns rate limit information including whether the request is allowed
func (rateLimiter *RateLimiter) CheckRateLimit(apiKey string) (*RateLimitResult, error) {
	// Hash the API key to look up in database
	keyHash := repository.HashAPIKey(apiKey)

	// Get the API key from database
	apiKeyRecord, err := rateLimiter.apiKeyRepository.GetByKeyHash(keyHash)
	if err != nil {
		return nil, err
	}

	// API key not found or inactive
	if apiKeyRecord == nil {
		return &RateLimitResult{
			Allowed:   false,
			Limit:     0,
			Remaining: 0,
			ResetTime: time.Now(),
		}, nil
	}

	// Calculate the current time window start
	windowDuration := time.Duration(apiKeyRecord.RateWindowSeconds) * time.Second
	windowStart := rateLimiter.calculateWindowStart(time.Now(), windowDuration)
	resetTime := windowStart.Add(windowDuration)

	// Increment request count and get new count atomically
	requestCount, err := rateLimiter.apiKeyRepository.IncrementRequestCount(apiKeyRecord.ID, windowStart)
	if err != nil {
		return nil, err
	}

	// Update last used timestamp (fire and forget - don't block on this)
	go func() {
		rateLimiter.apiKeyRepository.UpdateLastUsed(apiKeyRecord.ID)
	}()

	// Check if rate limit exceeded
	allowed := requestCount <= apiKeyRecord.RateLimit
	remaining := apiKeyRecord.RateLimit - requestCount
	if remaining < 0 {
		remaining = 0
	}

	return &RateLimitResult{
		Allowed:   allowed,
		Limit:     apiKeyRecord.RateLimit,
		Remaining: remaining,
		ResetTime: resetTime,
	}, nil
}

// calculateWindowStart calculates the start of the current time window
// Uses fixed windows aligned to the window duration
func (rateLimiter *RateLimiter) calculateWindowStart(currentTime time.Time, windowDuration time.Duration) time.Time {
	windowSeconds := int64(windowDuration.Seconds())
	currentUnix := currentTime.Unix()
	windowStartUnix := (currentUnix / windowSeconds) * windowSeconds
	return time.Unix(windowStartUnix, 0).UTC()
}

// ValidateAPIKey checks if an API key is valid without incrementing the counter
func (rateLimiter *RateLimiter) ValidateAPIKey(apiKey string) (bool, error) {
	keyHash := repository.HashAPIKey(apiKey)
	apiKeyRecord, err := rateLimiter.apiKeyRepository.GetByKeyHash(keyHash)
	if err != nil {
		return false, err
	}
	return apiKeyRecord != nil, nil
}
