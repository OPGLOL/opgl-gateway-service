package api

import (
	"net/http"

	"github.com/OPGLOL/opgl-gateway-service/internal/auth"
	"github.com/OPGLOL/opgl-gateway-service/internal/middleware"
	"github.com/OPGLOL/opgl-gateway-service/internal/ratelimit"
	"github.com/gorilla/mux"
)

// RouterConfig holds all dependencies for router setup
type RouterConfig struct {
	Handler      *Handler
	AdminHandler *AdminHandler
	AuthHandler  *AuthHandler
	RateLimiter  *ratelimit.RateLimiter
	AuthService  *auth.AuthService
}

// SetupRouter configures all routes for the gateway
func SetupRouter(config *RouterConfig) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint - no rate limiting, no auth
	router.HandleFunc("/health", config.Handler.HealthCheck).Methods("POST")

	// Auth routes - no rate limiting, no auth required (public endpoints)
	if config.AuthHandler != nil {
		authRouter := router.PathPrefix("/api/v1/auth").Subrouter()
		authRouter.HandleFunc("/register", config.AuthHandler.Register).Methods("POST")
		authRouter.HandleFunc("/login", config.AuthHandler.Login).Methods("POST")
		authRouter.HandleFunc("/refresh", config.AuthHandler.Refresh).Methods("POST")

		// Protected auth endpoint - requires valid access token
		if config.AuthService != nil {
			protectedAuthRouter := router.PathPrefix("/api/v1/auth").Subrouter()
			protectedAuthRouter.Use(middleware.AuthMiddleware(config.AuthService))
			protectedAuthRouter.HandleFunc("/me", config.AuthHandler.Me).Methods("POST")
		}
	}

	// API routes subrouter
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Apply rate limiting middleware if configured
	if config.RateLimiter != nil {
		apiRouter.Use(middleware.RateLimitMiddleware(config.RateLimiter))
	}

	// Proxied data endpoints (rate limited)
	apiRouter.HandleFunc("/summoner", config.Handler.GetSummoner).Methods("POST")
	apiRouter.HandleFunc("/matches", config.Handler.GetMatches).Methods("POST")

	// Orchestrated analysis endpoint (rate limited)
	apiRouter.HandleFunc("/analyze", config.Handler.AnalyzePlayer).Methods("POST")

	// Admin endpoints (no rate limiting for admin)
	if config.AdminHandler != nil {
		adminRouter := router.PathPrefix("/api/v1/admin").Subrouter()
		adminRouter.HandleFunc("/apikeys", config.AdminHandler.CreateAPIKey).Methods("POST")
		adminRouter.HandleFunc("/apikeys/list", config.AdminHandler.ListAPIKeys).Methods("POST")
		adminRouter.HandleFunc("/apikeys/delete", config.AdminHandler.DeleteAPIKey).Methods("POST")
		adminRouter.HandleFunc("/apikeys/{id}", config.AdminHandler.DeleteAPIKey).Methods("DELETE")
	}

	return router
}

// SetupRouterSimple configures routes with minimal dependencies (for testing)
func SetupRouterSimple(handler *Handler, adminHandler *AdminHandler, rateLimiter *ratelimit.RateLimiter) *mux.Router {
	return SetupRouter(&RouterConfig{
		Handler:      handler,
		AdminHandler: adminHandler,
		RateLimiter:  rateLimiter,
	})
}

// SetupRouterWithoutRateLimit configures routes without rate limiting (for backward compatibility)
func SetupRouterWithoutRateLimit(handler *Handler) http.Handler {
	return SetupRouterSimple(handler, nil, nil)
}
