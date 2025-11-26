package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/OPGLOL/opgl-gateway-service/internal/api"
	"github.com/OPGLOL/opgl-gateway-service/internal/middleware"
	"github.com/OPGLOL/opgl-gateway-service/internal/proxy"
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

	log.Info().
		Str("port", port).
		Str("data_service_url", dataServiceURL).
		Str("cortex_service_url", cortexServiceURL).
		Msg("Configuration loaded")

	// Initialize service proxy
	serviceProxy := proxy.NewServiceProxy(dataServiceURL, cortexServiceURL)

	// Initialize HTTP handler
	handler := api.NewHandler(serviceProxy)

	// Set up router
	router := api.SetupRouter(handler)

	// Wrap router with logging middleware
	loggedRouter := middleware.LoggingMiddleware(router)

	// Start server
	serverAddress := fmt.Sprintf(":%s", port)
	log.Info().
		Str("address", serverAddress).
		Str("port", port).
		Msg("OPGL Gateway listening")

	if err := http.ListenAndServe(serverAddress, loggedRouter); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
