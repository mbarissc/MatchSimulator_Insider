package models

// Team, ligdeki bir futbol takımını temsil eder.
type Team struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Strength       int    `json:"strength"` // Takım gücünü temsil eder (örneğin 1-10 arası)
	Played         int    `json:"played"`
	Wins           int    `json:"wins"`
	Draws          int    `json:"draws"`
	Losses         int    `json:"losses"`
	GoalsFor       int    `json:"goals_for"`
	GoalsAgainst   int    `json:"goals_against"`
	GoalDifference int    `json:"goal_difference"`
	Points         int    `json:"points"`
}
