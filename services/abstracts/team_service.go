package abstracts

import (
	"MatchSimulator_Insider/models"
	"context"
)

type TeamService interface {
	CreateTeam(ctx context.Context, team models.Team) (int, error)
	GetTeamByID(ctx context.Context, id int) (*models.Team, error)
	GetAllTeams(ctx context.Context) ([]models.Team, error)
	UpdateTeamStatsAfterMatch(ctx context.Context, teamID int, goalsScored int, goalsConceded int) error // Bu, normal maç oynandığında kullanılır.
	ResetAllTeamStats(ctx context.Context) error
	AdjustTeamStatsForScoreChange(ctx context.Context, teamID int, oldGoalsForTeam, oldGoalsAgainstTeam, newGoalsForTeam, newGoalsAgainstTeam int) error // YENİ METOT
	UpdateTeamStrength(ctx context.Context, teamID int, newStrength int) error
	UpdateTeamName(ctx context.Context, teamID int, newName string) error
	ResetTeamsToDefaults(ctx context.Context) error
}
