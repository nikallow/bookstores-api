-- name: CreateStore :one
INSERT INTO stores (name, address)
VALUES ($1, $2)
RETURNING *;

-- name: ListStores :many
SELECT *
FROM stores
WHERE deleted_at IS NULL
ORDER BY name;

-- name: GetStoreByUUID :one
SELECT *
FROM stores
WHERE uuid = $1
  AND deleted_at IS NULL;

-- name: UpdateStore :one
UPDATE stores
SET name       = $1,
    address    = $2,
    updated_at = now()
WHERE uuid = $3
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteStore :exec
UPDATE stores
SET deleted_at = now()
WHERE uuid = $1;
