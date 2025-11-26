package proxy

import "github.com/OPGLOL/opgl-gateway-service/internal/models"

// ServiceProxyInterface defines the interface for service proxy operations
// This interface enables mocking in tests
type ServiceProxyInterface interface {
	// GetSummonerByRiotID retrieves summoner data from opgl-data service using Riot ID
	GetSummonerByRiotID(region string, gameName string, tagLine string) (*models.Summoner, error)

	// GetMatchesByRiotID retrieves match history from opgl-data service using Riot ID
	GetMatchesByRiotID(region string, gameName string, tagLine string, count int) ([]models.Match, error)

	// GetMatchesByPUUID retrieves match history from opgl-data service using PUUID
	GetMatchesByPUUID(region string, puuid string, count int) ([]models.Match, error)

	// AnalyzePlayer sends analysis request to opgl-cortex-engine
	AnalyzePlayer(summoner *models.Summoner, matches []models.Match) (*models.AnalysisResult, error)
}
