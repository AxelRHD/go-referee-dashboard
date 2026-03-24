-- name: ListTeams :many
SELECT id, name, state, is_active, remarks, created_at, updated_at
FROM teams
ORDER BY name;

-- name: GetTeam :one
SELECT id, name, state, is_active, remarks, created_at, updated_at
FROM teams
WHERE id = ?;

-- name: CreateTeam :one
INSERT INTO teams (name, state, is_active, remarks)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateTeam :exec
UPDATE teams
SET name = ?, state = ?, is_active = ?, remarks = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteTeam :exec
DELETE FROM teams WHERE id = ?;
