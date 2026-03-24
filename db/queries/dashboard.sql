-- name: ListGamesFull :many
SELECT
    g.game_date,
    substr(g.game_date, 1, 4) AS year,
    g.position,
    g.league_id,
    COALESCE(NULLIF(l.short_name, ''), l.name) AS league_short,
    g.referee_fee,
    g.travel_costs,
    g.km_driven,
    COALESCE(NULLIF(v.stadium || ', ' || v.city, ', ' || v.city), v.city, '') AS venue_display,
    COALESCE(v.lat, 0.0) AS venue_lat,
    COALESCE(v.lon, 0.0) AS venue_lon
FROM games g
JOIN leagues l ON g.league_id = l.id
LEFT JOIN venues v ON g.venue_id = v.id
ORDER BY g.game_date;

-- name: ListGamesBySeason :many
SELECT
    g.game_date,
    g.game_time,
    substr(g.game_date, 6, 2) AS month,
    ht.name AS home_team_name,
    at2.name AS away_team_name,
    COALESCE(NULLIF(v.stadium || ', ' || v.city, ', ' || v.city), v.city, '') AS venue_display,
    COALESCE(v.lat, 0.0) AS venue_lat,
    COALESCE(v.lon, 0.0) AS venue_lon,
    COALESCE(NULLIF(l.short_name, ''), l.name) AS league_short,
    l.name AS league_long,
    g.league_id,
    g.position,
    g.referee_fee,
    g.travel_costs,
    g.km_driven,
    g.exhibition
FROM games g
JOIN teams ht ON g.home_team_id = ht.id
JOIN teams at2 ON g.away_team_id = at2.id
LEFT JOIN venues v ON g.venue_id = v.id
JOIN leagues l ON g.league_id = l.id
WHERE substr(g.game_date, 1, 4) = ?
ORDER BY g.game_date;
