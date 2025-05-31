package concretes

import (
	"MatchSimulator_Insider/models" 
	"MatchSimulator_Insider/queries"
	"MatchSimulator_Insider/services/abstracts"
	"context"
	"errors"
	"fmt"
	"log"
	"strings" 

	"github.com/jackc/pgx/v5"
)

type PostgresTeamService struct {
	DB *pgx.Conn
}


func NewPostgresTeamService(db *pgx.Conn) abstracts.TeamService {
	return &PostgresTeamService{DB: db}
}


func (s *PostgresTeamService) CreateTeam(ctx context.Context, team models.Team) (int, error) {
	var id int
	// Veritabanında team nesnesi ile aynı isimde bir takım olup olmadığını kontrol ediyoruz. Eğer varsa bu takımın id'sini id değişkenine atayacak
	err := s.DB.QueryRow(ctx, queries.CreateTeamCheckExistsSQL, team.Name).Scan(&id)
	// Eğer err = nil ise böyle bir takım bulunmuş demektir
	if err == nil {
		return id, nil
	}
	// Eğer err hatası pgx.ErrNoRows değilse, database'de farklı türden bir hata olmuşmuş demektir
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("PostgresTeamService.CreateTeam: Database error while checking for team '%s': %w", team.Name, err)
	}

	
	err = s.DB.QueryRow(ctx, queries.CreateTeamInsertSQL,
		team.Name, team.Strength,
	).Scan(&id)

	if err != nil {
		
		return 0, fmt.Errorf("PostgresTeamService.CreateTeam: Error adding team '%s': %w", team.Name, err)
	}
	
	return id, nil
}


func (s *PostgresTeamService) GetTeamByID(ctx context.Context, id int) (*models.Team, error) {
	var team models.Team
	
	// Scan komutu ile bütün değişkenler team nesnesine yazılır
	err := s.DB.QueryRow(ctx, queries.GetTeamByIDSQL, id).Scan(
		&team.ID, &team.Name, &team.Strength, &team.Played, &team.Wins, &team.Draws,
		&team.Losses, &team.GoalsFor, &team.GoalsAgainst, &team.GoalDifference, &team.Points,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("PostgresTeamService.GetTeamByID: Team with ID %d not found", id)
		}
		return nil, fmt.Errorf("PostgresTeamService.GetTeamByID: Error retrieving team (ID: %d): %w", id, err)
	}
	return &team, nil
}


func (s *PostgresTeamService) GetAllTeams(ctx context.Context) ([]models.Team, error) {
	rows, err := s.DB.Query(ctx, queries.GetAllTeamsSQL)
	if err != nil {
		return nil, fmt.Errorf("PostgresTeamService.GetAllTeams: Error retrieving teams: %w", err)
	}
	
	// Fonksiyon sonlanmadan hemen önce defer ile rows kapatılır (veritabanı sızıntısı olmaması için)
	defer rows.Close()
	var teams []models.Team
	
	// rows nesnesinin bütün satırları taranır ve teams slice'ına eklenir
	for rows.Next() {
		var team models.Team
		if err := rows.Scan(&team.ID, &team.Name, &team.Strength, &team.Played, &team.Wins, &team.Draws, &team.Losses, &team.GoalsFor, &team.GoalsAgainst, &team.GoalDifference, &team.Points); err != nil {
			return nil, fmt.Errorf("PostgresTeamService.GetAllTeams: Error scanning team row: %w", err)
		}
		teams = append(teams, team)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PostgresTeamService.GetAllTeams: Error processing rows: %w", err)
	}
	return teams, nil
}


func (s *PostgresTeamService) UpdateTeamStatsAfterMatch(ctx context.Context, teamID int, goalsScored int, goalsConceded int) error {
	// Kazanılan puan, kazanma/beraberlik/kaybetme sayısındaki artışlar (0 veya 1 değeri alır)
	var pointsEarned, winIncrement, drawIncrement, lossIncrement int
	if goalsScored > goalsConceded {
		pointsEarned = 3
		winIncrement = 1
	} else if goalsScored < goalsConceded {
		lossIncrement = 1
	} else { 
		pointsEarned = 1
		drawIncrement = 1
	}

	// Bir transaction başlatılır
	tx, err := s.DB.Begin(ctx)
	
	if err != nil {
		return fmt.Errorf("PostgresTeamService.UpdateTeamStatsAfterMatch: Could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // tx.Commit(ctx) başarısız olursa fonksiyon bitiminde transaction'ı geri alır

	// Update main stats (Played, Wins, Draws, Losses, GoalsFor, GoalsAgainst, Points)
	cmdTag, err := tx.Exec(ctx, queries.UpdateTeamMainStatsSQL,
		winIncrement, drawIncrement, lossIncrement,
		goalsScored, goalsConceded, pointsEarned,
		teamID,
	)
	
	if err != nil {
		return fmt.Errorf("PostgresTeamService.UpdateTeamStatsAfterMatch: Error updating main stats for team (ID: %d): %w", teamID, err)
	}
	if cmdTag.RowsAffected() == 0 { // UPDATE komutu hiçbir satırı etkilemediyse o ID'de bir takım yoktur
		return fmt.Errorf("PostgresTeamService.UpdateTeamStatsAfterMatch: Team (ID: %d) not found (main update)", teamID)
	}

	// Gol averajı güncellenir
	_, err = tx.Exec(ctx, queries.UpdateTeamGDSQL, teamID)
	if err != nil {
		return fmt.Errorf("PostgresTeamService.UpdateTeamStatsAfterMatch: Error updating goal difference for team (ID: %d): %w", teamID, err)
	}

	return tx.Commit(ctx)
}


// Name ve Strength dışında bütün takım istatistiklerini sıfırlar
func (s *PostgresTeamService) ResetAllTeamStats(ctx context.Context) error {
	cmdTag, err := s.DB.Exec(ctx, queries.ResetAllTeamStatsSQL)
	if err != nil {
		log.Printf("!!! PostgresTeamService.ResetAllTeamStats DB.Exec ERROR: %v", err)
		return fmt.Errorf("PostgresTeamService.ResetAllTeamStats: Error resetting team statistics: %w", err)
	}
	log.Printf("Team statistics reset. Rows affected: %d", cmdTag.RowsAffected())
	if cmdTag.RowsAffected() < 1 { 
		log.Println("Warning: ResetAllTeamStats affected fewer rows than expected or none at all. Team table might be empty or there's an issue.")
	}
	return nil
}


func calculateOutcomeMetrics(goalsFor, goalsAgainst int) (points, wins, draws, losses int) {
	if goalsFor > goalsAgainst {
		points = 3
		wins = 1
	} else if goalsFor < goalsAgainst {
		losses = 1
	} else { 
		points = 1
		draws = 1
	}
	return
}


func (s *PostgresTeamService) AdjustTeamStatsForScoreChange(ctx context.Context, teamID int, oldGoalsForTeam, oldGoalsAgainstTeam, newGoalsForTeam, newGoalsAgainstTeam int) error {
	log.Printf("PostgresTeamService.AdjustTeamStatsForScoreChange: TeamID: %d, OldScore: %d-%d, NewScore: %d-%d\n",
		teamID, oldGoalsForTeam, oldGoalsAgainstTeam, newGoalsForTeam, newGoalsAgainstTeam)

	oldPoints, oldWins, oldDraws, oldLosses := calculateOutcomeMetrics(oldGoalsForTeam, oldGoalsAgainstTeam)
	newPoints, newWins, newDraws, newLosses := calculateOutcomeMetrics(newGoalsForTeam, newGoalsAgainstTeam)

	deltaWins := newWins - oldWins
	deltaDraws := newDraws - oldDraws
	deltaLosses := newLosses - oldLosses
	deltaPoints := newPoints - oldPoints
	deltaGoalsFor := newGoalsForTeam - oldGoalsForTeam  // Atılan gol farkı
	deltaGoalsAgainst := newGoalsAgainstTeam - oldGoalsAgainstTeam  // Yenilen gol farkı
	

	log.Printf("  Calculated Deltas: dPts:%d, dW:%d, dD:%d, dL:%d, dGF:%d, dGA:%d\n",
		deltaPoints, deltaWins, deltaDraws, deltaLosses, deltaGoalsFor, deltaGoalsAgainst)

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("AdjustTeamStatsForScoreChange: Could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Ana istatistikler güncellenir
	cmdTag, err := tx.Exec(ctx, queries.AdjustTeamStatsSQL,
		deltaWins, deltaDraws, deltaLosses,
		deltaGoalsFor, deltaGoalsAgainst, deltaPoints,
		teamID,
	)
	if err != nil {
		return fmt.Errorf("AdjustTeamStatsForScoreChange: Error updating main stats for team (ID: %d): %w", teamID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("AdjustTeamStatsForScoreChange: Team (ID: %d) not found (main update)", teamID)
	}

	// Averaj yeniden hesaplanır
	_, err = tx.Exec(ctx, queries.UpdateTeamGDSQL, teamID)
	if err != nil {
		return fmt.Errorf("AdjustTeamStatsForScoreChange: Error updating goal difference for team (ID: %d): %w", teamID, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("AdjustTeamStatsForScoreChange: Could not commit transaction: %w", err)
	}

	log.Printf("  Team (ID: %d) statistics successfully adjusted for score change.\n", teamID)
	return nil
}


func (s *PostgresTeamService) UpdateTeamStrength(ctx context.Context, teamID int, newStrength int) error {
	// Strength 1 ile 100 arasında bir değer almalıdır
	if newStrength < 1 || newStrength > 100 { 
		return fmt.Errorf("invalid strength value: %d. Strength must be between 1 and 100", newStrength)
	}

	cmdTag, err := s.DB.Exec(ctx, queries.UpdateTeamStrengthSQL, newStrength, teamID)
	if err != nil {
		return fmt.Errorf("PostgresTeamService.UpdateTeamStrength: Error updating strength for team (ID: %d): %w", teamID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("PostgresTeamService.UpdateTeamStrength: Team (ID: %d) not found or strength not updated", teamID)
	}
	log.Printf("Team (ID: %d) strength successfully updated to %d.", teamID, newStrength)
	return nil
}



func (s *PostgresTeamService) UpdateTeamName(ctx context.Context, teamID int, newName string) error {
	// Takım isminin başındaki ve sonundaki boşluk karakterleri temizlenir
	trimmedName := strings.TrimSpace(newName)
	if trimmedName == "" {
		return fmt.Errorf("team name cannot be empty")
	}

	// Check if the new name is already used by another team
	var existingID int
	err := s.DB.QueryRow(ctx, queries.CreateTeamCheckExistsSQL, trimmedName).Scan(&existingID)
	if err == nil && existingID != teamID { // Name found and it belongs to a different team
		return fmt.Errorf("name '%s' is already in use by another team (ID: %d)", trimmedName, existingID)
	}
	// If err is not pgx.ErrNoRows, it's an unexpected DB error during the check
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("PostgresTeamService.UpdateTeamName: Error checking new name '%s': %w", trimmedName, err)
	}

	// Proceed to update the name
	cmdTag, err := s.DB.Exec(ctx, queries.UpdateTeamNameSQL, trimmedName, teamID)
	if err != nil {
		// Catch unique constraint violation specifically if the above check missed a race condition or if the DB enforces it differently.
		if strings.Contains(err.Error(), "violates unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("name '%s' is already in use or another unique constraint was violated", trimmedName)
		}
		return fmt.Errorf("PostgresTeamService.UpdateTeamName: Error updating name for team (ID: %d) to '%s': %w", teamID, trimmedName, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("PostgresTeamService.UpdateTeamName: Team (ID: %d) not found or name not updated", teamID)
	}
	log.Printf("Team (ID: %d) name successfully updated to '%s'.", teamID, trimmedName)
	return nil
}

// ResetTeamsToDefaults resets all teams to their default names and strengths,
// and also resets their league statistics.
func (s *PostgresTeamService) ResetTeamsToDefaults(ctx context.Context) error {
	log.Println("--- PostgresTeamService.ResetTeamsToDefaults STARTED ---")

	defaultTeams := []models.Team{
		{Name: "Chelsea", Strength: 85}, // Default names and strengths
		{Name: "Arsenal", Strength: 82},
		{Name: "Manchester City", Strength: 90},
		{Name: "Liverpool", Strength: 88},
	}

	// 1. Reset all league statistics for all teams first.
	if err := s.ResetAllTeamStats(ctx); err != nil {
		// The error from ResetAllTeamStats already includes context.
		return fmt.Errorf("ResetTeamsToDefaults: Error while resetting team statistics: %w", err)
	}

	// 2. Get current teams (now with zeroed stats), ordered by ID for consistent updates.
	currentTeams, err := s.getAllTeamsOrderedByID(ctx)
	if err != nil {
		return fmt.Errorf("ResetTeamsToDefaults: Error fetching current teams after stat reset: %w", err)
	}

	// We expect 4 teams to be present. If not, the behavior might be undefined or error out.
	// This function assumes 4 teams exist and will be updated.
	if len(currentTeams) < len(defaultTeams) {
		// This scenario should ideally be handled by ensuring 4 teams always exist,
		// or this function could also create missing teams up to the default count.
		// For now, log a warning and proceed if possible, or return an error.
		log.Printf("ResetTeamsToDefaults: Warning! Database has %d teams, but %d default configurations exist. Will update available teams.", len(currentTeams), len(defaultTeams))
		if len(currentTeams) == 0 && len(defaultTeams) > 0 {
			return fmt.Errorf("ResetTeamsToDefaults: No teams in database to reset to defaults. Please ensure teams are seeded first.")
		}
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ResetTeamsToDefaults: Could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update existing teams (up to the number of default teams) with default names and strengths
	for i, defaultTeam := range defaultTeams {
		if i < len(currentTeams) {
			teamToUpdate := currentTeams[i] // Update teams in their current ID order
			log.Printf("  Resetting to default: Team ID %d -> New Name: %s, New Strength: %d", teamToUpdate.ID, defaultTeam.Name, defaultTeam.Strength)

			// Check if new default name conflicts with an existing name (other than the current team being updated if names are shuffled)
			var conflictingID int
			errNameCheck := tx.QueryRow(ctx, queries.CreateTeamCheckExistsSQL, defaultTeam.Name).Scan(&conflictingID)
			if errNameCheck == nil && conflictingID != teamToUpdate.ID {
				
				return fmt.Errorf("ResetTeamsToDefaults: Default name '%s' for team ID %d is already in use by team ID %d", defaultTeam.Name, teamToUpdate.ID, conflictingID)
			}
			if errNameCheck != nil && !errors.Is(errNameCheck, pgx.ErrNoRows) {
				return fmt.Errorf("ResetTeamsToDefaults: Error checking name conflict for '%s': %w", defaultTeam.Name, errNameCheck)
			}

			cmdTag, err := tx.Exec(ctx, queries.UpdateTeamNameAndStrengthSQL, defaultTeam.Name, defaultTeam.Strength, teamToUpdate.ID)
			if err != nil {
				return fmt.Errorf("ResetTeamsToDefaults: Error updating name/strength for team (ID: %d): %w", teamToUpdate.ID, err)
			}
			if cmdTag.RowsAffected() == 0 {
				// This shouldn't happen if currentTeams[i] exists.
				log.Printf("ResetTeamsToDefaults: Name/strength update for team (ID: %d) affected 0 rows.", teamToUpdate.ID)
			}
		} else {
			log.Printf("ResetTeamsToDefaults: Warning: No existing team in DB at index %d to map to default team '%s'.", i, defaultTeam.Name)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ResetTeamsToDefaults: Could not commit transaction: %w", err)
	}

	log.Println("PostgresTeamService.ResetTeamsToDefaults: Team names and strengths (and stats) reset to defaults.")
	log.Println("--- PostgresTeamService.ResetTeamsToDefaults FINISHED ---")
	return nil
}

// getAllTeamsOrderedByID is an unexported helper method to retrieve all teams ordered by their ID.
func (s *PostgresTeamService) getAllTeamsOrderedByID(ctx context.Context) ([]models.Team, error) {
	rows, err := s.DB.Query(ctx, queries.GetAllTeamsOrderedByIDSQL)
	if err != nil {
		return nil, fmt.Errorf("PostgresTeamService.getAllTeamsOrderedByID: Error retrieving teams ordered by ID: %w", err)
	}
	defer rows.Close()
	var teams []models.Team
	for rows.Next() {
		var team models.Team
		if err := rows.Scan(&team.ID, &team.Name, &team.Strength, &team.Played, &team.Wins, &team.Draws, &team.Losses, &team.GoalsFor, &team.GoalsAgainst, &team.GoalDifference, &team.Points); err != nil {
			return nil, fmt.Errorf("PostgresTeamService.getAllTeamsOrderedByID: Error scanning team row: %w", err)
		}
		teams = append(teams, team)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PostgresTeamService.getAllTeamsOrderedByID: Error processing rows: %w", err)
	}
	return teams, nil
}
