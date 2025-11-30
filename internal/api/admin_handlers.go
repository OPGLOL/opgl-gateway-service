package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	apierrors "github.com/OPGLOL/opgl-gateway-service/internal/errors"
	"github.com/OPGLOL/opgl-gateway-service/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AdminHandler manages admin HTTP request handlers
type AdminHandler struct {
	apiKeyRepository repository.APIKeyRepository
}

// NewAdminHandler creates a new AdminHandler instance
func NewAdminHandler(apiKeyRepository repository.APIKeyRepository) *AdminHandler {
	return &AdminHandler{
		apiKeyRepository: apiKeyRepository,
	}
}

// CreateAPIKeyRequest represents the request body for creating an API key
type CreateAPIKeyRequest struct {
	Name              string `json:"name"`
	RateLimit         int    `json:"rateLimit"`
	RateWindowSeconds int    `json:"rateWindowSeconds"`
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	ID                string `json:"id"`
	APIKey            string `json:"apiKey"`
	Name              string `json:"name"`
	RateLimit         int    `json:"rateLimit"`
	RateWindowSeconds int    `json:"rateWindowSeconds"`
}

// APIKeyListItem represents an API key in list responses (without the actual key)
type APIKeyListItem struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	RateLimit         int    `json:"rateLimit"`
	RateWindowSeconds int    `json:"rateWindowSeconds"`
	IsActive          bool   `json:"isActive"`
	CreatedAt         string `json:"createdAt"`
	LastUsedAt        string `json:"lastUsedAt,omitempty"`
}

// CreateAPIKey handles POST /api/v1/admin/apikeys
func (adminHandler *AdminHandler) CreateAPIKey(writer http.ResponseWriter, request *http.Request) {
	var createRequest CreateAPIKeyRequest

	if err := json.NewDecoder(request.Body).Decode(&createRequest); err != nil {
		apierrors.WriteError(writer, apierrors.InvalidRequestBody("Invalid JSON format"))
		return
	}

	// Validate required fields
	if createRequest.Name == "" {
		apierrors.WriteError(writer, apierrors.ValidationFailed("name is required"))
		return
	}

	// Set defaults if not provided
	rateLimit := createRequest.RateLimit
	if rateLimit <= 0 {
		rateLimit = 100
	}

	rateWindowSeconds := createRequest.RateWindowSeconds
	if rateWindowSeconds <= 0 {
		rateWindowSeconds = 60
	}

	// Generate a random API key (32 bytes = 64 hex characters)
	apiKeyBytes := make([]byte, 32)
	if _, err := rand.Read(apiKeyBytes); err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to generate API key"))
		return
	}
	apiKey := hex.EncodeToString(apiKeyBytes)

	// Hash the API key for storage
	keyHash := repository.HashAPIKey(apiKey)

	// Create the API key record
	apiKeyRecord, err := adminHandler.apiKeyRepository.Create(createRequest.Name, keyHash, rateLimit, rateWindowSeconds)
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to create API key"))
		return
	}

	// Return the response with the plain API key (only shown once)
	response := CreateAPIKeyResponse{
		ID:                apiKeyRecord.ID.String(),
		APIKey:            apiKey,
		Name:              apiKeyRecord.Name,
		RateLimit:         apiKeyRecord.RateLimit,
		RateWindowSeconds: apiKeyRecord.RateWindowSeconds,
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(response)
}

// ListAPIKeys handles POST /api/v1/admin/apikeys/list
func (adminHandler *AdminHandler) ListAPIKeys(writer http.ResponseWriter, request *http.Request) {
	apiKeys, err := adminHandler.apiKeyRepository.List()
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to list API keys"))
		return
	}

	// Convert to response format (without exposing key hashes)
	var responseItems []APIKeyListItem
	for _, apiKey := range apiKeys {
		item := APIKeyListItem{
			ID:                apiKey.ID.String(),
			Name:              apiKey.Name,
			RateLimit:         apiKey.RateLimit,
			RateWindowSeconds: apiKey.RateWindowSeconds,
			IsActive:          apiKey.IsActive,
			CreatedAt:         apiKey.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if apiKey.LastUsedAt.Valid {
			item.LastUsedAt = apiKey.LastUsedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		responseItems = append(responseItems, item)
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(responseItems)
}

// DeleteAPIKeyRequest represents the request body for deleting an API key
type DeleteAPIKeyRequest struct {
	ID string `json:"id"`
}

// DeleteAPIKey handles POST /api/v1/admin/apikeys/delete
func (adminHandler *AdminHandler) DeleteAPIKey(writer http.ResponseWriter, request *http.Request) {
	// Try to get ID from URL path first (for RESTful style)
	vars := mux.Vars(request)
	idString := vars["id"]

	// If not in path, try request body
	if idString == "" {
		var deleteRequest DeleteAPIKeyRequest
		if err := json.NewDecoder(request.Body).Decode(&deleteRequest); err != nil {
			apierrors.WriteError(writer, apierrors.InvalidRequestBody("Invalid JSON format"))
			return
		}
		idString = deleteRequest.ID
	}

	if idString == "" {
		apierrors.WriteError(writer, apierrors.ValidationFailed("id is required"))
		return
	}

	// Parse UUID
	id, err := uuid.Parse(idString)
	if err != nil {
		apierrors.WriteError(writer, apierrors.ValidationFailed("invalid id format"))
		return
	}

	// Delete (soft delete) the API key
	if err := adminHandler.apiKeyRepository.Delete(id); err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to delete API key"))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]string{
		"message": "API key revoked successfully",
	})
}
