-- name: GetAllPositions :many
SELECT position, long, sorter
FROM positions
ORDER BY sorter;
