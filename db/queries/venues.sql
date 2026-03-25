-- name: ListVenues :many
SELECT * FROM venues ORDER BY short_name, city;

-- name: GetVenue :one
SELECT * FROM venues WHERE id = ?;

-- name: CreateVenue :one
INSERT INTO venues (city, stadium, short_name, lat, lon, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, datetime('now', 'localtime'), datetime('now', 'localtime'))
RETURNING *;

-- name: UpdateVenue :exec
UPDATE venues
SET city = ?, stadium = ?, short_name = ?, lat = ?, lon = ?, updated_at = datetime('now', 'localtime')
WHERE id = ?;

-- name: UpdateVenueCoords :exec
UPDATE venues
SET lat = ?, lon = ?, updated_at = datetime('now', 'localtime')
WHERE id = ?;

-- name: DeleteVenue :exec
DELETE FROM venues WHERE id = ?;
