package api

// EditMatchScoreRequest, maç skoru düzenleme isteğinin gövdesini tanımlar.
type EditMatchScoreRequest struct {
	HomeGoals int `json:"home_goals"`
	AwayGoals int `json:"away_goals"`
}

// UpdateTeamStrengthRequest, takım gücü güncelleme isteğinin gövdesini tanımlar.
type UpdateTeamStrengthRequest struct {
	Strength int `json:"strength"`
}

// UpdateTeamNameRequest, takım ismi güncelleme isteğinin gövdesini tanımlar.
type UpdateTeamNameRequest struct {
	Name string `json:"name"`
}
