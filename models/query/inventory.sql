-- name: CreateInventory :one
INSERT INTO inventory (
    name, unit, quantity, measure, category, location
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListInventory :many
SELECT id, name, unit, quantity, category, measure, location, created_at
FROM inventory
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteInventory :exec
DELETE FROM inventory
WHERE id = $1;