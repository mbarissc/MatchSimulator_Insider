package api

import (
	"MatchSimulator_Insider/services/abstracts"
	"log"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, leagueService abstracts.ILeagueService, teamService abstracts.TeamService, matchService abstracts.IMatchService) {
	log.Println("🛣️  API rotaları kaydediliyor...")

	leagueHandler := NewLeagueHandler(leagueService, teamService, matchService) // ITeamService kullandığımızı varsayıyorum

	// League endpoints
	mux.HandleFunc("GET /league-table", leagueHandler.GetLeagueTable)
	mux.HandleFunc("POST /next-week", leagueHandler.PlayNextWeek)
	mux.HandleFunc("GET /current-week", leagueHandler.GetCurrentWeekInfo)
	mux.HandleFunc("GET /predictions", leagueHandler.GetPredictions)
	mux.HandleFunc("POST /reset-league", leagueHandler.ResetLeague) // Bu, istatistikleri ve fikstürü sıfırlar
	mux.HandleFunc("POST /play-all", leagueHandler.PlayAllRemainingWeeks)

	// Match endpoints
	mux.HandleFunc("PUT /matches/{id}", leagueHandler.EditMatchScoreHandler)

	// Team endpoints
	mux.HandleFunc("PUT /teams/{id}/strength", leagueHandler.UpdateTeamStrengthHandler)
	mux.HandleFunc("PUT /teams/{id}/name", leagueHandler.UpdateTeamNameHandler)             // YENİ
	mux.HandleFunc("POST /teams/reset-defaults", leagueHandler.ResetTeamsToDefaultsHandler) // YENİ

	log.Println("✅ API rotaları başarıyla kaydedildi.")
}
