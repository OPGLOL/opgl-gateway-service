package api

import (
	"net/http"

	"github.com/OPGLOL/opgl-gateway-service/internal/middleware"
	"github.com/OPGLOL/opgl-gateway-service/internal/ratelimit"
	"github.com/gorilla/mux"
)

// SetupRouter configures all routes for the gateway
// rateLimiter can be nil if rate limiting is not enabled
func SetupRouter(handler *Handler, adminHandler *AdminHandler, rateLimiter *ratelimit.RateLimiter) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint - no rate limiting
	router.HandleFunc("/health", handler.HealthCheck).Methods("POST")

	// API routes subrouter
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Apply rate limiting middleware if configured
	if rateLimiter != nil {
		apiRouter.Use(middleware.RateLimitMiddleware(rateLimiter))
	}

	// Proxied data endpoints (rate limited)
	apiRouter.HandleFunc("/summoner", handler.GetSummoner).Methods("POST")
	apiRouter.HandleFunc("/matches", handler.GetMatches).Methods("POST")

	// Orchestrated analysis endpoint (rate limited)
	apiRouter.HandleFunc("/analyze", handler.AnalyzePlayer).Methods("POST")

	// Admin endpoints (no rate limiting for admin, but require separate authentication in future)
	if adminHandler != nil {
		adminRouter := router.PathPrefix("/api/v1/admin").Subrouter()
		adminRouter.HandleFunc("/apikeys", adminHandler.CreateAPIKey).Methods("POST")
		adminRouter.HandleFunc("/apikeys/list", adminHandler.ListAPIKeys).Methods("POST")
		adminRouter.HandleFunc("/apikeys/delete", adminHandler.DeleteAPIKey).Methods("POST")
		adminRouter.HandleFunc("/apikeys/{id}", adminHandler.DeleteAPIKey).Methods("DELETE")
	}

	return router
}

// SetupRouterWithoutRateLimit configures routes without rate limiting (for backward compatibility)
func SetupRouterWithoutRateLimit(handler *Handler) http.Handler {
	return SetupRouter(handler, nil, nil)
}
