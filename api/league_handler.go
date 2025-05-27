package api

import (
	"MatchSimulator_Insider/models"
	"MatchSimulator_Insider/services/abstracts"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
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

// respondWithError, istemciye bir hata mesajı ve HTTP durum kodu gönderir.
func respondWithError(w http.ResponseWriter, code int, message string) {
	log.Printf("API Response - Error %d: %s", code, message)
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON, istemciye JSON formatında bir cevap gönderir.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("JSON Marshal Error: %v, Payload: %+v", err, payload)
		http.Error(w, "Server Error: Could not format response data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err = w.Write(response); err != nil {
		log.Printf("HTTP response write error: %v", err)
	}
}

// logLeagueTableToConsole, lig tablosunu sunucu konsoluna loglar.
func logLeagueTableToConsole(header string, table []models.Team) {
	log.Printf("\n--- %s (API Handler Log) ---\n", header)
	log.Println(" Rank | Team              | Pld | W | D | L | GF | GA | GD | Pts")
	log.Println("------|-------------------|-----|---|---|---|----|----|----|----")
	for i, t := range table {
		log.Printf(" %-4d | %-17s | %-3d | %-1d | %-1d | %-1d | %-2d | %-2d | %-2d | %-3d\n",
			i+1, t.Name, t.Played, t.Wins, t.Draws, t.Losses, t.GoalsFor, t.GoalsAgainst, t.GoalDifference, t.Points)
	}
	log.Println("-----------------------------------------------------------------------")
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

	// Konsola loglama (isteğe bağlı, bu endpoint zaten tabloyu döndürüyor)
	// logLeagueTableToConsole("League Table requested via GET /league-table", table)

	if len(table) == 0 {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message": "League table is currently empty. Fixture might not be generated or no matches played yet.",
			"table":   []models.Team{},
		})
		return
	}
	respondWithJSON(w, http.StatusOK, table)
}

// PlayNextWeek, bir sonraki haftayı oynatır ve sonuçları döndürür.
// Ayrıca güncel lig tablosunu konsola loglar.
func (h *LeagueHandler) PlayNextWeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Only POST method is supported for this endpoint.")
		return
	}

	ctx := r.Context()
	playedWeek, weekMatches, leagueTable, err := h.leagueService.PlayNextWeek(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "Fikstür eksik") {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		if playedWeek == 0 && (weekMatches == nil || len(weekMatches) == 0) {
			finalTable, tableErr := h.leagueService.GetLeagueTable(ctx)
			if tableErr == nil && finalTable != nil { // Log the final table if available
				logLeagueTableToConsole(fmt.Sprintf("Final League Table (after trying to play next week, league ended)"), finalTable)
			}
			respondWithJSON(w, http.StatusOK, map[string]string{"message": "League already completed or no matches to play."})
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error playing next week: "+err.Error())
		return
	}

	// İşlem başarılıysa lig tablosunu konsola logla
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
// Ayrıca güncel lig tablosunu konsola loglar.
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

	// Lig tablosunu çek ve konsola logla
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
// Bu endpoint çağrıldığında da lig tablosunu loglayabiliriz.
func (h *LeagueHandler) GetPredictions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Only GET method is supported for this endpoint.")
		return
	}
	ctx := r.Context()

	// Lig tablosunu logla (tahmin öncesi durum)
	leagueTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("GetPredictions: Error retrieving league table for logging: %v", tableErr)
	} else if leagueTable != nil {
		logLeagueTableToConsole("League Table when /predictions was called", leagueTable)
	}

	allTeams, err := h.teamService.GetAllTeams(ctx) // Bu, sıralı tabloyu alır
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
	var teamIDs []int // Sıralama için
	for id := range predictionsByID {
		teamIDs = append(teamIDs, id)
	}
	sort.Ints(teamIDs) // ID'ye göre sırala (çıktıda tutarlılık için)

	for _, teamID := range teamIDs {
		prob := predictionsByID[teamID]
		teamName := "Unknown Team"   // Varsayılan
		for _, t := range allTeams { // allTeams zaten sıralı (GetLeagueTable gibi) olmayabilir, GetTeamByID daha iyi olabilir veya allTeams'i map'e çevir
			if t.ID == teamID {
				teamName = t.Name
				break
			}
		}
		displayPredictions = append(displayPredictions, predictionDisplayItem{TeamName: teamName, TeamID: teamID, Probability: prob * 100})
	}
	// Olasılığa göre sırala
	sort.Slice(displayPredictions, func(i, j int) bool {
		return displayPredictions[i].Probability > displayPredictions[j].Probability
	})
	respondWithJSON(w, http.StatusOK, displayPredictions)
}

// ResetLeague, ligdeki tüm ilerlemeyi sıfırlar.
// Ayrıca güncel (sıfırlanmış) lig tablosunu konsola loglar.
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

	// Sıfırlama sonrası lig tablosunu çek ve logla
	leagueTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("ResetLeague: Error retrieving league table for logging after reset: %v", tableErr)
	} else if leagueTable != nil {
		logLeagueTableToConsole("League Table after /reset-league", leagueTable)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "League reset successfully. Team statistics and fixture have been renewed."})
}

// PlayAllRemainingWeeks, ligdeki tüm kalan haftaları oynatır.
// Ayrıca final lig tablosunu konsola loglar.
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

	// İşlem başarılıysa final lig tablosunu konsola logla
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

// EditMatchScoreHandler, belirli bir maçın skorunu düzenler.
// Ayrıca güncel lig tablosunu konsola loglar.
func (h *LeagueHandler) EditMatchScoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Only PUT method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	matchIDStr := r.PathValue("id")
	if matchIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 2 || pathParts[0] != "matches" {
			respondWithError(w, http.StatusBadRequest, "Invalid URL format. Expected /matches/{id}")
			return
		}
		matchIDStr = pathParts[1]
	}
	matchID, err := strconv.Atoi(matchIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid match ID: Must be a number.")
		return
	}
	var reqBody EditMatchScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	defer r.Body.Close()
	if reqBody.HomeGoals < 0 || reqBody.AwayGoals < 0 {
		respondWithError(w, http.StatusBadRequest, "Goal scores cannot be negative.")
		return
	}
	err = h.leagueService.HandleMatchScoreEdit(ctx, matchID, reqBody.HomeGoals, reqBody.AwayGoals)
	if err != nil {
		if strings.Contains(err.Error(), "bulunamadı") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error editing match score: "+err.Error())
		}
		return
	}

	updatedLeagueTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("EditMatchScoreHandler: Match score edited but error retrieving updated league table for logging: %v", tableErr)
	} else if updatedLeagueTable != nil {
		logLeagueTableToConsole(fmt.Sprintf("League Table after editing Match ID %d to %d-%d", matchID, reqBody.HomeGoals, reqBody.AwayGoals), updatedLeagueTable)
	}

	// Yanıt için güncel tabloyu tekrar çekmek yerine yukarıdakini kullanabiliriz.
	if tableErr != nil { // Eğer loglama için tablo çekilirken hata olduysa, yanıtta da belirtelim
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": fmt.Sprintf("Match ID %d score successfully updated to %d-%d. Could not retrieve updated league table for response.", matchID, reqBody.HomeGoals, reqBody.AwayGoals),
		})
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":      fmt.Sprintf("Match ID %d score successfully updated to %d-%d. League table and statistics have been refreshed.", matchID, reqBody.HomeGoals, reqBody.AwayGoals),
		"league_table": updatedLeagueTable,
	})
}

// UpdateTeamStrengthHandler, belirli bir takımın gücünü günceller.
// Ayrıca güncel lig tablosunu konsola loglar.
func (h *LeagueHandler) UpdateTeamStrengthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Only PUT method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	teamIDStr := r.PathValue("id")
	if teamIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 || pathParts[0] != "teams" || pathParts[2] != "strength" {
			respondWithError(w, http.StatusBadRequest, "Invalid URL format. Expected /teams/{id}/strength")
			return
		}
		teamIDStr = pathParts[1]
	}
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid team ID: Must be a number.")
		return
	}
	var reqBody UpdateTeamStrengthRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	defer r.Body.Close()

	err = h.teamService.UpdateTeamStrength(ctx, teamID, reqBody.Strength)
	if err != nil {
		if strings.Contains(err.Error(), "bulunamadı") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "geçersiz güç değeri") {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error updating team strength: "+err.Error())
		}
		return
	}

	// Güncel lig tablosunu logla
	leagueTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("UpdateTeamStrengthHandler: Error retrieving league table for logging: %v", tableErr)
	} else if leagueTable != nil {
		updatedTeamForLog, _ := h.teamService.GetTeamByID(ctx, teamID)
		teamNameToLog := "Unknown"
		if updatedTeamForLog != nil {
			teamNameToLog = updatedTeamForLog.Name
		}
		logLeagueTableToConsole(fmt.Sprintf("League Table after updating strength of Team %s (ID %d) to %d", teamNameToLog, teamID, reqBody.Strength), leagueTable)
	}

	updatedTeam, teamErr := h.teamService.GetTeamByID(ctx, teamID)
	if teamErr != nil {
		log.Printf("UpdateTeamStrengthHandler: Team strength updated but error retrieving updated team info: %v", teamErr)
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": fmt.Sprintf("Team ID %d strength successfully updated to %d. Could not retrieve team details.", teamID, reqBody.Strength),
		})
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Team ID %d strength successfully updated to %d.", teamID, reqBody.Strength),
		"team":    updatedTeam,
	})
}

// UpdateTeamNameHandler, belirli bir takımın ismini günceller.
// Ayrıca güncel lig tablosunu konsola loglar.
func (h *LeagueHandler) UpdateTeamNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Only PUT method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	teamIDStr := r.PathValue("id")
	if teamIDStr == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 || pathParts[0] != "teams" || pathParts[2] != "name" {
			respondWithError(w, http.StatusBadRequest, "Invalid URL format. Expected /teams/{id}/name")
			return
		}
		teamIDStr = pathParts[1]
	}
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid team ID: Must be a number.")
		return
	}
	var reqBody UpdateTeamNameRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	defer r.Body.Close()
	if strings.TrimSpace(reqBody.Name) == "" {
		respondWithError(w, http.StatusBadRequest, "Team name cannot be empty.")
		return
	}

	err = h.teamService.UpdateTeamName(ctx, teamID, reqBody.Name)
	if err != nil {
		if strings.Contains(err.Error(), "bulunamadı") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "zaten kullanımda") || strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusConflict, err.Error())
		} else if strings.Contains(err.Error(), "boş olamaz") {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error updating team name: "+err.Error())
		}
		return
	}

	// Güncel lig tablosunu logla
	leagueTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("UpdateTeamNameHandler: Error retrieving league table for logging: %v", tableErr)
	} else if leagueTable != nil {
		logLeagueTableToConsole(fmt.Sprintf("League Table after updating name of Team ID %d to '%s'", teamID, reqBody.Name), leagueTable)
	}

	updatedTeam, teamErr := h.teamService.GetTeamByID(ctx, teamID)
	if teamErr != nil {
		log.Printf("UpdateTeamNameHandler: Team name updated but error retrieving updated team info: %v", teamErr)
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": fmt.Sprintf("Team ID %d name successfully updated to '%s'. Could not retrieve team details.", teamID, reqBody.Name),
		})
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Team ID %d name successfully updated to '%s'.", teamID, reqBody.Name),
		"team":    updatedTeam,
	})
}

// ResetTeamsToDefaultsHandler, tüm takımların isimlerini ve güçlerini varsayılana sıfırlar.
// Ayrıca tüm lig istatistiklerini ve fikstürü de sıfırlar.
func (h *LeagueHandler) ResetTeamsToDefaultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Only POST method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	err := h.teamService.ResetTeamsToDefaults(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error resetting teams to defaults: "+err.Error())
		return
	}
	err = h.leagueService.ResetLeague(ctx) // Ensure fixture and league stats are also fully reset
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error resetting league after setting teams to defaults: "+err.Error())
		return
	}

	finalTable, tableErr := h.leagueService.GetLeagueTable(ctx)
	if tableErr != nil {
		log.Printf("ResetTeamsToDefaultsHandler: Error retrieving league table for logging after full reset: %v", tableErr)
	} else if finalTable != nil {
		logLeagueTableToConsole("League Table after /teams/reset-defaults", finalTable)
	}

	// Yanıtta da güncel (sıfırlanmış) tabloyu gönderelim.
	if tableErr != nil {
		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": "All teams reset to defaults, league reset. Could not fetch current league table for response.",
		})
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "All teams have been reset to default names and strengths. League statistics and fixture have also been renewed.",
		"league_table": finalTable,
	})
}
