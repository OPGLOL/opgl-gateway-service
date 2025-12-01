package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/OPGLOL/opgl-gateway-service/internal/auth"
	apierrors "github.com/OPGLOL/opgl-gateway-service/internal/errors"
)

// AuthMiddleware creates middleware that validates JWT access tokens
func AuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Extract Authorization header
			authHeader := request.Header.Get("Authorization")

			if authHeader == "" {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeUnauthorized,
					"Authorization header is required",
					http.StatusUnauthorized,
				))
				return
			}

			// Check Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeUnauthorized,
					"Invalid authorization format. Use: Bearer <token>",
					http.StatusUnauthorized,
				))
				return
			}

			// Extract token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token
			userID, err := authService.ValidateAccessToken(tokenString)
			if err != nil {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeInvalidToken,
					"Invalid or expired access token",
					http.StatusUnauthorized,
				))
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(request.Context(), "userID", userID)
			request = request.WithContext(ctx)

			// Proceed to next handler
			next.ServeHTTP(responseWriter, request)
		})
	}
}

// OptionalAuthMiddleware creates middleware that validates JWT tokens if present
// but allows requests without tokens to proceed
func OptionalAuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Extract Authorization header
			authHeader := request.Header.Get("Authorization")

			// If no auth header, proceed without user context
			if authHeader == "" {
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Check Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Extract and validate token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := authService.ValidateAccessToken(tokenString)
			if err != nil {
				// Token invalid, proceed without user context
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(request.Context(), "userID", userID)
			request = request.WithContext(ctx)

			next.ServeHTTP(responseWriter, request)
		})
	}
}
