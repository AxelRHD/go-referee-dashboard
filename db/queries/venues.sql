-- name: ListVenues :many
SELECT id, city, stadium, lat, lon, created_at, updated_at
FROM venues
ORDER BY city;

-- name: GetVenue :one
SELECT id, city, stadium, lat, lon, created_at, updated_at
FROM venues
WHERE id = ?;

-- name: CreateVenue :one
INSERT INTO venues (city, stadium, lat, lon)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateVenue :exec
UPDATE venues
SET city = ?, stadium = ?, lat = ?, lon = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: UpdateVenueCoords :exec
UPDATE venues
SET lat = ?, lon = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteVenue :exec
DELETE FROM venues WHERE id = ?;
