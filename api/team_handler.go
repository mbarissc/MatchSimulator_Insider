package api

import (
	"MatchSimulator_Insider/services/abstracts"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type TeamHandler struct {
	teamService   abstracts.TeamService
	leagueService abstracts.ILeagueService
}

func NewTeamHandler(ts abstracts.TeamService, ls abstracts.ILeagueService) *TeamHandler {
	return &TeamHandler{
		teamService:   ts,
		leagueService: ls,
	}
}

// UpdateTeamStrengthHandler, belirli bir takımın gücünü günceller.
func (h *TeamHandler) UpdateTeamStrengthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Only PUT method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	teamIDStr := r.PathValue("id")
	if teamIDStr == "" { // Fallback for older Go versions or different routers
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
		if strings.Contains(err.Error(), "bulunamadı") { // Varsayım
			respondWithError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "geçersiz güç değeri") { // Varsayım
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error updating team strength: "+err.Error())
		}
		return
	}

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
func (h *TeamHandler) UpdateTeamNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Only PUT method is supported for this endpoint.")
		return
	}
	ctx := r.Context()
	teamIDStr := r.PathValue("id")
	if teamIDStr == "" { // Fallback
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
		} else if strings.Contains(err.Error(), "zaten kullanımda") || strings.Contains(err.Error(), "unique constraint") { // Varsayım
			respondWithError(w, http.StatusConflict, err.Error())
		} else if strings.Contains(err.Error(), "boş olamaz") {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error updating team name: "+err.Error())
		}
		return
	}

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
func (h *TeamHandler) ResetTeamsToDefaultsHandler(w http.ResponseWriter, r *http.Request) {
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
	err = h.leagueService.ResetLeague(ctx)
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
