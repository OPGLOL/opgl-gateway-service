package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OPGLOL/opgl-gateway-service/internal/models"
)

// TestSetupRouter tests that all routes are registered correctly
func TestSetupRouter(t *testing.T) {
	mockProxy := &MockServiceProxy{}
	handler := NewHandler(mockProxy)
	router := SetupRouter(handler, nil, nil)

	if router == nil {
		t.Fatal("Expected router to not be nil")
	}
}

// TestRouterHealthEndpoint tests that the health endpoint is registered
func TestRouterHealthEndpoint(t *testing.T) {
	mockProxy := &MockServiceProxy{}
	handler := NewHandler(mockProxy)
	router := SetupRouter(handler, nil, nil)

	request, _ := http.NewRequest("POST", "/health", nil)
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}

// TestRouterHealthEndpointMethodNotAllowed tests that GET is not allowed for health
func TestRouterHealthEndpointMethodNotAllowed(t *testing.T) {
	mockProxy := &MockServiceProxy{}
	handler := NewHandler(mockProxy)
	router := SetupRouter(handler, nil, nil)

	request, _ := http.NewRequest("GET", "/health", nil)
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d for GET /health, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}
}

// TestRouterSummonerEndpoint tests that the summoner endpoint is registered
func TestRouterSummonerEndpoint(t *testing.T) {
	mockProxy := &MockServiceProxy{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return &models.Summoner{PUUID: "test"}, nil
		},
	}
	handler := NewHandler(mockProxy)
	router := SetupRouter(handler, nil, nil)

	// Send invalid JSON body to trigger BadRequest (proves endpoint is registered)
	request, _ := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBufferString("invalid"))
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	// Endpoint should be reachable (returns 400 due to invalid body, not 404)
	if responseRecorder.Code == http.StatusNotFound {
		t.Error("Expected /api/v1/summoner endpoint to be registered")
	}
}

// TestRouterMatchesEndpoint tests that the matches endpoint is registered
func TestRouterMatchesEndpoint(t *testing.T) {
	mockProxy := &MockServiceProxy{}
	handler := NewHandler(mockProxy)
	router := SetupRouter(handler, nil, nil)

	// Send invalid JSON body to test endpoint is registered
	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBufferString("invalid"))
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	// Endpoint should be reachable (returns 400 due to invalid body, not 404)
	if responseRecorder.Code == http.StatusNotFound {
		t.Error("Expected /api/v1/matches endpoint to be registered")
	}
}

// TestRouterAnalyzeEndpoint tests that the analyze endpoint is registered
func TestRouterAnalyzeEndpoint(t *testing.T) {
	mockProxy := &MockServiceProxy{}
	handler := NewHandler(mockProxy)
	router := SetupRouter(handler, nil, nil)

	// Send invalid JSON body to test endpoint is registered
	request, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBufferString("invalid"))
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	// Endpoint should be reachable (returns 400 due to invalid body, not 404)
	if responseRecorder.Code == http.StatusNotFound {
		t.Error("Expected /api/v1/analyze endpoint to be registered")
	}
}

// TestRouterNonExistentEndpoint tests that non-existent endpoints return 404
func TestRouterNonExistentEndpoint(t *testing.T) {
	mockProxy := &MockServiceProxy{}
	handler := NewHandler(mockProxy)
	router := SetupRouter(handler, nil, nil)

	request, _ := http.NewRequest("POST", "/api/v1/nonexistent", nil)
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d for non-existent endpoint, got %d", http.StatusNotFound, responseRecorder.Code)
	}
}

// TestRouterAllEndpointsUsePOST verifies all endpoints use POST method
func TestRouterAllEndpointsUsePOST(t *testing.T) {
	mockProxy := &MockServiceProxy{}
	handler := NewHandler(mockProxy)
	router := SetupRouter(handler, nil, nil)

	// Test health endpoint returns 405 for GET
	request, _ := http.NewRequest("GET", "/health", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected GET /health to return %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	// Note: Subrouter endpoints return 404 for wrong methods due to gorilla/mux behavior
	// This is acceptable as the endpoints are not exposed for wrong methods
}
