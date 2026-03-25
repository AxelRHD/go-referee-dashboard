-- name: ListLeagues :many
SELECT id, name, short_name, sorter, remarks, created_at, updated_at
FROM leagues
ORDER BY sorter, name;

-- name: GetLeague :one
SELECT id, name, short_name, sorter, remarks, created_at, updated_at
FROM leagues
WHERE id = ?;

-- name: CreateLeague :one
INSERT INTO leagues (name, short_name, sorter, remarks, created_at, updated_at)
VALUES (?, ?, ?, ?, datetime('now', 'localtime'), datetime('now', 'localtime'))
RETURNING *;

-- name: UpdateLeague :exec
UPDATE leagues
SET name = ?, short_name = ?, sorter = ?, remarks = ?, updated_at = datetime('now', 'localtime')
WHERE id = ?;

-- name: DeleteLeague :exec
DELETE FROM leagues WHERE id = ?;
