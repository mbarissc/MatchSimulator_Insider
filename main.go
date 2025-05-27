// FILE: main.go
package main

import (
	"MatchSimulator_Insider/api"
	"MatchSimulator_Insider/config" // config paketini import et
	"MatchSimulator_Insider/models"
	"MatchSimulator_Insider/services/concretes"
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

func printLeagueTableForLog(header string, table []models.Team) {
	log.Printf("\n--- %s ---\n", header)
	log.Println(" Rank | Team              | Pld | W | D | L | GF | GA | GD | Pts")
	log.Println("------|-------------------|-----|---|---|---|----|----|----|----")
	for i, t := range table {
		log.Printf(" %-4d | %-17s | %-3d | %-1d | %-1d | %-1d | %-2d | %-2d | %-2d | %-3d\n",
			i+1, t.Name, t.Played, t.Wins, t.Draws, t.Losses, t.GoalsFor, t.GoalsAgainst, t.GoalDifference, t.Points)
	}
}

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("config.json") // Dosya adını tam olarak belirtiyoruz
	if err != nil {
		log.Printf("WARNING: Error loading configuration from 'config.json': %v. Attempting to use hardcoded defaults.", err)
		// Varsayılan bir config objesi oluştur veya kritik hata ver
		cfg = &config.Config{ // Viper'daki gibi SetDefault mantığını burada uyguluyoruz
			Database: config.DBConfig{
				ConnectionString: "postgres://postgres:1234@localhost:5432/football_league_sim?sslmode=disable",
			},
			Server: config.APIConfig{
				Port: "8080",
			},
		}
	}
	log.Println("Configuration successfully loaded or defaults applied.")

	// 2. Application Startup Settings
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("▶️  Application starting...")

	// 3. Database Connection Setup
	connStr := cfg.Database.ConnectionString
	log.Printf("INFO: Connecting to database using connection string from config...")

	dbConn, errDb := pgx.Connect(context.Background(), connStr)
	if errDb != nil {
		log.Fatalf("Could not connect to the database: %v", errDb)
	}
	defer dbConn.Close(context.Background())

	if errDb = dbConn.Ping(context.Background()); errDb != nil {
		log.Fatalf("Could not ping the database: %v", errDb)
	}
	log.Println("Successfully connected to PostgreSQL database!")

	// 4. Initialization of Services
	teamService := concretes.NewPostgresTeamService(dbConn)
	matchService := concretes.NewPostgresMatchService(dbConn)
	leagueService := concretes.NewLeagueService(teamService, matchService)
	log.Println("INFO: All services successfully created.")

	// 5. League Setup Check (Startup)
	// ... (Bu kısım aynı kalabilir, log mesajları hariç) ...
	log.Println("\n--- League Setup Check (Startup) ---")
	teamsToSeed := []models.Team{
		{Name: "Chelsea", Strength: 85}, {Name: "Arsenal", Strength: 82},
		{Name: "Manchester City", Strength: 90}, {Name: "Liverpool", Strength: 88},
	}
	allCurrentTeams, err := teamService.GetAllTeams(context.Background())
	if err != nil {
		log.Printf("WARNING: Error fetching initial teams: %v. Assuming no teams exist and proceeding with seeding.", err)
		allCurrentTeams = []models.Team{}
	}
	currentTeamCount := len(allCurrentTeams)
	if currentTeamCount < 4 {
		log.Printf("INFO: Teams missing or insufficient in database (%d found), attempting to create/check seed teams...", currentTeamCount)
		for _, teamData := range teamsToSeed {
			createdID, createErr := teamService.CreateTeam(context.Background(), teamData)
			if createErr != nil {
				log.Printf("Error processing team %s during seed: %v", teamData.Name, createErr)
			} else {
				log.Printf("INFO: Create/check operation completed for team %s (ID: %d).", teamData.Name, createdID)
			}
		}
		allCurrentTeams, err = teamService.GetAllTeams(context.Background())
		if err != nil {
			log.Fatalf("Could not fetch teams after seeding attempt: %v", err)
		}
	} else {
		log.Printf("INFO: Sufficient number of teams (%d) already seem to exist in the database.", currentTeamCount)
	}
	if len(allCurrentTeams) < 4 {
		log.Fatalf("Still insufficient teams for setup (%d). At least 4 teams required.", len(allCurrentTeams))
	}
	teamsForFixture := allCurrentTeams[:4]
	if len(allCurrentTeams) > 4 {
		log.Printf("⚠ %d teams found in database, using the first 4 for the fixture.", len(allCurrentTeams))
	}
	existingMatches, err := matchService.GetAllMatches(context.Background())
	if err != nil {
		log.Fatalf("Error checking existing matches: %v", err)
	}
	if len(existingMatches) == 0 {
		log.Println("INFO: No matches found. Initializing new league: resetting team stats and generating fixture...")
		if err := teamService.ResetAllTeamStats(context.Background()); err != nil {
			log.Fatalf("Error resetting team stats for initial fixture generation: %v", err)
		}
		log.Println("Team statistics reset for new league.")
		if errGen := matchService.GenerateAndStoreFixture(context.Background(), teamsForFixture); errGen != nil {
			log.Fatalf("Critical error while creating initial league fixture: %v", errGen)
		}
		log.Println("New league fixture successfully generated.")
	} else {
		log.Println("INFO: Existing matches found. League will attempt to resume. Use /reset-league API for a full manual reset.")
	}
	initialTable, err := leagueService.GetLeagueTable(context.Background())
	if err == nil && initialTable != nil {
		printLeagueTableForLog("League Table at API Startup", initialTable)
	} else if err != nil {
		log.Printf("Error retrieving initial league table: %v", err)
	}

	// 6. Start API Server
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, leagueService, teamService, matchService)

	port := cfg.Server.Port // Yapılandırmadan portu al
	log.Printf("API server starting on http://localhost:%s ...", port)

	errDb = http.ListenAndServe(":"+port, mux)
	if errDb != nil {
		log.Fatalf("Critical error starting API server: %v", errDb)
	}
}
