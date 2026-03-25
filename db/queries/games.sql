-- name: ListGames :many
SELECT
    g.id, g.game_date, g.game_time,
    g.home_team_id, g.away_team_id, g.venue_id, g.league_id, g.position,
    g.referee_fee, g.travel_costs, g.km_driven, g.exhibition, g.remarks,
    g.created_at, g.updated_at,
    ht.name AS home_team_name,
    at2.name AS away_team_name,
    COALESCE(v.city, '') AS venue_city, COALESCE(v.stadium, '') AS venue_stadium,
    l.name AS league_name, l.short_name AS league_short_name
FROM games g
JOIN teams ht ON g.home_team_id = ht.id
JOIN teams at2 ON g.away_team_id = at2.id
LEFT JOIN venues v ON g.venue_id = v.id
JOIN leagues l ON g.league_id = l.id
ORDER BY g.game_date DESC, g.game_time ASC;

-- name: GetGame :one
SELECT id, game_date, game_time,
    home_team_id, away_team_id, venue_id, league_id, position,
    referee_fee, travel_costs, km_driven, exhibition, remarks,
    created_at, updated_at
FROM games
WHERE id = ?;

-- name: CreateGame :one
INSERT INTO games (
    game_date, game_time, home_team_id, away_team_id, venue_id,
    league_id, position, referee_fee, travel_costs, km_driven,
    exhibition, remarks, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
    datetime('now', 'localtime'), datetime('now', 'localtime'))
RETURNING *;

-- name: UpdateGame :exec
UPDATE games
SET game_date = ?, game_time = ?, home_team_id = ?, away_team_id = ?,
    venue_id = ?, league_id = ?, position = ?, referee_fee = ?,
    travel_costs = ?, km_driven = ?, exhibition = ?, remarks = ?,
    updated_at = datetime('now', 'localtime')
WHERE id = ?;

-- name: DeleteGame :exec
DELETE FROM games WHERE id = ?;

-- name: ListSeasons :many
SELECT DISTINCT substr(game_date, 1, 4) AS season
FROM games
ORDER BY season DESC;
