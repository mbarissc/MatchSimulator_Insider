package concretes

import (
	"MatchSimulator_Insider/models"
	"context"
	"testing"

	"github.com/jackc/pgx/v5" // Can remain because NewPostgresMatchService takes *pgx.Conn
)

// TestPostgresMatchService_SimulateMatchOutcome tests the match simulation logic for various scenarios.
func TestPostgresMatchService_SimulateMatchOutcome(t *testing.T) {
	// Since SimulateMatchOutcome doesn't use the DB, a dummy connection can be used to create the service.
	// If NewPostgresMatchService fails with a nil DB, this part should be adjusted.
	var dummyDbConn *pgx.Conn
	service := NewPostgresMatchService(dummyDbConn)

	// The value of maxPotentialGoals was defined as 6 in the SimulateMatchOutcome implementation.
	// This value is the theoretical maximum number of goals a team can score in a match.
	const maxGoalsImplemented = 6

	testCases := []struct {
		name     string
		homeTeam models.Team
		awayTeam models.Team
	}{
		{
			name:     "Equal Medium Strengths",
			homeTeam: models.Team{ID: 1, Name: "Home Team Medium", Strength: 50},
			awayTeam: models.Team{ID: 2, Name: "Away Team Medium", Strength: 50},
		},
		{
			name:     "Equal High Strengths",
			homeTeam: models.Team{ID: 3, Name: "Home Team High", Strength: 90},
			awayTeam: models.Team{ID: 4, Name: "Away Team High", Strength: 90},
		},
		{
			name:     "Equal Low Strengths",
			homeTeam: models.Team{ID: 5, Name: "Home Team Low", Strength: 10},
			awayTeam: models.Team{ID: 6, Name: "Away Team Low", Strength: 10},
		},
		{
			name:     "Strong Home vs Weak Away",
			homeTeam: models.Team{ID: 7, Name: "Strong Home", Strength: 95},
			awayTeam: models.Team{ID: 8, Name: "Weak Away", Strength: 20},
		},
		{
			name:     "Weak Home vs Strong Away",
			homeTeam: models.Team{ID: 9, Name: "Weak Home", Strength: 20},
			awayTeam: models.Team{ID: 10, Name: "Strong Away", Strength: 95},
		},
		{
			name:     "Extreme Strength Difference (Home Max, Away Min)",
			homeTeam: models.Team{ID: 11, Name: "Max Strength Home", Strength: 100},
			awayTeam: models.Team{ID: 12, Name: "Min Strength Away", Strength: 1},
		},
		{
			name:     "Extreme Strength Difference (Home Min, Away Max)",
			homeTeam: models.Team{ID: 13, Name: "Min Strength Home", Strength: 1},
			awayTeam: models.Team{ID: 14, Name: "Max Strength Away", Strength: 100},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run each scenario multiple times to observe that randomness produces different results,
			// but the core validations remain the same.
			for i := 0; i < 5; i++ { // Run 5 simulations for each scenario
				homeGoals, awayGoals, err := service.SimulateMatchOutcome(context.Background(), tc.homeTeam, tc.awayTeam)

				if err != nil {
					t.Errorf("Test Case: %s (Iteration %d) - SimulateMatchOutcome returned an error: %v", tc.name, i, err)
					continue // If there's an error, skip the other checks
				}

				if homeGoals < 0 || awayGoals < 0 {
					t.Errorf("Test Case: %s (Iteration %d) - Simulated goals cannot be negative. Got Home: %d, Away: %d", tc.name, i, homeGoals, awayGoals)
				}

				// Check against maxPotentialGoals defined in SimulateMatchOutcome
				if homeGoals > maxGoalsImplemented {
					t.Errorf("Test Case: %s (Iteration %d) - Home goals %d exceeded maximum expected %d.", tc.name, i, homeGoals, maxGoalsImplemented)
				}
				if awayGoals > maxGoalsImplemented {
					t.Errorf("Test Case: %s (Iteration %d) - Away goals %d exceeded maximum expected %d.", tc.name, i, awayGoals, maxGoalsImplemented)
				}
			}
		})
	}
}
