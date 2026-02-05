-- name: CreateSKU :one
INSERT INTO skus (book_id, store_id, price_in_kopeks, stock_count)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSKUByUUID :one
SELECT sqlc.embed(s), sqlc.embed(b)
FROM skus s
         JOIN books b ON s.book_id = b.id
WHERE s.uuid = $1
  AND s.deleted_at IS NULL;

-- name: ListSKUsByStore :many
SELECT sqlc.embed(s), sqlc.embed(b)
FROM skus s
         JOIN books b ON s.book_id = b.id
WHERE s.store_id = $1
  AND s.deleted_at IS NULL;

-- name: ListBookAvailability :many
SELECT sqlc.embed(s), sqlc.embed(st)
FROM skus s
         JOIN stores st ON s.store_id = st.id
WHERE s.book_id = $1
  AND s.deleted_at IS NULL
  AND st.deleted_at IS NULL;

-- name: UpdateSKUPrice :one
UPDATE skus
SET price_in_kopeks = $2,
    updated_at      = now()
WHERE uuid = $1
RETURNING *;

-- name: AdjustSKUStock :one
UPDATE skus
SET stock_count = stock_count + sqlc.arg(change_by),
    updated_at  = now()
WHERE uuid = $1
RETURNING *;
