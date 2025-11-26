package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OPGLOL/opgl-gateway-service/internal/models"
)

// MockServiceProxy is a mock implementation of ServiceProxyInterface for testing
type MockServiceProxy struct {
	GetSummonerByRiotIDFunc func(region, gameName, tagLine string) (*models.Summoner, error)
	GetMatchesByRiotIDFunc  func(region, gameName, tagLine string, count int) ([]models.Match, error)
	GetMatchesByPUUIDFunc   func(region, puuid string, count int) ([]models.Match, error)
	AnalyzePlayerFunc       func(summoner *models.Summoner, matches []models.Match) (*models.AnalysisResult, error)
}

func (m *MockServiceProxy) GetSummonerByRiotID(region, gameName, tagLine string) (*models.Summoner, error) {
	if m.GetSummonerByRiotIDFunc != nil {
		return m.GetSummonerByRiotIDFunc(region, gameName, tagLine)
	}
	return nil, nil
}

func (m *MockServiceProxy) GetMatchesByRiotID(region, gameName, tagLine string, count int) ([]models.Match, error) {
	if m.GetMatchesByRiotIDFunc != nil {
		return m.GetMatchesByRiotIDFunc(region, gameName, tagLine, count)
	}
	return nil, nil
}

func (m *MockServiceProxy) GetMatchesByPUUID(region, puuid string, count int) ([]models.Match, error) {
	if m.GetMatchesByPUUIDFunc != nil {
		return m.GetMatchesByPUUIDFunc(region, puuid, count)
	}
	return nil, nil
}

func (m *MockServiceProxy) AnalyzePlayer(summoner *models.Summoner, matches []models.Match) (*models.AnalysisResult, error) {
	if m.AnalyzePlayerFunc != nil {
		return m.AnalyzePlayerFunc(summoner, matches)
	}
	return nil, nil
}

// TestNewHandler tests the NewHandler constructor
func TestNewHandler(t *testing.T) {
	mockProxy := &MockServiceProxy{}
	handler := NewHandler(mockProxy)

	if handler == nil {
		t.Fatal("Expected handler to not be nil")
	}

	if handler.serviceProxy != mockProxy {
		t.Error("Expected serviceProxy to be set correctly")
	}
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	handler := &Handler{serviceProxy: nil}

	request, err := http.NewRequest("POST", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	handler.HealthCheck(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response map[string]string
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}

	if response["service"] != "opgl-gateway" {
		t.Errorf("Expected service 'opgl-gateway', got '%s'", response["service"])
	}
}

// TestHealthCheckContentType tests that health check returns JSON content type
func TestHealthCheckContentType(t *testing.T) {
	handler := &Handler{serviceProxy: nil}

	request, err := http.NewRequest("POST", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	handler.HealthCheck(responseRecorder, request)

	contentType := responseRecorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// TestGetSummoner_Success tests successful summoner lookup
func TestGetSummoner_Success(t *testing.T) {
	expectedSummoner := &models.Summoner{
		ID:            "test-id",
		AccountID:     "test-account-id",
		PUUID:         "test-puuid",
		Name:          "TestPlayer",
		ProfileIconID: 1234,
		SummonerLevel: 100,
	}

	mockProxy := &MockServiceProxy{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			if region != "na" || gameName != "TestPlayer" || tagLine != "NA1" {
				t.Errorf("Unexpected parameters: region=%s, gameName=%s, tagLine=%s", region, gameName, tagLine)
			}
			return expectedSummoner, nil
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]string{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, err := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetSummoner(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response models.Summoner
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.PUUID != expectedSummoner.PUUID {
		t.Errorf("Expected PUUID '%s', got '%s'", expectedSummoner.PUUID, response.PUUID)
	}
}

// TestGetSummoner_InvalidJSON tests invalid JSON request body
func TestGetSummoner_InvalidJSON(t *testing.T) {
	handler := NewHandler(&MockServiceProxy{})

	request, err := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBufferString("invalid json"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	handler.GetSummoner(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestGetSummoner_MissingFields tests missing required fields
func TestGetSummoner_MissingFields(t *testing.T) {
	testCases := []struct {
		name        string
		requestBody map[string]string
	}{
		{"missing region", map[string]string{"gameName": "Test", "tagLine": "NA1"}},
		{"missing gameName", map[string]string{"region": "na", "tagLine": "NA1"}},
		{"missing tagLine", map[string]string{"region": "na", "gameName": "Test"}},
		{"empty region", map[string]string{"region": "", "gameName": "Test", "tagLine": "NA1"}},
		{"empty gameName", map[string]string{"region": "na", "gameName": "", "tagLine": "NA1"}},
		{"empty tagLine", map[string]string{"region": "na", "gameName": "Test", "tagLine": ""}},
	}

	handler := NewHandler(&MockServiceProxy{})

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(testCase.requestBody)
			request, _ := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBuffer(bodyBytes))
			request.Header.Set("Content-Type", "application/json")

			responseRecorder := httptest.NewRecorder()
			handler.GetSummoner(responseRecorder, request)

			if responseRecorder.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
			}
		})
	}
}

// TestGetSummoner_ServiceError tests service error handling
func TestGetSummoner_ServiceError(t *testing.T) {
	mockProxy := &MockServiceProxy{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return nil, errors.New("service error")
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]string{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetSummoner(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestGetMatches_Success tests successful match history lookup
func TestGetMatches_Success(t *testing.T) {
	expectedMatches := []models.Match{
		{MatchID: "NA1_123", GameMode: "CLASSIC"},
		{MatchID: "NA1_124", GameMode: "CLASSIC"},
	}

	mockProxy := &MockServiceProxy{
		GetMatchesByRiotIDFunc: func(region, gameName, tagLine string, count int) ([]models.Match, error) {
			if region != "na" || gameName != "TestPlayer" || tagLine != "NA1" {
				t.Errorf("Unexpected parameters: region=%s, gameName=%s, tagLine=%s", region, gameName, tagLine)
			}
			return expectedMatches, nil
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]interface{}{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
		"count":    10,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatches(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response []models.Match
	err := json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != len(expectedMatches) {
		t.Errorf("Expected %d matches, got %d", len(expectedMatches), len(response))
	}
}

// TestGetMatches_DefaultCount tests default count when not provided
func TestGetMatches_DefaultCount(t *testing.T) {
	var capturedCount int

	mockProxy := &MockServiceProxy{
		GetMatchesByRiotIDFunc: func(region, gameName, tagLine string, count int) ([]models.Match, error) {
			capturedCount = count
			return []models.Match{}, nil
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]interface{}{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatches(responseRecorder, request)

	if capturedCount != 20 {
		t.Errorf("Expected default count 20, got %d", capturedCount)
	}
}

// TestGetMatches_InvalidJSON tests invalid JSON request body
func TestGetMatches_InvalidJSON(t *testing.T) {
	handler := NewHandler(&MockServiceProxy{})

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBufferString("invalid json"))

	responseRecorder := httptest.NewRecorder()
	handler.GetMatches(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestGetMatches_MissingFields tests missing required fields
func TestGetMatches_MissingFields(t *testing.T) {
	testCases := []struct {
		name        string
		requestBody map[string]interface{}
	}{
		{"missing region", map[string]interface{}{"gameName": "Test", "tagLine": "NA1"}},
		{"missing gameName", map[string]interface{}{"region": "na", "tagLine": "NA1"}},
		{"missing tagLine", map[string]interface{}{"region": "na", "gameName": "Test"}},
	}

	handler := NewHandler(&MockServiceProxy{})

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(testCase.requestBody)
			request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
			request.Header.Set("Content-Type", "application/json")

			responseRecorder := httptest.NewRecorder()
			handler.GetMatches(responseRecorder, request)

			if responseRecorder.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
			}
		})
	}
}

// TestGetMatches_ServiceError tests service error handling
func TestGetMatches_ServiceError(t *testing.T) {
	mockProxy := &MockServiceProxy{
		GetMatchesByRiotIDFunc: func(region, gameName, tagLine string, count int) ([]models.Match, error) {
			return nil, errors.New("service error")
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]interface{}{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatches(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestAnalyzePlayer_Success tests successful player analysis
func TestAnalyzePlayer_Success(t *testing.T) {
	expectedSummoner := &models.Summoner{
		PUUID: "test-puuid",
		Name:  "TestPlayer",
	}
	expectedMatches := []models.Match{
		{MatchID: "NA1_123", GameMode: "CLASSIC"},
	}
	expectedAnalysis := &models.AnalysisResult{
		PlayerStats:      map[string]interface{}{"avgKills": 5.5},
		ImprovementAreas: []string{"CS improvement"},
		AnalyzedAt:       time.Now(),
	}

	mockProxy := &MockServiceProxy{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return expectedSummoner, nil
		},
		GetMatchesByPUUIDFunc: func(region, puuid string, count int) ([]models.Match, error) {
			if puuid != expectedSummoner.PUUID {
				t.Errorf("Expected PUUID '%s', got '%s'", expectedSummoner.PUUID, puuid)
			}
			return expectedMatches, nil
		},
		AnalyzePlayerFunc: func(summoner *models.Summoner, matches []models.Match) (*models.AnalysisResult, error) {
			return expectedAnalysis, nil
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]string{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.AnalyzePlayer(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}

// TestAnalyzePlayer_InvalidJSON tests invalid JSON request body
func TestAnalyzePlayer_InvalidJSON(t *testing.T) {
	handler := NewHandler(&MockServiceProxy{})

	request, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBufferString("invalid json"))

	responseRecorder := httptest.NewRecorder()
	handler.AnalyzePlayer(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestAnalyzePlayer_MissingFields tests missing required fields
func TestAnalyzePlayer_MissingFields(t *testing.T) {
	testCases := []struct {
		name        string
		requestBody map[string]string
	}{
		{"missing region", map[string]string{"gameName": "Test", "tagLine": "NA1"}},
		{"missing gameName", map[string]string{"region": "na", "tagLine": "NA1"}},
		{"missing tagLine", map[string]string{"region": "na", "gameName": "Test"}},
	}

	handler := NewHandler(&MockServiceProxy{})

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(testCase.requestBody)
			request, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(bodyBytes))
			request.Header.Set("Content-Type", "application/json")

			responseRecorder := httptest.NewRecorder()
			handler.AnalyzePlayer(responseRecorder, request)

			if responseRecorder.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
			}
		})
	}
}

// TestAnalyzePlayer_SummonerError tests error during summoner lookup
func TestAnalyzePlayer_SummonerError(t *testing.T) {
	mockProxy := &MockServiceProxy{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return nil, errors.New("summoner not found")
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]string{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.AnalyzePlayer(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestAnalyzePlayer_MatchHistoryError tests error during match history lookup
func TestAnalyzePlayer_MatchHistoryError(t *testing.T) {
	mockProxy := &MockServiceProxy{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return &models.Summoner{PUUID: "test-puuid"}, nil
		},
		GetMatchesByPUUIDFunc: func(region, puuid string, count int) ([]models.Match, error) {
			return nil, errors.New("match history error")
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]string{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.AnalyzePlayer(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestAnalyzePlayer_AnalysisError tests error during analysis
func TestAnalyzePlayer_AnalysisError(t *testing.T) {
	mockProxy := &MockServiceProxy{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return &models.Summoner{PUUID: "test-puuid"}, nil
		},
		GetMatchesByPUUIDFunc: func(region, puuid string, count int) ([]models.Match, error) {
			return []models.Match{}, nil
		},
		AnalyzePlayerFunc: func(summoner *models.Summoner, matches []models.Match) (*models.AnalysisResult, error) {
			return nil, errors.New("analysis error")
		},
	}

	handler := NewHandler(mockProxy)

	requestBody := map[string]string{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.AnalyzePlayer(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}
