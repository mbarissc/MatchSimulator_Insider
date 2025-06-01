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


type PostgresMatchService struct {
	DB *pgx.Conn
}


func NewPostgresMatchService(db *pgx.Conn) abstracts.IMatchService {
	return &PostgresMatchService{DB: db}
}


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

	// 6 hafta totalde 12 maçın fikstür yapısı
	schedule := [][][2]int{
		{{teamIDs[0], teamIDs[1]}, {teamIDs[2], teamIDs[3]}}, // 1. hafta
		{{teamIDs[0], teamIDs[2]}, {teamIDs[1], teamIDs[3]}}, // 2. hafta
		{{teamIDs[0], teamIDs[3]}, {teamIDs[1], teamIDs[2]}}, // 3. hafta
		{{teamIDs[1], teamIDs[0]}, {teamIDs[3], teamIDs[2]}}, // 4. hafta
		{{teamIDs[2], teamIDs[0]}, {teamIDs[3], teamIDs[1]}}, // 5. hafta
		{{teamIDs[3], teamIDs[0]}, {teamIDs[2], teamIDs[1]}}, // 6. hafta
	}

	currentWeek := 1

	// iç içe for döngüsüyle her maç ayrı ayrı bilgileriyle matchesToCreate slice'ına yazdırılır
	for _, weeklyMatches := range schedule {
		for _, matchPair := range weeklyMatches {
			matchesToCreate = append(matchesToCreate, models.Match{
				Week:       currentWeek,
				HomeTeamID: matchPair[0],
				AwayTeamID: matchPair[1],
				IsPlayed:   false,
				HomeGoals:  nil, // henüz oynanmadı
				AwayGoals:  nil, // henüz oynanmadı
			})
		}
		currentWeek++
	}

	// transaction başlatılır
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PostgresMatchService.GenerateAndStoreFixture: Could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// maçlar veritabanına insert edilir
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

// Belirli bir haftanın maçlarını getirir
func (s *PostgresMatchService) GetMatchesByWeek(ctx context.Context, week int) ([]models.Match, error) {
	
	// maçlar veritabanından çekilir
	rows, err := s.DB.Query(ctx, queries.GetMatchesByWeekSQL, week)
	if err != nil {
		return nil, fmt.Errorf("PostgresMatchService.GetMatchesByWeek: Error retrieving matches for week %d: %w", week, err)
	}
	defer rows.Close()

	var matches []models.Match
	
	// rows nesnesi üstünden match verileri alınır ve matches slice'ına yazdırılır.
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

// id'ye göre maç bilgilerini alma metodu.
// dönüş tipi pointer olarak belirlenmesinin sebebi maç veritabanında bulunamadığında bunu nil ile ifade edebilmektir.
// doğrudan models.Match olarak döndürseydi gerçekten boş bir maç mı döndü yoksa maç mı bulunamadı ayrımını yapmak daha zor olurdu.
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

// veritabanındaki tüm maçları hafta ve id ye göre sıralı şekilde alır
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

// İki takım arasındaki oynanan maçları simüle eder
func (s *PostgresMatchService) SimulateMatchOutcome(ctx context.Context, homeTeam models.Team, awayTeam models.Team) (homeGoals int, awayGoals int, err error) {
	
	maxPotentialGoals := 6 // Atılabilecek maksimum potansiyel gol (her iki takım için ayrı ayrı)
	strengthDivisor := 140
	homeAdvantage := 10 // Ev sahibi takım için +10 bonus strength verilir

	effectiveHomeStrength := homeTeam.Strength + homeAdvantage
	
	if effectiveHomeStrength < 0 {
		effectiveHomeStrength = 0
	}

	effectiveAwayStrength := awayTeam.Strength
	if effectiveAwayStrength < 0 {
		effectiveAwayStrength = 0
	}

	// gol hesaplama
	for i := 0; i < maxPotentialGoals; i++ {
		// strengthDivisor ile rastgele bir sayı üretilir (0-140) bu sayı efektif güçten düşükse takım gol attı kabul edilir
		if rand.Intn(strengthDivisor) < effectiveHomeStrength {
			homeGoals++
		}
		
		if rand.Intn(strengthDivisor) < effectiveAwayStrength {
			awayGoals++
		}
	}
	
	return homeGoals, awayGoals, nil
}


func (s *PostgresMatchService) EditMatchScore(ctx context.Context, matchID int, newHomeGoals int, newAwayGoals int) (originalMatch models.Match, err error) {
	log.Printf("PostgresMatchService.EditMatchScore: Initiating score edit for Match ID %d. New score: %d-%d", matchID, newHomeGoals, newAwayGoals)

	
	originalMatchPtr, err := s.GetMatchByID(ctx, matchID)
	if err != nil {
		return models.Match{}, fmt.Errorf("PostgresMatchService.EditMatchScore: Could not find or retrieve match to edit (ID: %d): %w", matchID, err)
	}
	originalMatch = *originalMatchPtr

	err = s.UpdateMatchResult(ctx, matchID, newHomeGoals, newAwayGoals, true)
	if err != nil {
		
		return originalMatch, fmt.Errorf("PostgresMatchService.EditMatchScore: Error updating score for match (ID: %d): %w", matchID, err)
	}

	log.Printf("PostgresMatchService.EditMatchScore: Score for Match ID %d successfully updated to %d-%d.", matchID, newHomeGoals, newAwayGoals)
	return originalMatch, nil
}
