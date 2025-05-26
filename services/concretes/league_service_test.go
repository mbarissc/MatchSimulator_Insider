package concretes

import (
	"MatchSimulator_Insider/models" // Path to your models package
	"context"
	"errors"  // For creating test errors
	"reflect" // For DeepEqual
	"testing"
	// "fmt" // Not directly used in this version, can be added if detailed error printing is needed
)

// TestUpdateTeamStatsInMemory tests the helper function updateTeamStatsInMemory.
func TestUpdateTeamStatsInMemory(t *testing.T) {
	// initialTeamStats provides a fresh team struct for each test case.
	initialTeamStats := func() models.Team {
		return models.Team{ID: 1, Name: "Test Team", Strength: 80, Played: 0, Wins: 0, Draws: 0, Losses: 0, GoalsFor: 0, GoalsAgainst: 0, GoalDifference: 0, Points: 0}
	}

	testCases := []struct {
		name              string
		initialStats      models.Team
		goalsScored       int
		goalsConceded     int
		expectedTeamStats models.Team
	}{
		{
			name:          "Win Scenario", // Turkish: "Galibiyet Durumu"
			initialStats:  initialTeamStats(),
			goalsScored:   3,
			goalsConceded: 1,
			expectedTeamStats: models.Team{
				ID: 1, Name: "Test Team", Strength: 80,
				Played: 1, Wins: 1, Draws: 0, Losses: 0,
				GoalsFor: 3, GoalsAgainst: 1, GoalDifference: 2, Points: 3,
			},
		},
		{
			name:          "Loss Scenario", // Turkish: "Mağlubiyet Durumu"
			initialStats:  initialTeamStats(),
			goalsScored:   0,
			goalsConceded: 2,
			expectedTeamStats: models.Team{
				ID: 1, Name: "Test Team", Strength: 80,
				Played: 1, Wins: 0, Draws: 0, Losses: 1,
				GoalsFor: 0, GoalsAgainst: 2, GoalDifference: -2, Points: 0,
			},
		},
		{
			name:          "Draw Scenario", // Turkish: "Beraberlik Durumu"
			initialStats:  initialTeamStats(),
			goalsScored:   2,
			goalsConceded: 2,
			expectedTeamStats: models.Team{
				ID: 1, Name: "Test Team", Strength: 80,
				Played: 1, Wins: 0, Draws: 1, Losses: 0,
				GoalsFor: 2, GoalsAgainst: 2, GoalDifference: 0, Points: 1,
			},
		},
		{
			name: "Win on Top of Existing Stats", // Turkish: "Mevcut İstatistikler Üzerine Galibiyet"
			initialStats: models.Team{
				ID: 1, Name: "Test Team", Strength: 80,
				Played: 2, Wins: 1, Draws: 0, Losses: 1,
				GoalsFor: 5, GoalsAgainst: 3, GoalDifference: 2, Points: 3,
			},
			goalsScored:   4,
			goalsConceded: 0,
			expectedTeamStats: models.Team{
				ID: 1, Name: "Test Team", Strength: 80,
				Played: 3, Wins: 2, Draws: 0, Losses: 1,
				GoalsFor: 9, GoalsAgainst: 3, GoalDifference: 6, Points: 6,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			teamToUpdate := tc.initialStats // Create a copy for each test run
			updateTeamStatsInMemory(&teamToUpdate, tc.goalsScored, tc.goalsConceded)

			if !reflect.DeepEqual(teamToUpdate, tc.expectedTeamStats) {
				t.Errorf("Incorrect Team Statistics:\nExpected: %+v\nGot:      %+v (Scenario: %s, Score: %d-%d)",
					tc.expectedTeamStats, teamToUpdate, tc.name, tc.goalsScored, tc.goalsConceded)
			}
		})
	}
}

// --- MockTeamService for testing LeagueService.GetLeagueTable ---
// mockTeamService is a mock implementation of the ITeamService interface.
type mockTeamService struct {
	// GetAllTeamsFunc allows defining a custom function for GetAllTeams for each test case.
	GetAllTeamsFunc func(ctx context.Context) ([]models.Team, error)
	// Other ITeamService methods can be added here if needed for other tests.
}

// CreateTeam is a mock implementation.
func (m *mockTeamService) CreateTeam(ctx context.Context, team models.Team) (int, error) {
	return 0, nil
}

// GetTeamByID is a mock implementation.
func (m *mockTeamService) GetTeamByID(ctx context.Context, id int) (*models.Team, error) {
	return nil, nil
}

// UpdateTeamStatsAfterMatch is a mock implementation.
func (m *mockTeamService) UpdateTeamStatsAfterMatch(ctx context.Context, teamID int, goalsScored int, goalsConceded int) error {
	return nil
}

// ResetAllTeamStats is a mock implementation.
func (m *mockTeamService) ResetAllTeamStats(ctx context.Context) error { return nil }

// AdjustTeamStatsForScoreChange is a mock implementation.
func (m *mockTeamService) AdjustTeamStatsForScoreChange(ctx context.Context, teamID int, oldGS, oldGA, newGS, newGA int) error {
	return nil
}

// UpdateTeamStrength is a mock implementation.
func (m *mockTeamService) UpdateTeamStrength(ctx context.Context, teamID int, newStrength int) error {
	return nil
}

// UpdateTeamName is a mock implementation.
func (m *mockTeamService) UpdateTeamName(ctx context.Context, teamID int, newName string) error {
	return nil
}

// ResetTeamsToDefaults is a mock implementation.
func (m *mockTeamService) ResetTeamsToDefaults(ctx context.Context) error { return nil }

// GetAllTeams provides the mock implementation for ITeamService.GetAllTeams.
// It calls GetAllTeamsFunc if defined, otherwise returns an error.
func (m *mockTeamService) GetAllTeams(ctx context.Context) ([]models.Team, error) {
	if m.GetAllTeamsFunc != nil {
		return m.GetAllTeamsFunc(ctx)
	}
	return nil, errors.New("mockTeamService.GetAllTeamsFunc not defined")
}

// --- MockMatchService (Needed for LeagueService constructor, even if not directly used by GetLeagueTable) ---
// mockMatchService is a mock implementation of the IMatchService interface.
type mockMatchService struct {
	// Add Func fields for IMatchService methods if they need to be mocked in other tests.
}

// Implement IMatchService methods (those not used can return nil or default values).
func (m *mockMatchService) GenerateAndStoreFixture(ctx context.Context, teams []models.Team) error {
	return nil
}
func (m *mockMatchService) GetMatchesByWeek(ctx context.Context, week int) ([]models.Match, error) {
	return nil, nil
}
func (m *mockMatchService) GetMatchByID(ctx context.Context, id int) (*models.Match, error) {
	return nil, nil
}
func (m *mockMatchService) UpdateMatchResult(ctx context.Context, matchID int, homeGoals, awayGoals int, isPlayed bool) error {
	return nil
}
func (m *mockMatchService) GetAllMatches(ctx context.Context) ([]models.Match, error) {
	return nil, nil
}
func (m *mockMatchService) SimulateMatchOutcome(ctx context.Context, homeTeam models.Team, awayTeam models.Team) (int, int, error) {
	return 0, 0, nil
}
func (m *mockMatchService) EditMatchScore(ctx context.Context, matchID int, newHomeGoals int, newAwayGoals int) (models.Match, error) {
	return models.Match{}, nil
}

// TestLeagueService_GetLeagueTable tests the sorting logic of the GetLeagueTable method.
func TestLeagueService_GetLeagueTable(t *testing.T) {
	// Create mock services
	mockTS := &mockTeamService{}
	mockMS := &mockMatchService{} // GetLeagueTable doesn't directly depend on MatchService, but LeagueService constructor needs it.

	// Initialize LeagueService with mock dependencies
	leagueService := NewLeagueService(mockTS, mockMS)

	testCases := []struct {
		name          string
		teamsToReturn []models.Team // Teams that mockTeamService.GetAllTeams will return
		expectedOrder []models.Team // Expected sorted list of teams after GetLeagueTable
		expectedError error
	}{
		{
			name:          "Empty Team List", // Turkish: "Boş Takım Listesi"
			teamsToReturn: []models.Team{},
			expectedOrder: []models.Team{},
			expectedError: nil,
		},
		{
			name: "Already Sorted Team List (Different Points)", // Turkish: "Sıralı Takım Listesi (Puan Farklı)"
			teamsToReturn: []models.Team{
				{ID: 1, Name: "Team A", Points: 10, GoalDifference: 5, GoalsFor: 15},
				{ID: 2, Name: "Team B", Points: 7, GoalDifference: 2, GoalsFor: 10},
			},
			expectedOrder: []models.Team{
				{ID: 1, Name: "Team A", Points: 10, GoalDifference: 5, GoalsFor: 15},
				{ID: 2, Name: "Team B", Points: 7, GoalDifference: 2, GoalsFor: 10},
			},
			expectedError: nil,
		},
		{
			name: "Team List to be Sorted (Equal Points, Different GD)", // Turkish: "Sıralanması Gereken Takım Listesi (Puan Eşit, Averaj Farklı)"
			teamsToReturn: []models.Team{
				{ID: 1, Name: "Team C", Points: 10, GoalDifference: 2, GoalsFor: 12},
				{ID: 2, Name: "Team D", Points: 10, GoalDifference: 5, GoalsFor: 15}, // Better GD
			},
			expectedOrder: []models.Team{
				{ID: 2, Name: "Team D", Points: 10, GoalDifference: 5, GoalsFor: 15},
				{ID: 1, Name: "Team C", Points: 10, GoalDifference: 2, GoalsFor: 12},
			},
			expectedError: nil,
		},
		{
			name: "Team List to be Sorted (Equal Points & GD, Different GF)", // Turkish: "Sıralanması Gereken Takım Listesi (Puan ve Averaj Eşit, Atılan Gol Farklı)"
			teamsToReturn: []models.Team{
				{ID: 1, Name: "Team E", Points: 10, GoalDifference: 5, GoalsFor: 10},
				{ID: 2, Name: "Team F", Points: 10, GoalDifference: 5, GoalsFor: 15}, // More Goals For
			},
			expectedOrder: []models.Team{
				{ID: 2, Name: "Team F", Points: 10, GoalDifference: 5, GoalsFor: 15},
				{ID: 1, Name: "Team E", Points: 10, GoalDifference: 5, GoalsFor: 10},
			},
			expectedError: nil,
		},
		{
			name: "Complex Sorting Scenario", // Turkish: "Karmaşık Sıralama Senaryosu"
			teamsToReturn: []models.Team{
				{ID: 1, Name: "Liverpool", Points: 7, GoalDifference: 2, GoalsFor: 10},
				{ID: 2, Name: "Chelsea", Points: 10, GoalDifference: 5, GoalsFor: 15},
				{ID: 3, Name: "Arsenal", Points: 7, GoalDifference: 2, GoalsFor: 12}, // Same Pts and GD as Liverpool, but more GF
				{ID: 4, Name: "Man City", Points: 10, GoalDifference: 3, GoalsFor: 11},
			},
			expectedOrder: []models.Team{
				{ID: 2, Name: "Chelsea", Points: 10, GoalDifference: 5, GoalsFor: 15},
				{ID: 4, Name: "Man City", Points: 10, GoalDifference: 3, GoalsFor: 11},
				{ID: 3, Name: "Arsenal", Points: 7, GoalDifference: 2, GoalsFor: 12},
				{ID: 1, Name: "Liverpool", Points: 7, GoalDifference: 2, GoalsFor: 10},
			},
			expectedError: nil,
		},
		{
			name:          "Error from TeamService.GetAllTeams", // Turkish: "TeamService'ten Hata Gelmesi"
			teamsToReturn: nil,
			expectedOrder: nil,
			expectedError: errors.New("mock GetAllTeams error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configure mockTeamService.GetAllTeamsFunc for each test case
			mockTS.GetAllTeamsFunc = func(ctx context.Context) ([]models.Team, error) {
				if tc.expectedError != nil && tc.name == "Error from TeamService.GetAllTeams" {
					return nil, tc.expectedError
				}
				// Return a copy to prevent modification of tc.teamsToReturn during the test
				teamsCopy := make([]models.Team, len(tc.teamsToReturn))
				copy(teamsCopy, tc.teamsToReturn)
				return teamsCopy, nil
			}

			actualTable, actualError := leagueService.GetLeagueTable(context.Background())

			// Check for errors
			if tc.expectedError != nil {
				if actualError == nil {
					t.Fatalf("Expected an error (%v) but got nil.", tc.expectedError)
				}
				// Optionally, check if the error message contains the expected error string
				// if !strings.Contains(actualError.Error(), tc.expectedError.Error()) {
				// 	t.Errorf("Expected error message to contain '%v', got '%v'", tc.expectedError, actualError)
				// }
				return // If an error was expected, no need to check the table.
			}
			if actualError != nil {
				t.Fatalf("Did not expect an error but got: %v", actualError)
			}

			// Deep compare the results
			if !reflect.DeepEqual(actualTable, tc.expectedOrder) {
				t.Errorf("Sorting Incorrect:\nExpected Order: %+v\nGot Order:      %+v", tc.expectedOrder, actualTable)
			}
		})
	}
}
