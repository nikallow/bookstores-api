-- name: CreateBook :one
INSERT INTO books (isbn, title, author, description, page_count, publication_year)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (isbn)
WHERE isbn IS NOT NULL DO
UPDATE
SET title      = EXCLUDED.title,
    author     = EXCLUDED.author,
    updated_at = now()
RETURNING *;

-- name: ListBooks :many
SELECT *
FROM books
WHERE deleted_at IS NULL
ORDER BY title;

-- name: GetBookByID :one
SELECT *
FROM books
WHERE id = $1
  AND deleted_at IS NULL;

-- name: SearchBooks :many
SELECT *
FROM books
WHERE (title ILIKE '%' || sqlc.arg(query) || '%' OR author ILIKE '%' || sqlc.arg(query) || '%')
  AND deleted_at IS NULL
LIMIT 10;
