package api

import (
	"MatchSimulator_Insider/models"
	"MatchSimulator_Insider/services/abstracts"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

type LeagueHandler struct {
	leagueService abstracts.ILeagueService
	teamService   abstracts.TeamService   
	matchService  abstracts.IMatchService 
}

// NewLeagueHandler, yeni bir LeagueHandler örneği oluşturur.
func NewLeagueHandler(ls abstracts.ILeagueService, ts abstracts.TeamService, ms abstracts.IMatchService) *LeagueHandler {
	return &LeagueHandler{
		leagueService: ls,
		teamService:   ts,
		matchService:  ms,
	}
}

// GetLeagueTable, güncel lig tablosunu JSON olarak döndürür.
func (h *LeagueHandler) GetLeagueTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Only GET method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	table, err := h.leagueService.GetLeagueTable(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving league table: "+err.Error())
		return
	}

	if len(table) == 0 {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message": "League table is currently empty. Fixture might not be generated or no matches played yet.",
			"table":   []models.Team{},
		})
		return
	}
	respondWithJSON(w, http.StatusOK, table)
}

// PlayNextWeek, bir sonraki haftayı oynatır, sonuçları döndürür ve lig tablosunu konsola loglar
func (h *LeagueHandler) PlayNextWeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Only POST method is supported for this endpoint.")
		return
	}

	ctx := r.Context()
	playedWeek, weekMatches, leagueTable, err := h.leagueService.PlayNextWeek(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "Fikstür eksik") { // Varsayım: Hata mesajı bu şekilde olabilir.
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		if playedWeek == 0 && (weekMatches == nil || len(weekMatches) == 0) {
			finalTable, tableErr := h.leagueService.GetLeagueTable(ctx)
			if tableErr == nil && finalTable != nil {
				logLeagueTableToConsole(fmt.Sprintf("Final League Table (after trying to play next week, league ended)"), finalTable)
			}
			respondWithJSON(w, http.StatusOK, map[string]string{"message": "League already completed or no matches to play."})
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error playing next week: "+err.Error())
		return
	}

	if leagueTable != nil {
		logLeagueTableToConsole(fmt.Sprintf("League Table after Week %d played via /next-week", playedWeek), leagueTable)
	}

	response := struct {
		PlayedWeek  int            `json:"played_week"`
		WeekMatches []models.Match `json:"week_matches"`
		LeagueTable []models.Team  `json:"league_table"`
		Message     string         `json:"message,omitempty"`
	}{
		PlayedWeek:  playedWeek,
		WeekMatches: weekMatches,
		LeagueTable: leagueTable,
	}

	if playedWeek == 0 && (weekMatches == nil || len(weekMatches) == 0) {
		response.Message = "League completed. No more weeks to play."
	} else {
		response.Message = fmt.Sprintf("Week %d played successfully.", playedWeek)
	}
	respondWithJSON(w, http.StatusOK, response)
}

// GetCurrentWeekInfo, mevcut oynanacak hafta bilgisini döndürür.
func (h *LeagueHandler) GetCurrentWeekInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Only GET method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	currentWeek, err := h.leagueService.GetCurrentWeek(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving current week information: "+err.Error())
		return
	}

	leagueTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("GetCurrentWeekInfo: Error retrieving league table for logging: %v", tableErr)
	} else if leagueTable != nil {
		logLeagueTableToConsole(fmt.Sprintf("Current League Table when /current-week (next playable: %d) was called", currentWeek), leagueTable)
	}

	var message, leagueStatus string
	if currentWeek == -1 {
		message = "All matches have been played, the league is completed."
		leagueStatus = "Completed"
	} else if currentWeek == 1 {
		allMatches, _ := h.matchService.GetAllMatches(ctx)
		if len(allMatches) == 0 {
			message = "Fixture not yet generated. The league needs a fixture to start."
			leagueStatus = "Not Started (No Fixture)"
		} else {
			message = fmt.Sprintf("Current playable week: %d", currentWeek)
			leagueStatus = "In Progress"
		}
	} else {
		message = fmt.Sprintf("Current playable week: %d", currentWeek)
		leagueStatus = "In Progress"
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"current_playable_week": currentWeek,
		"status_message":        message,
		"league_status":         leagueStatus,
	})
}

// GetPredictions, şampiyonluk tahminlerini döndürür.
func (h *LeagueHandler) GetPredictions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Only GET method is supported for this endpoint.")
		return
	}
	ctx := r.Context()

	leagueTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("GetPredictions: Error retrieving league table for logging: %v", tableErr)
	} else if leagueTable != nil {
		logLeagueTableToConsole("League Table when /predictions was called", leagueTable)
	}
	
	allTeams, err := h.teamService.GetAllTeams(ctx)
	if err != nil || len(allTeams) == 0 {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve team information for predictions or no teams exist.")
		return
	}

	predictionsByID, err := h.leagueService.GetChampionshipPredictions(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "en az 4 hafta tamamlandıktan sonra") {
			respondWithError(w, http.StatusPreconditionFailed, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving championship predictions: "+err.Error())
		return
	}
	type predictionDisplayItem struct {
		TeamName    string  `json:"team_name"`
		TeamID      int     `json:"team_id"`
		Probability float64 `json:"probability_percentage"`
	}
	var displayPredictions []predictionDisplayItem
	var teamIDs []int
	for id := range predictionsByID {
		teamIDs = append(teamIDs, id)
	}
	sort.Ints(teamIDs)

	for _, teamID := range teamIDs {
		prob := predictionsByID[teamID]
		teamName := "Unknown Team"
		for _, t := range allTeams {
			if t.ID == teamID {
				teamName = t.Name
				break
			}
		}
		displayPredictions = append(displayPredictions, predictionDisplayItem{TeamName: teamName, TeamID: teamID, Probability: prob * 100})
	}
	sort.Slice(displayPredictions, func(i, j int) bool {
		return displayPredictions[i].Probability > displayPredictions[j].Probability
	})
	respondWithJSON(w, http.StatusOK, displayPredictions)
}

// ResetLeague, ligdeki tüm ilerlemeyi sıfırlar.
func (h *LeagueHandler) ResetLeague(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Only POST method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	if err := h.leagueService.ResetLeague(ctx); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error resetting league: "+err.Error())
		return
	}

	leagueTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("ResetLeague: Error retrieving league table for logging after reset: %v", tableErr)
	} else if leagueTable != nil {
		logLeagueTableToConsole("League Table after /reset-league", leagueTable)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "League reset successfully. Team statistics and fixture have been renewed."})
}

// PlayAllRemainingWeeks, ligdeki tüm kalan haftaları oynatır.
func (h *LeagueHandler) PlayAllRemainingWeeks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Only POST method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	allPlayedMatches, finalTable, err := h.leagueService.PlayAllRemainingWeeks(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error playing all remaining weeks: "+err.Error())
		return
	}

	if finalTable != nil {
		logLeagueTableToConsole("Final League Table after /play-all", finalTable)
	}

	response := struct {
		Message             string                 `json:"message"`
		PlayedMatchesByWeek map[int][]models.Match `json:"played_matches_by_week"`
		FinalLeagueTable    []models.Team          `json:"final_league_table"`
	}{
		Message:             "All remaining weeks played successfully.",
		PlayedMatchesByWeek: allPlayedMatches,
		FinalLeagueTable:    finalTable,
	}
	if (allPlayedMatches == nil || len(allPlayedMatches) == 0) && finalTable != nil {
		currentWeek, _ := h.leagueService.GetCurrentWeek(ctx) 
		if currentWeek == -1 {
			response.Message = "League was already completed. No additional weeks were played."
		} else if currentWeek == 1 {
			dbMatches, _ := h.matchService.GetAllMatches(ctx) 
			if len(dbMatches) == 0 {
				response.Message = "Fixture not yet generated."
			} else {
				response.Message = "Could not play weeks; league might be at the start or an error occurred."
			}
		}
	}
	respondWithJSON(w, http.StatusOK, response)
}
