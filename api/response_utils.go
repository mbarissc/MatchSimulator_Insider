package api

import (
	"MatchSimulator_Insider/models"
	"encoding/json"
	"log"
	"net/http"
)

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
