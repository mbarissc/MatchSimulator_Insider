package abstracts

import (
	"MatchSimulator_Insider/models"
	"context"
)

type ILeagueService interface {
	PlayNextWeek(ctx context.Context) (int, []models.Match, []models.Team, error)
	GetLeagueTable(ctx context.Context) ([]models.Team, error)
	GetCurrentWeek(ctx context.Context) (int, error)
	GetChampionshipPredictions(ctx context.Context) (map[int]float64, error)
	ResetLeague(ctx context.Context) error
	PlayAllRemainingWeeks(ctx context.Context) (map[int][]models.Match, []models.Team, error)
	HandleMatchScoreEdit(ctx context.Context, matchID int, newHomeGoals int, newAwayGoals int) error // YENİ METOT
}
