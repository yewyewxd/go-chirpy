-- name: CreateChirp :one
INSERT INTO
    chirps (
        id,
        created_at,
        updated_at,
        body,
        user_id
    )
VALUES (
        gen_random_uuid (),
        NOW(),
        NOW(),
        $1,
        $2
    ) RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps ORDER BY
  CASE WHEN sqlc.arg(sort_desc)::bool THEN created_at END DESC,
  CASE WHEN NOT sqlc.arg(sort_desc)::bool THEN created_at END ASC;

-- name: GetChirpsByUser :many
SELECT *
FROM chirps
WHERE user_id = $1
ORDER BY
  CASE WHEN sqlc.arg(sort_desc)::bool THEN created_at END DESC,
  CASE WHEN NOT sqlc.arg(sort_desc)::bool THEN created_at END ASC;

-- name: GetChirpById :one
SELECT * FROM chirps WHERE id = $1;

-- name: DeleteChirpForUser :exec
DELETE FROM chirps WHERE id = $1 AND user_id = $2;