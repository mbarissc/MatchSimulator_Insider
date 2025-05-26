package models

type Match struct {
	ID         int  `json:"id"`
	Week       int  `json:"week"`
	HomeTeamID int  `json:"home_team_id"`
	AwayTeamID int  `json:"away_team_id"`
	HomeGoals  *int `json:"home_goals,omitempty"` // Maç oynanmadıysa null olabilir
	AwayGoals  *int `json:"away_goals,omitempty"` // Maç oynanmadıysa null olabilir
	IsPlayed   bool `json:"is_played"`
}
