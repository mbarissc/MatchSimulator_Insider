package models

type Match struct {
	ID         int  `json:"id"`
	Week       int  `json:"week"`
	HomeTeamID int  `json:"home_team_id"`
	AwayTeamID int  `json:"away_team_id"`
	HomeGoals  *int `json:"home_goals,omitempty"` 
	AwayGoals  *int `json:"away_goals,omitempty"` 
	IsPlayed   bool `json:"is_played"`
}
