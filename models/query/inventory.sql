-- name: CreateInventory :one
INSERT INTO inventory (
    name, quantity, category, located
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: DeleteInventory :exec
DELETE FROM inventory
WHERE id = $1;