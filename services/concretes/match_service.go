package concretes

import (
	"MatchSimulator_Insider/models"
	"MatchSimulator_Insider/queries"
	"MatchSimulator_Insider/services/abstracts"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"

	"github.com/jackc/pgx/v5"
)

// PostgresMatchService implements IMatchService using a PostgreSQL database.
type PostgresMatchService struct {
	DB *pgx.Conn
}

// NewPostgresMatchService creates a new instance of PostgresMatchService.
func NewPostgresMatchService(db *pgx.Conn) abstracts.IMatchService {
	return &PostgresMatchService{DB: db}
}

// GenerateAndStoreFixture generates the league fixture for the given teams and stores it in the database.
// It currently supports exactly 4 teams and will delete any existing matches before creating new ones.
func (s *PostgresMatchService) GenerateAndStoreFixture(ctx context.Context, teams []models.Team) error {
	if len(teams) != 4 {
		return fmt.Errorf("PostgresMatchService.GenerateAndStoreFixture: Fixture generation is currently only supported for 4 teams, received: %d", len(teams))
	}
	_, err := s.DB.Exec(ctx, queries.DeleteAllMatchesSQL)
	if err != nil {
		return fmt.Errorf("PostgresMatchService.GenerateAndStoreFixture: Error clearing existing fixture: %w", err)
	}

	var matchesToCreate []models.Match
	teamIDs := make([]int, len(teams))
	for i, t := range teams {
		teamIDs[i] = t.ID
	}

	// Static fixture schedule for 4 teams (6 weeks, 12 matches total)
	schedule := [][][2]int{
		{{teamIDs[0], teamIDs[1]}, {teamIDs[2], teamIDs[3]}}, // Week 1
		{{teamIDs[0], teamIDs[2]}, {teamIDs[1], teamIDs[3]}}, // Week 2
		{{teamIDs[0], teamIDs[3]}, {teamIDs[1], teamIDs[2]}}, // Week 3
		{{teamIDs[1], teamIDs[0]}, {teamIDs[3], teamIDs[2]}}, // Week 4 (Return legs)
		{{teamIDs[2], teamIDs[0]}, {teamIDs[3], teamIDs[1]}}, // Week 5
		{{teamIDs[3], teamIDs[0]}, {teamIDs[2], teamIDs[1]}}, // Week 6
	}

	currentWeek := 1
	for _, weeklyMatches := range schedule {
		for _, matchPair := range weeklyMatches {
			matchesToCreate = append(matchesToCreate, models.Match{
				Week:       currentWeek,
				HomeTeamID: matchPair[0],
				AwayTeamID: matchPair[1],
				IsPlayed:   false,
				HomeGoals:  nil, // Not played yet
				AwayGoals:  nil, // Not played yet
			})
		}
		currentWeek++
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PostgresMatchService.GenerateAndStoreFixture: Could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback in case of error

	for _, match := range matchesToCreate {
		_, err = tx.Exec(ctx, queries.InsertMatchSQL,
			match.Week, match.HomeTeamID, match.AwayTeamID,
			match.IsPlayed, match.HomeGoals, match.AwayGoals,
		)
		if err != nil {
			return fmt.Errorf("PostgresMatchService.GenerateAndStoreFixture: Error adding match for week %d (%d vs %d): %w", match.Week, match.HomeTeamID, match.AwayTeamID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("PostgresMatchService.GenerateAndStoreFixture: Could not commit transaction: %w", err)
	}
	return nil
}

// GetMatchesByWeek retrieves all matches for a specific week, ordered by ID.
func (s *PostgresMatchService) GetMatchesByWeek(ctx context.Context, week int) ([]models.Match, error) {
	rows, err := s.DB.Query(ctx, queries.GetMatchesByWeekSQL, week)
	if err != nil {
		return nil, fmt.Errorf("PostgresMatchService.GetMatchesByWeek: Error retrieving matches for week %d: %w", week, err)
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var match models.Match
		if err := rows.Scan(
			&match.ID, &match.Week, &match.HomeTeamID, &match.AwayTeamID,
			&match.HomeGoals, &match.AwayGoals, &match.IsPlayed,
		); err != nil {
			return nil, fmt.Errorf("PostgresMatchService.GetMatchesByWeek: Error scanning match row: %w", err)
		}
		matches = append(matches, match)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PostgresMatchService.GetMatchesByWeek: Error processing rows: %w", err)
	}
	return matches, nil
}

// GetMatchByID retrieves a specific match by its ID.
func (s *PostgresMatchService) GetMatchByID(ctx context.Context, id int) (*models.Match, error) {
	var match models.Match
	err := s.DB.QueryRow(ctx, queries.GetMatchByIDSQL, id).Scan(
		&match.ID, &match.Week, &match.HomeTeamID, &match.AwayTeamID,
		&match.HomeGoals, &match.AwayGoals, &match.IsPlayed,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("PostgresMatchService.GetMatchByID: Match with ID %d not found", id)
		}
		return nil, fmt.Errorf("PostgresMatchService.GetMatchByID: Error retrieving match (ID: %d): %w", id, err)
	}
	return &match, nil
}

// UpdateMatchResult updates the score and played status of a specific match.
func (s *PostgresMatchService) UpdateMatchResult(ctx context.Context, matchID int, homeGoals, awayGoals int, isPlayed bool) error {
	cmdTag, err := s.DB.Exec(ctx, queries.UpdateMatchResultSQL, homeGoals, awayGoals, isPlayed, matchID)
	if err != nil {
		return fmt.Errorf("PostgresMatchService.UpdateMatchResult: Error updating match result (ID: %d): %w", matchID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("PostgresMatchService.UpdateMatchResult: Match (ID: %d) not found or not updated", matchID)
	}
	return nil
}

// GetAllMatches retrieves all matches from the database, ordered by week and then ID.
func (s *PostgresMatchService) GetAllMatches(ctx context.Context) ([]models.Match, error) {
	rows, err := s.DB.Query(ctx, queries.GetAllMatchesSQL)
	if err != nil {
		return nil, fmt.Errorf("PostgresMatchService.GetAllMatches: Error retrieving all matches: %w", err)
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var match models.Match
		if err := rows.Scan(
			&match.ID, &match.Week, &match.HomeTeamID, &match.AwayTeamID,
			&match.HomeGoals, &match.AwayGoals, &match.IsPlayed,
		); err != nil {
			return nil, fmt.Errorf("PostgresMatchService.GetAllMatches: Error scanning match row: %w", err)
		}
		matches = append(matches, match)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PostgresMatchService.GetAllMatches: Error processing rows: %w", err)
	}
	return matches, nil
}

// SimulateMatchOutcome simulates the score of a match between two teams
func (s *PostgresMatchService) SimulateMatchOutcome(ctx context.Context, homeTeam models.Team, awayTeam models.Team) (homeGoals int, awayGoals int, err error) {
	// Parameters and constants for simulation
	maxPotentialGoals := 6 // Max potential scoring opportunities for each team
	strengthDivisor := 140 // Divisor to adjust probability basada on strength; higher means fewer goals
	homeAdvantage := 10    // Bonus strength for the home team

	effectiveHomeStrength := homeTeam.Strength + homeAdvantage
	// Ensure effective strength is not negative
	if effectiveHomeStrength < 0 {
		effectiveHomeStrength = 0
	}

	effectiveAwayStrength := awayTeam.Strength
	if effectiveAwayStrength < 0 {
		effectiveAwayStrength = 0
	}

	// Calculate goals
	for i := 0; i < maxPotentialGoals; i++ {
		// Chance for home team to score this potential goal
		if rand.Intn(strengthDivisor) < effectiveHomeStrength {
			homeGoals++
		}
		// Chance for away team to score this potential goal
		if rand.Intn(strengthDivisor) < effectiveAwayStrength {
			awayGoals++
		}
	}
	// This simulation currently always succeeds, so no error is returned.
	return homeGoals, awayGoals, nil
}

// EditMatchScore updates the score of a specific match and returns the original match data.
// The match's 'is_played' status is set to true as part of this operation.
func (s *PostgresMatchService) EditMatchScore(ctx context.Context, matchID int, newHomeGoals int, newAwayGoals int) (originalMatch models.Match, err error) {
	log.Printf("PostgresMatchService.EditMatchScore: Initiating score edit for Match ID %d. New score: %d-%d", matchID, newHomeGoals, newAwayGoals)

	// 1. Get the original state of the match (to know the old score)
	originalMatchPtr, err := s.GetMatchByID(ctx, matchID)
	if err != nil {
		return models.Match{}, fmt.Errorf("PostgresMatchService.EditMatchScore: Could not find or retrieve match to edit (ID: %d): %w", matchID, err)
	}
	originalMatch = *originalMatchPtr // Dereference pointer

	// If a score is being edited, it's logical to consider the match as played.
	// The UpdateMatchResult method already handles setting is_played to true.

	// 2. Update the match score and played status using the existing UpdateMatchResult method.
	err = s.UpdateMatchResult(ctx, matchID, newHomeGoals, newAwayGoals, true)
	if err != nil {
		// Return originalMatch data even if update fails, so LeagueService can know what the scores were.
		return originalMatch, fmt.Errorf("PostgresMatchService.EditMatchScore: Error updating score for match (ID: %d): %w", matchID, err)
	}

	log.Printf("PostgresMatchService.EditMatchScore: Score for Match ID %d successfully updated to %d-%d.", matchID, newHomeGoals, newAwayGoals)
	return originalMatch, nil // Return the original (pre-update) match data
}
