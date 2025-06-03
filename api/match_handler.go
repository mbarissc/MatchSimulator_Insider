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

type MatchHandler struct {
	leagueService abstracts.ILeagueService
}

func NewMatchHandler(ls abstracts.ILeagueService) *MatchHandler {
	return &MatchHandler{
		leagueService: ls,
	}
}

// EditMatchScoreHandler, belirli bir maçın skorunu düzenler.
func (h *MatchHandler) EditMatchScoreHandler(w http.ResponseWriter, r *http.Request) {
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

	if tableErr != nil {
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
