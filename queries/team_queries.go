// FILE: queries/team_queries.go
package queries

const (
	// CreateTeamCheckExistsSQL, bir takımın isme göre var olup olmadığını kontrol eder.
	CreateTeamCheckExistsSQL = `SELECT id FROM teams WHERE name = $1`

	// CreateTeamInsertSQL, yeni bir takımı sıfır istatistikle ekler.
	// Parametreler: $1 name, $2 strength
	CreateTeamInsertSQL = `
		INSERT INTO teams (name, strength, played, wins, draws, losses, goals_for, goals_against, goal_difference, points)
		VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0, 0)
		RETURNING id`

	// GetTeamByIDSQL, ID'ye göre bir takımı getirir.
	// Parametreler: $1 = teamID
	GetTeamByIDSQL = `
		SELECT id, name, strength, played, wins, draws, losses, goals_for, goals_against, goal_difference, points 
		FROM teams 
		WHERE id = $1`

	// GetAllTeamsSQL, tüm takımları puan durumu sıralamasına göre getirir.
	GetAllTeamsSQL = `
		SELECT id, name, strength, played, wins, draws, losses, goals_for, goals_against, goal_difference, points 
		FROM teams 
		ORDER BY points DESC, goal_difference DESC, goals_for DESC, name ASC`

	// UpdateTeamMainStatsSQL, bir maç sonrası takımın ana istatistiklerini günceller.
	// Parametreler: $1=winIncrement, $2=drawIncrement, $3=lossIncrement, $4=goalsScored, $5=goalsConceded, $6=pointsEarned, $7=teamID
	UpdateTeamMainStatsSQL = `
		UPDATE teams
		SET
			played = played + 1,
			wins = wins + $1,        
			draws = draws + $2,        
			losses = losses + $3,      
			goals_for = goals_for + $4,  
			goals_against = goals_against + $5, 
			points = points + $6       
		WHERE id = $7`

	// UpdateTeamGDSQL, bir takımın gol averajını günceller.
	// Parametreler: $1 = teamID
	UpdateTeamGDSQL = `UPDATE teams SET goal_difference = goals_for - goals_against WHERE id = $1`

	// ResetAllTeamStatsSQL, tüm takımların istatistiklerini sıfırlar.
	ResetAllTeamStatsSQL = `
		UPDATE teams
		SET
			played = 0,
			wins = 0,
			draws = 0,
			losses = 0,
			goals_for = 0,
			goals_against = 0,
			goal_difference = 0,
			points = 0`

	// AdjustTeamStatsSQL, skor değişikliği sonrası takımın ana istatistiklerini ayarlar.
	// Parametreler: $1=deltaWins, $2=deltaDraws, $3=deltaLosses, $4=deltaGoalsFor, $5=deltaGoalsAgainst, $6=deltaPoints, $7=teamID
	AdjustTeamStatsSQL = `
		UPDATE teams
		SET
			wins = wins + $1,            
			draws = draws + $2,            
			losses = losses + $3,          
			goals_for = goals_for + $4,    
			goals_against = goals_against + $5, 
			points = points + $6           
		WHERE id = $7`

	// UpdateTeamStrengthSQL, bir takımın gücünü günceller.
	// Parametreler: $1=newStrength, $2=teamID
	UpdateTeamStrengthSQL = `UPDATE teams SET strength = $1 WHERE id = $2`

	// UpdateTeamNameSQL, bir takımın ismini günceller.
	// Parametreler: $1=newName, $2=teamID
	UpdateTeamNameSQL = `UPDATE teams SET name = $1 WHERE id = $2`

	// UpdateTeamNameAndStrengthSQL, bir takımın ismini ve gücünü günceller.
	// Parametreler: $1=newName, $2=newStrength, $3=teamID
	UpdateTeamNameAndStrengthSQL = `UPDATE teams SET name = $1, strength = $2 WHERE id = $3`

	// GetAllTeamsOrderedByIDSQL, tüm takımları ID'ye göre sıralı getirir (varsayılana reset için).
	GetAllTeamsOrderedByIDSQL = `
		SELECT id, name, strength, played, wins, draws, losses, goals_for, goals_against, goal_difference, points 
		FROM teams 
		ORDER BY id ASC`
)
