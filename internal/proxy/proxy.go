package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/OPGLOL/opgl-gateway-service/internal/models"
)

// ServiceProxy handles communication with microservices
type ServiceProxy struct {
	dataServiceURL   string
	cortexServiceURL string
	httpClient       *http.Client
}

// NewServiceProxy creates a new ServiceProxy instance
func NewServiceProxy(dataServiceURL string, cortexServiceURL string) *ServiceProxy {
	return &ServiceProxy{
		dataServiceURL:   dataServiceURL,
		cortexServiceURL: cortexServiceURL,
		httpClient:       &http.Client{},
	}
}

// GetSummonerByRiotID retrieves summoner data from opgl-data service using Riot ID
func (proxy *ServiceProxy) GetSummonerByRiotID(region string, gameName string, tagLine string) (*models.Summoner, error) {
	url := fmt.Sprintf("%s/api/v1/summoner", proxy.dataServiceURL)

	requestBody := map[string]string{
		"region":   region,
		"gameName": gameName,
		"tagLine":  tagLine,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	response, err := proxy.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call data service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("data service returned error %d: %s", response.StatusCode, string(body))
	}

	var summoner models.Summoner
	if err := json.NewDecoder(response.Body).Decode(&summoner); err != nil {
		return nil, fmt.Errorf("failed to decode summoner response: %w", err)
	}

	return &summoner, nil
}

// GetMatchesByRiotID retrieves match history from opgl-data service using Riot ID
func (proxy *ServiceProxy) GetMatchesByRiotID(region string, gameName string, tagLine string, count int) ([]models.Match, error) {
	url := fmt.Sprintf("%s/api/v1/matches", proxy.dataServiceURL)

	requestBody := map[string]interface{}{
		"region":   region,
		"gameName": gameName,
		"tagLine":  tagLine,
		"count":    count,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	response, err := proxy.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call data service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("data service returned error %d: %s", response.StatusCode, string(body))
	}

	var matches []models.Match
	if err := json.NewDecoder(response.Body).Decode(&matches); err != nil {
		return nil, fmt.Errorf("failed to decode matches response: %w", err)
	}

	return matches, nil
}

// GetMatchesByPUUID retrieves match history from opgl-data service using PUUID (internal use)
func (proxy *ServiceProxy) GetMatchesByPUUID(region string, puuid string, count int) ([]models.Match, error) {
	url := fmt.Sprintf("%s/api/v1/matches", proxy.dataServiceURL)

	requestBody := map[string]interface{}{
		"region": region,
		"puuid":  puuid,
		"count":  count,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	response, err := proxy.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call data service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("data service returned error %d: %s", response.StatusCode, string(body))
	}

	var matches []models.Match
	if err := json.NewDecoder(response.Body).Decode(&matches); err != nil {
		return nil, fmt.Errorf("failed to decode matches response: %w", err)
	}

	return matches, nil
}

// AnalyzePlayer sends analysis request to opgl-cortex-engine
func (proxy *ServiceProxy) AnalyzePlayer(summoner *models.Summoner, matches []models.Match) (*models.AnalysisResult, error) {
	requestBody := map[string]interface{}{
		"summoner": summoner,
		"matches":  matches,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/analyze", proxy.cortexServiceURL)
	response, err := proxy.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call cortex engine: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("cortex engine returned error %d: %s", response.StatusCode, string(body))
	}

	var analysisResult models.AnalysisResult
	if err := json.NewDecoder(response.Body).Decode(&analysisResult); err != nil {
		return nil, fmt.Errorf("failed to decode analysis response: %w", err)
	}

	return &analysisResult, nil
}
