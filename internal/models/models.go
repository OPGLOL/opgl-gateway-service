package models

import "time"

// Summoner represents a League of Legends player account (internal use)
type Summoner struct {
	ID            string `json:"id"`
	AccountID     string `json:"accountId"`
	PUUID         string `json:"puuid"`
	Name          string `json:"name"`
	ProfileIconID int    `json:"profileIconId"`
	SummonerLevel int64  `json:"summonerLevel"`
}

// SummonerResponse represents summoner data returned to external clients
// PUUID is excluded for security reasons
type SummonerResponse struct {
	ID            string `json:"id"`
	AccountID     string `json:"accountId"`
	Name          string `json:"name"`
	ProfileIconID int    `json:"profileIconId"`
	SummonerLevel int64  `json:"summonerLevel"`
}

// Match represents a single League of Legends match
type Match struct {
	MatchID      string        `json:"matchId"`
	GameCreation time.Time     `json:"gameCreation"`
	GameDuration int           `json:"gameDuration"`
	GameMode     string        `json:"gameMode"`
	GameType     string        `json:"gameType"`
	Participants []Participant `json:"participants"`
}

// Participant represents a player's performance in a specific match
type Participant struct {
	PUUID                       string `json:"puuid"`
	SummonerName                string `json:"summonerName"`
	ChampionID                  int    `json:"championId"`
	ChampionName                string `json:"championName"`
	Kills                       int    `json:"kills"`
	Deaths                      int    `json:"deaths"`
	Assists                     int    `json:"assists"`
	GoldEarned                  int    `json:"goldEarned"`
	TotalDamageDealtToChampions int    `json:"totalDamageDealtToChampions"`
	TotalDamageTaken            int    `json:"totalDamageTaken"`
	VisionScore                 int    `json:"visionScore"`
	TotalMinionsKilled          int    `json:"totalMinionsKilled"`
	Win                         bool   `json:"win"`
	TeamPosition                string `json:"teamPosition"`
}

// AnalysisResult contains the complete analysis for a player
type AnalysisResult struct {
	PlayerStats      interface{} `json:"playerStats"`
	ImprovementAreas interface{} `json:"improvementAreas"`
	AnalyzedAt       time.Time   `json:"analyzedAt"`
}
