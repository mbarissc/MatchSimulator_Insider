package concretes

import (
	"MatchSimulator_Insider/models"
	"MatchSimulator_Insider/services/abstracts"
	"context"
	"fmt"
	"log"
	"sort"
)

// LeagueService manages the overall league progression, simulations, and state.
type LeagueService struct {
	teamService  abstracts.TeamService
	matchService abstracts.IMatchService
}

// NewLeagueService creates a new instance of LeagueService.
func NewLeagueService(ts abstracts.TeamService, ms abstracts.IMatchService) abstracts.ILeagueService {
	return &LeagueService{
		teamService:  ts,
		matchService: ms,
	}
}

// GetCurrentWeek determines the earliest unplayed week in the league.
// Returns -1 if all matches are played, or 1 if no fixture exists.
func (s *LeagueService) GetCurrentWeek(ctx context.Context) (int, error) {
	allMatches, err := s.matchService.GetAllMatches(ctx)
	if err != nil {
		return 0, fmt.Errorf("LeagueService.GetCurrentWeek: Error retrieving matches: %w", err)
	}
	
	if len(allMatches) == 0 {
		return 1, nil
	}
	minUnplayedWeek := -1
	allMatchesEverPlayed := true
	matchesByWeek := make(map[int][]models.Match)
	for _, match := range allMatches {
		matchesByWeek[match.Week] = append(matchesByWeek[match.Week], match)
	}
	var weeks []int
	for k := range matchesByWeek {
		weeks = append(weeks, k)
	}
	sort.Ints(weeks)
	for _, weekNum := range weeks {
		allMatchesInThisWeekPlayed := true
		for _, match := range matchesByWeek[weekNum] {
			if !match.IsPlayed {
				allMatchesInThisWeekPlayed = false
				allMatchesEverPlayed = false
				break
			}
		}
		if !allMatchesInThisWeekPlayed {
			minUnplayedWeek = weekNum
			break
		}
	}
	if allMatchesEverPlayed {
		return -1, nil
	}
	if minUnplayedWeek == -1 && !allMatchesEverPlayed {
		if len(weeks) > 0 {
			return weeks[0], fmt.Errorf("LeagueService.GetCurrentWeek: Inconsistency! No unplayed week found, but the league is not finished. First week in data: %d", weeks[0])
		}
		return 1, fmt.Errorf("LeagueService.GetCurrentWeek: Critical error! Could not determine unplayed week and no week data exists, though matches were found.")
	}
	return minUnplayedWeek, nil
}

// PlayNextWeek simulates the next unplayed week, updates stats, and returns results.
// Returns playedWeekNum=0 if the league is finished.
func (s *LeagueService) PlayNextWeek(ctx context.Context) (playedWeekNum int, weekMatches []models.Match, leagueTable []models.Team, err error) {
	currentWeek, err := s.GetCurrentWeek(ctx)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("LeagueService.PlayNextWeek: Error determining week to play: %w", err)
	}
	// rest of the logic is the same
	if currentWeek == -1 { // Lig bitmiş
		finalTable, errTable := s.GetLeagueTable(ctx)
		return 0, nil, finalTable, errTable
	}

	matchesForThisWeek, err := s.matchService.GetMatchesByWeek(ctx, currentWeek)
	if err != nil {
		return currentWeek, nil, nil, fmt.Errorf("LeagueService.PlayNextWeek: Error retrieving matches for week %d: %w", currentWeek, err)
	}

	if len(matchesForThisWeek) == 0 {
		currentTable, tableErr := s.GetLeagueTable(ctx)
		if tableErr != nil {
			log.Printf("LeagueService.PlayNextWeek: Additionally, error getting league table: %v", tableErr)
		}
		return currentWeek, nil, currentTable, fmt.Errorf("LeagueService.PlayNextWeek: ℹ️ No matches found for week %d. Fixture might be missing or incomplete", currentWeek)
	}

	playedMatchesResult := make([]models.Match, 0, len(matchesForThisWeek))
	for _, matchToPlay := range matchesForThisWeek {
		if matchToPlay.IsPlayed {
			updatedMatch, _ := s.matchService.GetMatchByID(ctx, matchToPlay.ID)
			if updatedMatch != nil {
				playedMatchesResult = append(playedMatchesResult, *updatedMatch)
			} else {
				playedMatchesResult = append(playedMatchesResult, matchToPlay)
			}
			continue
		}

		homeTeam, errHT := s.teamService.GetTeamByID(ctx, matchToPlay.HomeTeamID)
		if errHT != nil {
			return currentWeek, nil, nil, fmt.Errorf("LeagueService.PlayNextWeek: Could not retrieve home team (ID: %d) info: %w", matchToPlay.HomeTeamID, errHT)
		}
		awayTeam, errAT := s.teamService.GetTeamByID(ctx, matchToPlay.AwayTeamID)
		if errAT != nil {
			return currentWeek, nil, nil, fmt.Errorf("LeagueService.PlayNextWeek: Could not retrieve away team (ID: %d) info: %w", matchToPlay.AwayTeamID, errAT)
		}

		homeGoals, awayGoals, simErr := s.matchService.SimulateMatchOutcome(ctx, *homeTeam, *awayTeam)
		if simErr != nil {
			return currentWeek, nil, nil, fmt.Errorf("LeagueService.PlayNextWeek: Error in match (ID: %d) simulation: %w", matchToPlay.ID, simErr)
		}

		errUpdate := s.matchService.UpdateMatchResult(ctx, matchToPlay.ID, homeGoals, awayGoals, true)
		if errUpdate != nil {
			return currentWeek, nil, nil, fmt.Errorf("LeagueService.PlayNextWeek: Error updating match (ID: %d) result: %w", matchToPlay.ID, errUpdate)
		}

		errHTStats := s.teamService.UpdateTeamStatsAfterMatch(ctx, homeTeam.ID, homeGoals, awayGoals)
		if errHTStats != nil {
			return currentWeek, nil, nil, fmt.Errorf("LeagueService.PlayNextWeek: Error updating stats for home team (%s): %w", homeTeam.Name, errHTStats)
		}
		errATStats := s.teamService.UpdateTeamStatsAfterMatch(ctx, awayTeam.ID, awayGoals, homeGoals)
		if errATStats != nil {
			return currentWeek, nil, nil, fmt.Errorf("LeagueService.PlayNextWeek: Error updating stats for away team (%s): %w", awayTeam.Name, errATStats)
		}

		updatedMatch, errGetMatch := s.matchService.GetMatchByID(ctx, matchToPlay.ID)
		if errGetMatch != nil {
			log.Printf("LeagueService.PlayNextWeek: Warning! Error retrieving updated match info (ID: %d) after playing: %v. Using simulated scores.", matchToPlay.ID, errGetMatch)
			matchToPlay.HomeGoals = &homeGoals
			matchToPlay.AwayGoals = &awayGoals
			matchToPlay.IsPlayed = true
			playedMatchesResult = append(playedMatchesResult, matchToPlay)
		} else if updatedMatch != nil {
			playedMatchesResult = append(playedMatchesResult, *updatedMatch)
		}
	}

	finalLeagueTable, errTable := s.GetLeagueTable(ctx)
	if errTable != nil {
		return currentWeek, playedMatchesResult, nil, fmt.Errorf("LeagueService.PlayNextWeek: Error retrieving league table after playing week: %w", errTable)
	}
	return currentWeek, playedMatchesResult, finalLeagueTable, nil
}

// GetLeagueTable retrieves and sorts all teams to represent the current league standings.
func (s *LeagueService) GetLeagueTable(ctx context.Context) ([]models.Team, error) {
	// rest of the logic is the same
	teams, err := s.teamService.GetAllTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("LeagueService.GetLeagueTable: Could not retrieve teams for league table: %w", err)
	}

	sort.SliceStable(teams, func(i, j int) bool {
		if teams[i].Points != teams[j].Points {
			return teams[i].Points > teams[j].Points
		}
		if teams[i].GoalDifference != teams[j].GoalDifference {
			return teams[i].GoalDifference > teams[j].GoalDifference
		}
		return teams[i].GoalsFor > teams[j].GoalsFor
	})
	return teams, nil
}

// updateTeamStatsInMemory is a helper to update a team's stats in-memory for simulations.
func updateTeamStatsInMemory(teamStats *models.Team, goalsScored int, goalsConceded int) {

	teamStats.Played++
	teamStats.GoalsFor += goalsScored
	teamStats.GoalsAgainst += goalsConceded
	teamStats.GoalDifference = teamStats.GoalsFor - teamStats.GoalsAgainst

	if goalsScored > goalsConceded {
		teamStats.Wins++
		teamStats.Points += 3
	} else if goalsScored == goalsConceded {
		teamStats.Draws++
		teamStats.Points += 1
	} else {
		teamStats.Losses++
	}
}

// GetChampionshipPredictions calculates championship probabilities using Monte Carlo simulation.
// Meaningful after 4 weeks are completed.
func (s *LeagueService) GetChampionshipPredictions(ctx context.Context) (map[int]float64, error) {

	nextPlayableWeek, err := s.GetCurrentWeek(ctx)
	if err != nil {
		return nil, fmt.Errorf("LeagueService.GetChampionshipPredictions: Could not determine current week: %w", err)
	}

	predictions := make(map[int]float64)

	if nextPlayableWeek == -1 {
		finalTable, errTable := s.GetLeagueTable(ctx)
		if errTable != nil || len(finalTable) == 0 {
			return nil, fmt.Errorf("LeagueService.GetChampionshipPredictions: League finished but final table could not be retrieved: %w", errTable)
		}
		for index, team := range finalTable {
			if index == 0 {
				predictions[team.ID] = 1.0
			} else {
				predictions[team.ID] = 0.0
			}
		}
		return predictions, nil
	}

	if nextPlayableWeek <= 4 && nextPlayableWeek > 0 {
		return nil, fmt.Errorf("championship predictions are available after at least 4 weeks are completed. Current playable week: %d", nextPlayableWeek)
	}

	originalTeamsFromDB, err := s.teamService.GetAllTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("LeagueService.GetChampionshipPredictions: Could not retrieve teams for prediction: %w", err)
	}
	if len(originalTeamsFromDB) == 0 {
		return nil, fmt.Errorf("LeagueService.GetChampionshipPredictions: No teams found in database for prediction")
	}

	teamsMapOriginal := make(map[int]models.Team)
	for _, t := range originalTeamsFromDB {
		teamsMapOriginal[t.ID] = t
	}

	allMatchesFromDB, err := s.matchService.GetAllMatches(ctx)
	if err != nil {
		return nil, fmt.Errorf("LeagueService.GetChampionshipPredictions: Could not retrieve matches for prediction: %w", err)
	}

	var unplayedMatches []models.Match
	for _, match := range allMatchesFromDB {
		if !match.IsPlayed {
			unplayedMatches = append(unplayedMatches, match)
		}
	}

	if len(unplayedMatches) == 0 && nextPlayableWeek != -1 {
		log.Println("LeagueService.GetChampionshipPredictions: Warning - No unplayed matches found, but league appears unfinished according to GetCurrentWeek. Predicting based on current table.")
		finalTable, _ := s.GetLeagueTable(ctx)
		if len(finalTable) > 0 {
			for _, team := range finalTable {
				predictions[team.ID] = 0.0
			}
			predictions[finalTable[0].ID] = 1.0
		}
		return predictions, nil
	}
	if len(unplayedMatches) == 0 {
		log.Println("LeagueService.GetChampionshipPredictions: No unplayed matches to simulate, league is considered finished.")
		finalTable, _ := s.GetLeagueTable(ctx)
		if len(finalTable) > 0 {
			for index, team := range finalTable {
				if index == 0 {
					predictions[team.ID] = 1.0
				} else {
					predictions[team.ID] = 0.0
				}
			}
		}
		return predictions, nil
	}

	numberOfSimulations := 2000
	championshipWinsCount := make(map[int]int)
	for _, team := range originalTeamsFromDB {
		championshipWinsCount[team.ID] = 0
	}

	for simCount := 0; simCount < numberOfSimulations; simCount++ {
		currentSimTeamStats := make(map[int]models.Team)
		for _, team := range originalTeamsFromDB {
			copiedTeam := team
			currentSimTeamStats[team.ID] = copiedTeam
		}

		for _, matchToSimulate := range unplayedMatches {
			homeTeamOriginal := teamsMapOriginal[matchToSimulate.HomeTeamID]
			awayTeamOriginal := teamsMapOriginal[matchToSimulate.AwayTeamID]

			homeGoals, awayGoals, _ := s.matchService.SimulateMatchOutcome(ctx, homeTeamOriginal, awayTeamOriginal)

			homeTeamSimStats := currentSimTeamStats[matchToSimulate.HomeTeamID]
			updateTeamStatsInMemory(&homeTeamSimStats, homeGoals, awayGoals)
			currentSimTeamStats[matchToSimulate.HomeTeamID] = homeTeamSimStats

			awayTeamSimStats := currentSimTeamStats[matchToSimulate.AwayTeamID]
			updateTeamStatsInMemory(&awayTeamSimStats, awayGoals, homeGoals)
			currentSimTeamStats[matchToSimulate.AwayTeamID] = awayTeamSimStats
		}

		var simTable []models.Team
		for _, teamStats := range currentSimTeamStats {
			simTable = append(simTable, teamStats)
		}

		sort.SliceStable(simTable, func(k, l int) bool {
			if simTable[k].Points != simTable[l].Points {
				return simTable[k].Points > simTable[l].Points
			}
			if simTable[k].GoalDifference != simTable[l].GoalDifference {
				return simTable[k].GoalDifference > simTable[l].GoalDifference
			}
			return simTable[k].GoalsFor > simTable[l].GoalsFor
		})

		if len(simTable) > 0 {
			championTeamID := simTable[0].ID
			championshipWinsCount[championTeamID]++
		}
	}

	if numberOfSimulations > 0 {
		for teamID, wins := range championshipWinsCount {
			predictions[teamID] = float64(wins) / float64(numberOfSimulations)
		}
	}
	return predictions, nil
}

// ResetLeague resets all team statistics and regenerates the fixture.
func (s *LeagueService) ResetLeague(ctx context.Context) error {

	log.Println("LeagueService.ResetLeague: League reset process STARTED.")
	err := s.teamService.ResetAllTeamStats(ctx)
	if err != nil {
		log.Printf("LeagueService.ResetLeague ERROR: Could not reset team statistics: %v", err)
		return fmt.Errorf("LeagueService.ResetLeague: Error while resetting team statistics: %w", err)
	}
	log.Println("LeagueService.ResetLeague: Call to reset team statistics made.")

	teams, err := s.teamService.GetAllTeams(ctx)
	if err != nil {
		log.Printf("LeagueService.ResetLeague ERROR: Could not fetch reset teams: %v", err)
		return fmt.Errorf("LeagueService.ResetLeague: Error retrieving teams for fixture (after stats reset): %w", err)
	}

	var teamsForFixture []models.Team
	if len(teams) >= 4 {
		teamsForFixture = teams[:4]
	} else {
		err := fmt.Errorf("LeagueService.ResetLeague: Insufficient teams to generate fixture. At least 4 teams required, found: %d", len(teams))
		log.Printf("LeagueService.ResetLeague ERROR: %v", err)
		return err
	}
	log.Printf("LeagueService.ResetLeague: %d teams will be used for the fixture.", len(teamsForFixture))

	err = s.matchService.GenerateAndStoreFixture(ctx, teamsForFixture)
	if err != nil {
		log.Printf("LeagueService.ResetLeague ERROR: Could not regenerate fixture: %v", err)
		return fmt.Errorf("LeagueService.ResetLeague: Error regenerating fixture: %w", err)
	}
	log.Println("LeagueService.ResetLeague: League successfully reset (statistics and fixture).")
	return nil
}

// PlayAllRemainingWeeks plays all remaining unplayed weeks in the league.
func (s *LeagueService) PlayAllRemainingWeeks(ctx context.Context) (map[int][]models.Match, []models.Team, error) {

	allPlayedMatchesByWeek := make(map[int][]models.Match)
	var finalLeagueTable []models.Team
	var lastSuccessfullyPlayedWeek int

	log.Println("LeagueService.PlayAllRemainingWeeks: Playing all remaining weeks...")
	for i := 0; i < 10; i++ {
		nextWeekToPlay, err := s.GetCurrentWeek(ctx)
		if err != nil {
			return allPlayedMatchesByWeek, finalLeagueTable, fmt.Errorf("LeagueService.PlayAllRemainingWeeks: Error determining current week: %w", err)
		}

		if nextWeekToPlay == -1 {
			log.Println("LeagueService.PlayAllRemainingWeeks: League already completed.")
			break
		}

		playedWeek, weekMatches, currentLeagueTable, playErr := s.PlayNextWeek(ctx)

		if playErr != nil {
			log.Printf("LeagueService.PlayAllRemainingWeeks: Error playing week %d: %v. Halting simulation.", nextWeekToPlay, playErr)
			if currentLeagueTable != nil {
				finalLeagueTable = currentLeagueTable
			} else if lastSuccessfullyPlayedWeek > 0 {
				finalLeagueTable, _ = s.GetLeagueTable(ctx)
			}
			return allPlayedMatchesByWeek, finalLeagueTable, playErr
		}

		if playedWeek == 0 {
			log.Println("LeagueService.PlayAllRemainingWeeks: League completed as of this iteration.")
			finalLeagueTable = currentLeagueTable
			if len(weekMatches) > 0 && nextWeekToPlay > 0 {
				allPlayedMatchesByWeek[nextWeekToPlay] = weekMatches
			}
			break
		}

		if playedWeek > 0 && len(weekMatches) > 0 {
			allPlayedMatchesByWeek[playedWeek] = weekMatches
			lastSuccessfullyPlayedWeek = playedWeek
		}
		finalLeagueTable = currentLeagueTable
	}

	if finalLeagueTable == nil {
		var errTable error
		finalLeagueTable, errTable = s.GetLeagueTable(ctx)
		if errTable != nil {
			return allPlayedMatchesByWeek, nil, fmt.Errorf("LeagueService.PlayAllRemainingWeeks: Error retrieving final league table after loop: %w", errTable)
		}
	}

	log.Printf("LeagueService.PlayAllRemainingWeeks: All remaining weeks played (last successfully played week: %d).", lastSuccessfullyPlayedWeek)
	return allPlayedMatchesByWeek, finalLeagueTable, nil
}

// HandleMatchScoreEdit manages editing a match score and adjusting team statistics.
func (s *LeagueService) HandleMatchScoreEdit(ctx context.Context, matchID int, newHomeGoals int, newAwayGoals int) error {

	log.Printf("LeagueService.HandleMatchScoreEdit: Score edit process started for Match ID %d. New score: %d-%d", matchID, newHomeGoals, newAwayGoals)

	originalMatch, err := s.matchService.EditMatchScore(ctx, matchID, newHomeGoals, newAwayGoals)
	if err != nil {
		return fmt.Errorf("HandleMatchScoreEdit: Error updating match score via MatchService: %w", err)
	}

	var oldHomeScoreForStatAdjust, oldAwayScoreForStatAdjust int
	if originalMatch.IsPlayed && originalMatch.HomeGoals != nil && originalMatch.AwayGoals != nil {
		oldHomeScoreForStatAdjust = *originalMatch.HomeGoals
		oldAwayScoreForStatAdjust = *originalMatch.AwayGoals
	} else {
		if !originalMatch.IsPlayed {
			log.Printf("HandleMatchScoreEdit: Warning - Match ID %d was not previously marked as played. Its old contribution to stats is 0. The edit will mark it as played.", matchID)
		}
		oldHomeScoreForStatAdjust = 0
		oldAwayScoreForStatAdjust = 0
		if originalMatch.HomeGoals != nil {
			oldHomeScoreForStatAdjust = *originalMatch.HomeGoals
		}
		if originalMatch.AwayGoals != nil {
			oldAwayScoreForStatAdjust = *originalMatch.AwayGoals
		}
	}

	log.Printf("  Adjusting stats for home team (ID: %d)... Old: %d-%d, New: %d-%d",
		originalMatch.HomeTeamID, oldHomeScoreForStatAdjust, oldAwayScoreForStatAdjust, newHomeGoals, newAwayGoals)
	err = s.teamService.AdjustTeamStatsForScoreChange(ctx, originalMatch.HomeTeamID,
		oldHomeScoreForStatAdjust, oldAwayScoreForStatAdjust,
		newHomeGoals, newAwayGoals,
	)
	if err != nil {
		return fmt.Errorf("HandleMatchScoreEdit: Error adjusting stats for home team (ID: %d): %w", originalMatch.HomeTeamID, err)
	}

	log.Printf("  Adjusting stats for away team (ID: %d)... Old: %d-%d, New: %d-%d",
		originalMatch.AwayTeamID, oldAwayScoreForStatAdjust, oldHomeScoreForStatAdjust, newAwayGoals, newHomeGoals)
	err = s.teamService.AdjustTeamStatsForScoreChange(ctx, originalMatch.AwayTeamID,
		oldAwayScoreForStatAdjust, oldHomeScoreForStatAdjust,
		newAwayGoals, newHomeGoals,
	)
	if err != nil {
		return fmt.Errorf("HandleMatchScoreEdit: Error adjusting stats for away team (ID: %d): %w", originalMatch.AwayTeamID, err)
	}

	log.Printf("LeagueService.HandleMatchScoreEdit: Score edit and stat adjustment completed for Match ID %d.", matchID)
	return nil
}
