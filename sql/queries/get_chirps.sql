-- name: GetChirps :many
SELECT *
FROM chirps
ORDER BY 
    CASE WHEN sqlc.arg('sort_order') = 'desc' THEN created_at END DESC,
    CASE WHEN sqlc.arg('sort_order') = 'asc' THEN created_at END ASC;
