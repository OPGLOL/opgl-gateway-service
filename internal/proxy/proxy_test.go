package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OPGLOL/opgl-gateway-service/internal/models"
)

// TestNewServiceProxy tests the NewServiceProxy constructor
func TestNewServiceProxy(t *testing.T) {
	dataURL := "http://localhost:8081"
	cortexURL := "http://localhost:8082"

	proxy := NewServiceProxy(dataURL, cortexURL)

	if proxy == nil {
		t.Fatal("Expected proxy to not be nil")
	}

	if proxy.dataServiceURL != dataURL {
		t.Errorf("Expected dataServiceURL '%s', got '%s'", dataURL, proxy.dataServiceURL)
	}

	if proxy.cortexServiceURL != cortexURL {
		t.Errorf("Expected cortexServiceURL '%s', got '%s'", cortexURL, proxy.cortexServiceURL)
	}

	if proxy.httpClient == nil {
		t.Error("Expected httpClient to not be nil")
	}
}

// TestGetSummonerByRiotID_Success tests successful summoner lookup
func TestGetSummonerByRiotID_Success(t *testing.T) {
	expectedSummoner := models.Summoner{
		ID:            "test-id",
		AccountID:     "test-account-id",
		PUUID:         "test-puuid",
		Name:          "TestPlayer",
		ProfileIconID: 1234,
		SummonerLevel: 100,
	}

	// Create mock data service server
	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/summoner" {
			t.Errorf("Expected path '/api/v1/summoner', got '%s'", request.URL.Path)
		}

		if request.Method != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", request.Method)
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(expectedSummoner)
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy(mockServer.URL, "http://localhost:8082")

	summoner, err := proxy.GetSummonerByRiotID("na", "TestPlayer", "NA1")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if summoner.PUUID != expectedSummoner.PUUID {
		t.Errorf("Expected PUUID '%s', got '%s'", expectedSummoner.PUUID, summoner.PUUID)
	}
}

// TestGetSummonerByRiotID_ServerError tests server error handling
func TestGetSummonerByRiotID_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy(mockServer.URL, "http://localhost:8082")

	summoner, err := proxy.GetSummonerByRiotID("na", "TestPlayer", "NA1")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if summoner != nil {
		t.Error("Expected summoner to be nil on error")
	}
}

// TestGetSummonerByRiotID_ConnectionError tests connection error handling
func TestGetSummonerByRiotID_ConnectionError(t *testing.T) {
	// Use invalid URL to simulate connection error
	proxy := NewServiceProxy("http://localhost:99999", "http://localhost:8082")

	summoner, err := proxy.GetSummonerByRiotID("na", "TestPlayer", "NA1")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if summoner != nil {
		t.Error("Expected summoner to be nil on error")
	}
}

// TestGetSummonerByRiotID_InvalidJSON tests invalid JSON response handling
func TestGetSummonerByRiotID_InvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.Write([]byte("invalid json"))
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy(mockServer.URL, "http://localhost:8082")

	summoner, err := proxy.GetSummonerByRiotID("na", "TestPlayer", "NA1")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if summoner != nil {
		t.Error("Expected summoner to be nil on error")
	}
}

// TestGetMatchesByRiotID_Success tests successful match history lookup
func TestGetMatchesByRiotID_Success(t *testing.T) {
	expectedMatches := []models.Match{
		{MatchID: "NA1_123", GameMode: "CLASSIC"},
		{MatchID: "NA1_124", GameMode: "ARAM"},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/matches" {
			t.Errorf("Expected path '/api/v1/matches', got '%s'", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(expectedMatches)
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy(mockServer.URL, "http://localhost:8082")

	matches, err := proxy.GetMatchesByRiotID("na", "TestPlayer", "NA1", 10)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(matches) != len(expectedMatches) {
		t.Errorf("Expected %d matches, got %d", len(expectedMatches), len(matches))
	}
}

// TestGetMatchesByRiotID_ServerError tests server error handling
func TestGetMatchesByRiotID_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy(mockServer.URL, "http://localhost:8082")

	matches, err := proxy.GetMatchesByRiotID("na", "TestPlayer", "NA1", 10)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if matches != nil {
		t.Error("Expected matches to be nil on error")
	}
}

// TestGetMatchesByPUUID_Success tests successful match history lookup by PUUID
func TestGetMatchesByPUUID_Success(t *testing.T) {
	expectedMatches := []models.Match{
		{MatchID: "NA1_123", GameMode: "CLASSIC"},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(expectedMatches)
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy(mockServer.URL, "http://localhost:8082")

	matches, err := proxy.GetMatchesByPUUID("na", "test-puuid", 20)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(matches) != len(expectedMatches) {
		t.Errorf("Expected %d matches, got %d", len(expectedMatches), len(matches))
	}
}

// TestGetMatchesByPUUID_ServerError tests server error handling
func TestGetMatchesByPUUID_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy(mockServer.URL, "http://localhost:8082")

	matches, err := proxy.GetMatchesByPUUID("na", "test-puuid", 20)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if matches != nil {
		t.Error("Expected matches to be nil on error")
	}
}

// TestAnalyzePlayer_Success tests successful player analysis
func TestAnalyzePlayer_Success(t *testing.T) {
	expectedResult := models.AnalysisResult{
		PlayerStats:      map[string]interface{}{"avgKills": 5.5},
		ImprovementAreas: []string{"CS improvement"},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/analyze" {
			t.Errorf("Expected path '/api/v1/analyze', got '%s'", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(expectedResult)
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy("http://localhost:8081", mockServer.URL)

	summoner := &models.Summoner{PUUID: "test-puuid"}
	matches := []models.Match{{MatchID: "NA1_123"}}

	result, err := proxy.AnalyzePlayer(summoner, matches)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}
}

// TestAnalyzePlayer_ServerError tests server error handling
func TestAnalyzePlayer_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	proxy := NewServiceProxy("http://localhost:8081", mockServer.URL)

	summoner := &models.Summoner{PUUID: "test-puuid"}
	matches := []models.Match{{MatchID: "NA1_123"}}

	result, err := proxy.AnalyzePlayer(summoner, matches)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil on error")
	}
}

// TestAnalyzePlayer_ConnectionError tests connection error handling
func TestAnalyzePlayer_ConnectionError(t *testing.T) {
	proxy := NewServiceProxy("http://localhost:8081", "http://localhost:99999")

	summoner := &models.Summoner{PUUID: "test-puuid"}
	matches := []models.Match{{MatchID: "NA1_123"}}

	result, err := proxy.AnalyzePlayer(summoner, matches)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil on error")
	}
}

// TestServiceProxyImplementsInterface verifies ServiceProxy implements ServiceProxyInterface
func TestServiceProxyImplementsInterface(t *testing.T) {
	var proxyInterface ServiceProxyInterface = NewServiceProxy("http://localhost:8081", "http://localhost:8082")

	if proxyInterface == nil {
		t.Error("ServiceProxy should implement ServiceProxyInterface")
	}
}
