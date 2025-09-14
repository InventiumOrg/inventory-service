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

-- name: GetInventory :one
SELECT * FROM inventory
WHERE id = $1;

-- name: UpdateInventory :one
UPDATE inventory
SET name = $2,
    unit = $3,
    quantity = $4,
    measure = $5,
    category = $6,
    location = $7
WHERE id = $1
RETURNING *;