package api

import (
	"encoding/json"
	"net/http"

	"github.com/OPGLOL/opgl-gateway-service/internal/proxy"
)

// Handler manages HTTP request handlers for the gateway
type Handler struct {
	serviceProxy proxy.ServiceProxyInterface
}

// NewHandler creates a new Handler instance
func NewHandler(serviceProxy proxy.ServiceProxyInterface) *Handler {
	return &Handler{
		serviceProxy: serviceProxy,
	}
}

// HealthCheck handles health check requests
func (handler *Handler) HealthCheck(writer http.ResponseWriter, request *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "opgl-gateway",
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(response)
}

// GetSummoner proxies summoner requests to opgl-data service using Riot ID
func (handler *Handler) GetSummoner(writer http.ResponseWriter, request *http.Request) {
	var summonerRequest struct {
		Region   string `json:"region"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
	}

	if err := json.NewDecoder(request.Body).Decode(&summonerRequest); err != nil {
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	if summonerRequest.Region == "" || summonerRequest.GameName == "" || summonerRequest.TagLine == "" {
		http.Error(writer, "region, gameName, and tagLine are required", http.StatusBadRequest)
		return
	}

	summoner, err := handler.serviceProxy.GetSummonerByRiotID(summonerRequest.Region, summonerRequest.GameName, summonerRequest.TagLine)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(summoner)
}

// GetMatches proxies match history requests to opgl-data service using Riot ID
func (handler *Handler) GetMatches(writer http.ResponseWriter, request *http.Request) {
	var matchRequest struct {
		Region   string `json:"region"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
		Count    int    `json:"count"`
	}

	if err := json.NewDecoder(request.Body).Decode(&matchRequest); err != nil {
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	if matchRequest.Region == "" || matchRequest.GameName == "" || matchRequest.TagLine == "" {
		http.Error(writer, "region, gameName, and tagLine are required", http.StatusBadRequest)
		return
	}

	count := matchRequest.Count
	if count <= 0 {
		count = 20
	}

	matches, err := handler.serviceProxy.GetMatchesByRiotID(matchRequest.Region, matchRequest.GameName, matchRequest.TagLine, count)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(matches)
}

// AnalyzePlayer orchestrates player analysis by calling both data and cortex services using Riot ID
func (handler *Handler) AnalyzePlayer(writer http.ResponseWriter, request *http.Request) {
	var analyzeRequest struct {
		Region   string `json:"region"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
	}

	if err := json.NewDecoder(request.Body).Decode(&analyzeRequest); err != nil {
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	if analyzeRequest.Region == "" || analyzeRequest.GameName == "" || analyzeRequest.TagLine == "" {
		http.Error(writer, "region, gameName, and tagLine are required", http.StatusBadRequest)
		return
	}

	// Step 1: Get summoner data from opgl-data
	summoner, err := handler.serviceProxy.GetSummonerByRiotID(analyzeRequest.Region, analyzeRequest.GameName, analyzeRequest.TagLine)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 2: Get match history from opgl-data (using internal method with PUUID)
	matches, err := handler.serviceProxy.GetMatchesByPUUID(analyzeRequest.Region, summoner.PUUID, 20)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 3: Send data to opgl-cortex-engine for analysis
	analysisResult, err := handler.serviceProxy.AnalyzePlayer(summoner, matches)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(analysisResult)
}
