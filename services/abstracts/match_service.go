package abstracts

import (
	"MatchSimulator_Insider/models"
	"context"
)

type IMatchService interface {
	GenerateAndStoreFixture(ctx context.Context, teams []models.Team) error
	GetMatchesByWeek(ctx context.Context, week int) ([]models.Match, error)
	GetMatchByID(ctx context.Context, id int) (*models.Match, error)
	UpdateMatchResult(ctx context.Context, matchID int, homeGoals, awayGoals int, isPlayed bool) error // Bu zaten vardı, skor güncelleme için kullanılabilir.
	GetAllMatches(ctx context.Context) ([]models.Match, error)
	SimulateMatchOutcome(ctx context.Context, homeTeam models.Team, awayTeam models.Team) (homeGoals int, awayGoals int, err error)

	// EditMatchScore, belirli bir maçın skorunu günceller ve eski maç verisini döndürür.
	// Maçın 'is_played' durumu true olarak güncellenir.
	EditMatchScore(ctx context.Context, matchID int, newHomeGoals int, newAwayGoals int) (originalMatch models.Match, err error) // YENİ METOT
}
