package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	apierrors "github.com/OPGLOL/opgl-gateway-service/internal/errors"
	"github.com/OPGLOL/opgl-gateway-service/internal/ratelimit"
)

// RateLimitMiddleware creates middleware that enforces rate limiting based on API keys
func RateLimitMiddleware(rateLimiter *ratelimit.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Extract API key from header
			apiKey := request.Header.Get("X-API-Key")

			// If no API key provided, reject the request
			if apiKey == "" {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeMissingAPIKey,
					"API key is required. Include X-API-Key header in your request.",
					http.StatusUnauthorized,
				))
				return
			}

			// Check rate limit
			rateLimitResult, err := rateLimiter.CheckRateLimit(apiKey)
			if err != nil {
				apierrors.WriteError(responseWriter, apierrors.InternalError("Rate limit check failed"))
				return
			}

			// Add rate limit headers to response
			responseWriter.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimitResult.Limit))
			responseWriter.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rateLimitResult.Remaining))
			responseWriter.Header().Set("X-RateLimit-Reset", strconv.FormatInt(rateLimitResult.ResetTime.Unix(), 10))

			// If API key is invalid (Limit is 0), reject
			if rateLimitResult.Limit == 0 {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeInvalidAPIKey,
					"Invalid or inactive API key.",
					http.StatusUnauthorized,
				))
				return
			}

			// If rate limit exceeded, reject with 429
			if !rateLimitResult.Allowed {
				retryAfter := rateLimitResult.ResetTime.Unix() - time.Now().Unix()
				if retryAfter < 0 {
					retryAfter = 1
				}
				responseWriter.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))

				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeRateLimitExceeded,
					fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retryAfter),
					http.StatusTooManyRequests,
				))
				return
			}

			// Request allowed, proceed to next handler
			next.ServeHTTP(responseWriter, request)
		})
	}
}

// OptionalRateLimitMiddleware creates middleware that enforces rate limiting only if API key is provided
// This is useful for endpoints that should work without API key but have rate limits when one is provided
func OptionalRateLimitMiddleware(rateLimiter *ratelimit.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Extract API key from header
			apiKey := request.Header.Get("X-API-Key")

			// If no API key provided, allow request without rate limiting
			if apiKey == "" {
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Check rate limit
			rateLimitResult, err := rateLimiter.CheckRateLimit(apiKey)
			if err != nil {
				apierrors.WriteError(responseWriter, apierrors.InternalError("Rate limit check failed"))
				return
			}

			// Add rate limit headers to response
			responseWriter.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimitResult.Limit))
			responseWriter.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rateLimitResult.Remaining))
			responseWriter.Header().Set("X-RateLimit-Reset", strconv.FormatInt(rateLimitResult.ResetTime.Unix(), 10))

			// If API key is invalid, reject
			if rateLimitResult.Limit == 0 {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeInvalidAPIKey,
					"Invalid or inactive API key.",
					http.StatusUnauthorized,
				))
				return
			}

			// If rate limit exceeded, reject with 429
			if !rateLimitResult.Allowed {
				responseWriter.Header().Set("Retry-After", strconv.FormatInt(rateLimitResult.ResetTime.Unix(), 10))
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeRateLimitExceeded,
					"Rate limit exceeded.",
					http.StatusTooManyRequests,
				))
				return
			}

			next.ServeHTTP(responseWriter, request)
		})
	}
}
