package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OPGLOL/opgl-gateway-service/internal/api"
	"github.com/OPGLOL/opgl-gateway-service/internal/db"
	"github.com/OPGLOL/opgl-gateway-service/internal/middleware"
	"github.com/OPGLOL/opgl-gateway-service/internal/proxy"
	"github.com/OPGLOL/opgl-gateway-service/internal/ratelimit"
	"github.com/OPGLOL/opgl-gateway-service/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize zerolog with colorized console output for development
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Caller().Logger()

	// Set global log level (can be configured via environment variable)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("Starting OPGL Gateway")

	// Get configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataServiceURL := os.Getenv("OPGL_DATA_URL")
	if dataServiceURL == "" {
		dataServiceURL = "http://localhost:8081"
	}

	cortexServiceURL := os.Getenv("OPGL_CORTEX_URL")
	if cortexServiceURL == "" {
		cortexServiceURL = "http://localhost:8082"
	}

	// Database configuration
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	log.Info().
		Str("port", port).
		Str("data_service_url", dataServiceURL).
		Str("cortex_service_url", cortexServiceURL).
		Msg("Configuration loaded")

	// Initialize database connection (optional - service runs without DB if not configured)
	var database *db.Database
	if dbHost != "" && dbPassword != "" {
		var err error
		database, err = db.NewPostgresConnection(dbHost, dbPort, dbUser, dbPassword, dbName)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to connect to database")
		}
		log.Info().
			Str("host", dbHost).
			Str("port", dbPort).
			Str("database", dbName).
			Msg("Database connection established")
	} else {
		log.Warn().Msg("Database not configured - running without database")
	}

	// Initialize service proxy
	serviceProxy := proxy.NewServiceProxy(dataServiceURL, cortexServiceURL)

	// Initialize HTTP handler
	handler := api.NewHandler(serviceProxy)

	// Initialize rate limiting and admin handler if database is configured
	var rateLimiter *ratelimit.RateLimiter
	var adminHandler *api.AdminHandler

	if database != nil {
		// Initialize API key repository
		apiKeyRepository := repository.NewPostgresAPIKeyRepository(database.DB)

		// Initialize rate limiter
		rateLimiter = ratelimit.NewRateLimiter(apiKeyRepository)
		log.Info().Msg("Rate limiting enabled")

		// Initialize admin handler
		adminHandler = api.NewAdminHandler(apiKeyRepository)
		log.Info().Msg("Admin endpoints enabled")
	} else {
		log.Warn().Msg("Rate limiting disabled - database not configured")
	}

	// Set up router with rate limiting
	router := api.SetupRouter(handler, adminHandler, rateLimiter)

	// Wrap router with CORS middleware first to handle preflight requests
	corsRouter := middleware.CORSMiddleware(router)

	// Wrap with logging middleware
	loggedRouter := middleware.LoggingMiddleware(corsRouter)

	// Create HTTP server
	serverAddress := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:    serverAddress,
		Handler: loggedRouter,
	}

	// Channel to listen for shutdown signals
	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Info().
			Str("address", serverAddress).
			Str("port", port).
			Msg("OPGL Gateway listening")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for shutdown signal
	<-shutdownChannel
	log.Info().Msg("Shutting down server...")

	// Create shutdown context with timeout
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// Gracefully shutdown HTTP server
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}

	// Close database connection if established
	if database != nil {
		if err := database.Close(); err != nil {
			log.Error().Err(err).Msg("Database close error")
		} else {
			log.Info().Msg("Database connection closed")
		}
	}

	log.Info().Msg("Server stopped")
}
