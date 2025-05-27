package queries

const (
	// DeleteAllMatchesSQL, matches tablosundaki tüm kayıtları siler.
	DeleteAllMatchesSQL = `DELETE FROM matches`

	// InsertMatchSQL, yeni bir maçı matches tablosuna ekler.
	// Parametreler: $1=week, $2=home_team_id, $3=away_team_id, $4=is_played, $5=home_goals, $6=away_goals
	InsertMatchSQL = `
		INSERT INTO matches (week, home_team_id, away_team_id, is_played, home_goals, away_goals)
		VALUES ($1, $2, $3, $4, $5, $6)`

	// GetMatchesByWeekSQL, belirtilen haftadaki maçları ID'ye göre sıralı getirir.
	// Parametreler: $1 = week
	GetMatchesByWeekSQL = `
		SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, is_played
		FROM matches
		WHERE week = $1
		ORDER BY id ASC`

	// GetMatchByIDSQL, ID'ye göre bir maçı getirir.
	// Parametreler: $1 = matchID
	GetMatchByIDSQL = `
		SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, is_played
		FROM matches
		WHERE id = $1`

	// UpdateMatchResultSQL, bir maçın skorunu ve oynanma durumunu günceller.
	// Parametreler: $1=home_goals, $2=away_goals, $3=is_played, $4=matchID
	UpdateMatchResultSQL = `
		UPDATE matches
		SET home_goals = $1, away_goals = $2, is_played = $3
		WHERE id = $4`

	// GetAllMatchesSQL, tüm maçları hafta ve ID'ye göre sıralı getirir.
	GetAllMatchesSQL = `
		SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, is_played
		FROM matches
		ORDER BY week ASC, id ASC`
)
