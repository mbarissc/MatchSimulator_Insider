# Football League Simulation API

## Overview

This project is a GoLang application that simulates a 4-team football league. It calculates match results based on team strengths, updates a league table according to Premier League rules, shows weekly progress, and provides championship predictions after the 4th week using Monte Carlo simulation. All interactions are managed via API endpoints.

The project utilizes an interface-based design and struct composition as requested. It uses PostgreSQL for data persistence.



## 🌐 Live API Access 

This project has been deployed to Google Cloud Run and is publicly accessible via the following link:

**Base API URL:** `https://football-sim-api-606877919106.europe-west1.run.app`



## Features

* **4-Team League Simulation:** Simulates a full season for 4 distinct football teams.
* **Team Strengths:** Teams can have different strength values, which influence match outcomes. Team names and strengths can be updated via API.
* **Premier League Rules:** Applies standard Premier League rules for match points (3 for a win, 1 for a draw) and league table sorting (Points > Goal Difference > Goals For). 
* **Weekly Progression:** Simulates the league week by week. 
* **Match Results & League Table:** Displays match results and the updated league table after each week. 
* **Championship Predictions:** Provides championship probability estimations for each team after the 4th week. 
* **API Driven:** All league operations are managed through well-defined API endpoints. 
* **Full Season Simulation (`/play-all`):** (Extra Feature) Plays all remaining weeks automatically and lists results by week. 
* **Edit Match Results (`/matches/{id}`):** (Extra Feature) Allows editing scores of previously played matches, with automatic recalculation of standings. 
* **Team Customization:** API endpoints to update team names and strengths.
* **League Reset:** API endpoints to reset the league to its initial state or reset teams to default configurations.

---
## Setup and Installation

Follow these instructions to set up and run the project locally.

### Prerequisites

* **Go:** Version 1.20 or later is recommended. ([Go Installation Guide](https://go.dev/doc/install))
* **PostgreSQL:** Version 12 or later is recommended. ([PostgreSQL Downloads](https://www.postgresql.org/download/))
* An API testing tool like Postman or `curl`.

### Database Setup

1.  **Install PostgreSQL:** Ensure PostgreSQL is installed and running on your system.
2.  **Create Database:**
    * Connect to your PostgreSQL instance (e.g., using `psql -U your_postgres_user`).
    * Create a new database. The application can be configured to use a specific database name via `config.json`. A suggested name for your main database is `football_league_sim`.
        ```sql
        CREATE DATABASE football_league_sim;
        ```
    * (For tests, it's also recommended to create a separate test database, e.g., `football_league_sim_test`.)
3.  **Create Tables:** Connect to your newly created database (e.g., `\c football_league_sim` in `psql`) and run the SQL commands found in **Section 5: SQL Schema** of this document.

### Application Setup

1.  **Clone/Download Project:** Get the project source code.
2.  **Navigate to Project Root:** Open your terminal and navigate to the project's root directory.
3.  **Install Dependencies:** Run to fetch Go module dependencies:
    ```bash
    go mod tidy
    ```
4.  **Configuration (`config.json`):**
    * In the project root, create a `config.json` file. You can use the template below (or create a `config.example.json` with this content in your repository).
    * Update the `connectionString` with your actual PostgreSQL details.

        **`config.json` Template (Use your actual values locally):**
        ```json
        {
          "database": {
            "connectionString": "postgres://YOUR_DB_USER:YOUR_DB_PASSWORD@YOUR_DB_HOST:YOUR_DB_PORT/YOUR_DB_NAME?sslmode=disable"
          },
          "server": {
            "port": "YOUR_API_PORT"
          }
        }
        ```
        **Example of a filled `config.json` (for local use, do not commit with real credentials to public repos):**
        ```json
        {
          "database": {
            "connectionString": "postgres://postgres:1234@localhost:5432/football_league_sim?sslmode=disable"
          },
          "server": {
            "port": "8080"
          }
        }
        ```
    * **Important:** If you are committing this project to a public repository, ensure your actual `config.json` (with real credentials) is listed in your `.gitignore` file.
5.  **Run the Application:**
    ```bash
    go run main.go
    ```
    The API server will start, typically on `http://localhost:8080` (or the port specified in `config.json`).

## 4. SQL Schema

The database schema consists of two main tables: `teams` and `matches`.

```sql
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    strength INTEGER DEFAULT 50,
    played INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    draws INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    goals_for INTEGER DEFAULT 0,
    goals_against INTEGER DEFAULT 0,
    goal_difference INTEGER DEFAULT 0,
    points INTEGER DEFAULT 0
);

CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    week INTEGER NOT NULL,
    home_team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    away_team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    home_goals INTEGER,
    away_goals INTEGER,
    is_played BOOLEAN DEFAULT FALSE,
    CONSTRAINT check_different_teams CHECK (home_team_id <> away_team_id)
);

CREATE INDEX idx_matches_week ON matches(week);
```

---
## 5. API Endpoint Documentation

The API allows interaction with the football league simulation. All request/response bodies are in JSON format. The default base URL is `http://localhost:8080`.

*(Note: Team and Match IDs in examples are illustrative and may vary.)*

### League State & Progression

* **`GET /league-table`**
    * **Description:** Retrieves the current league standings.
    * **Success Response (200 OK):** Array of team objects with their stats.
        ```json
        [
            {"id":1,"name":"Chelsea","strength":99,"played":6,"wins":3,"draws":2,"losses":1,"goals_for":26,"goals_against":26,"goal_difference":0,"points":11}
            // ... other teams
        ]
        ```

* **`POST /next-week`**
    * **Description:** Simulates the next unplayed week.
    * **Success Response (200 OK):**
        ```json
        {
            "played_week": 1,
            "week_matches": [
                {"id":1,"week":1,"home_team_id":2,"away_team_id":1,"home_goals":3,"away_goals":4,"is_played":true}
            ],
            "league_table": [ /* updated league table */ ],
            "message": "Week 1 played successfully."
        }
        ```

* **`GET /current-week`**
    * **Description:** Returns the current playable week number and league status.
    * **Success Response (200 OK):**
        * In progress: `{"current_playable_week": 3, "league_status": "In Progress", "status_message": "Current playable week: 3"}`
        * Completed: `{"current_playable_week": -1, "league_status": "Completed", "status_message": "All matches have been played, the league is completed."}`

* **`POST /play-all`** (Extra Feature)
    * **Description:** Simulates all remaining weeks of the league.
    * **Success Response (200 OK):**
        ```json
        {
            "message": "All remaining weeks played successfully.",
            "played_matches_by_week": {
                "4": [/* matches for week 4 */]
            },
            "final_league_table": [ /* final league table */ ]
        }
        ```

### Predictions

* **`GET /predictions`**
    * **Description:** Retrieves championship predictions. Available after 4 weeks are completed. 
    * **Success Response (200 OK):**
        ```json
        [
            {"team_name":"Liverpool","team_id":4,"probability_percentage":60.50}
            // ... other teams
        ]
        ```
    * **Error Response (412 Precondition Failed):** If called before 4 weeks are complete.

### Management & Editing

* **`POST /reset-league`**
    * **Description:** Resets all team statistics and regenerates a fresh fixture. Team names and strengths are NOT reset by this.
    * **Success Response (200 OK):** `{"message": "League reset successfully. Team statistics and fixture have been renewed."}`

* **`POST /teams/reset-defaults`**
    * **Description:** Resets all teams to their default names and strengths. Also resets all league statistics and regenerates the fixture.
    * **Success Response (200 OK):**
        ```json
        {
            "message": "All teams have been reset to default names and strengths. League statistics and fixture have also been renewed.",
            "league_table": [ /* league table with teams having 0 stats and default names/strengths */ ]
        }
        ```

* **`PUT /teams/{id}/strength`**
    * **Description:** Updates the strength of a specific team.
    * **Path Parameter:** `{id}` - ID of the team.
    * **Request Body (JSON):** `{"strength": 95}`
    * **Success Response (200 OK):**
        ```json
        {
            "message": "Team ID <id> strength successfully updated to 95.",
            "team": { /* updated team object */ }
        }
        ```

* **`PUT /teams/{id}/name`**
    * **Description:** Updates the name of a specific team. Name must be unique.
    * **Path Parameter:** `{id}` - ID of the team.
    * **Request Body (JSON):** `{"name": "New Club Name"}`
    * **Success Response (200 OK):**
        ```json
        {
            "message": "Team ID <id> name successfully updated to 'New Club Name'.",
            "team": { /* updated team object */ }
        }
        ```

* **`PUT /matches/{id}`** (Extra Feature)
    * **Description:** Edits the score of a previously played match. Standings are recalculated. 
    * **Path Parameter:** `{id}` - ID of the match.
    * **Request Body (JSON):** `{"home_goals": 3, "away_goals": 1}`
    * **Success Response (200 OK):**
        ```json
        {
            "message": "Match ID <id> score successfully updated to 3-1. League table and statistics have been refreshed.",
            "league_table": [ /* updated league table */ ]
        }
        ```

---


## 6. Key SQL Queries Used

The application utilizes SQL queries stored as constants in the /queries package.

From queries/team_queries.go:
`
// CreateTeamCheckExistsSQL: Checks if a team exists by name.
const CreateTeamCheckExistsSQL =
```sql
`SELECT id FROM teams WHERE name = $1`
```
// CreateTeamInsertSQL: Inserts a new team with zeroed stats.
const CreateTeamInsertSQL = 
    ```sql
    
    INSERT INTO teams (name, strength, played, wins, draws, losses, goals_for, goals_against, goal_difference, points)
    VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0, 0)
    RETURNING id`


// GetTeamByIDSQL: Retrieves a team by its ID.
const GetTeamByIDSQL =

   ```sql
    SELECT id, name, strength, played, wins, draws, losses, goals_for, goals_against, goal_difference, points 
    FROM teams 
    WHERE id = $1`
   ```
// GetAllTeamsSQL: Retrieves all teams, ordered for league table display.
const GetAllTeamsSQL = 
```sql
    SELECT id, name, strength, played, wins, draws, losses, goals_for, goals_against, goal_difference, points 
    FROM teams 
    ORDER BY points DESC, goal_difference DESC, goals_for DESC, name ASC`
```

// UpdateTeamMainStatsSQL: Updates a team's main statistics after a match.
const UpdateTeamMainStatsSQL = 
```sql
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
```

// UpdateTeamGDSQL: Updates a team's goal difference.
const UpdateTeamGDSQL = 
```sql
`UPDATE teams SET goal_difference = goals_for - goals_against WHERE id = $1`
```
// ResetAllTeamStatsSQL: Resets all statistics for all teams to zero.
const ResetAllTeamStatsSQL = `
   ```sql 
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
```
// AdjustTeamStatsSQL: Adjusts a team's statistics after a match score edit.
const AdjustTeamStatsSQL = 
   
   ```sql
    UPDATE teams
    SET
        wins = wins + $1,            
        draws = draws + $2,            
        losses = losses + $3,          
        goals_for = goals_for + $4,    
        goals_against = goals_against + $5, 
        points = points + $6           
    WHERE id = $7

```

// UpdateTeamStrengthSQL: Updates a team's strength.
const UpdateTeamStrengthSQL = 
```sql
`UPDATE teams SET strength = $1 WHERE id = $2`
```
// UpdateTeamNameSQL: Updates a team's name.
const UpdateTeamNameSQL =
```sql
`UPDATE teams SET name = $1 WHERE id = $2`
```
// UpdateTeamNameAndStrengthSQL: Updates a team's name and strength.
const UpdateTeamNameAndStrengthSQL = 
```sql
`UPDATE teams SET name = $1, strength = $2 WHERE id = $3`
```
// GetAllTeamsOrderedByIDSQL: Retrieves all teams ordered by their ID.
const GetAllTeamsOrderedByIDSQL = 
    
   ```sql   
    SELECT id, name, strength, played, wins, draws, losses, goals_for, goals_against, goal_difference, points 
    FROM teams 
    ORDER BY id ASC`
   ```

From queries/match_queries.go:

// DeleteAllMatchesSQL: Deletes all records from the matches table.
const DeleteAllMatchesSQL = 
```sql
`DELETE FROM matches`
```

// InsertMatchSQL: Inserts a new match into the matches table.
const InsertMatchSQL = `
   ```sql
    INSERT INTO matches (week, home_team_id, away_team_id, is_played, home_goals, away_goals)
    VALUES ($1, $2, $3, $4, $5, $6)`
```

// GetMatchesByWeekSQL: Retrieves matches for a specific week, ordered by ID.
const GetMatchesByWeekSQL = `
```sql
    SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, is_played
    FROM matches
    WHERE week = $1
    ORDER BY id ASC`
```

// GetMatchByIDSQL: Retrieves a match by its ID.
const GetMatchByIDSQL = `
```sql
    SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, is_played
    FROM matches
    WHERE id = $1`
```

// UpdateMatchResultSQL: Updates the score and played status of a match.
const UpdateMatchResultSQL = `
```sql
    UPDATE matches
    SET home_goals = $1, away_goals = $2, is_played = $3
    WHERE id = $4`
```

// GetAllMatchesSQL: Retrieves all matches, ordered by week and then ID.
const GetAllMatchesSQL = `
```sql
    SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, is_played
    FROM matches
    ORDER BY week ASC, id ASC`
```












