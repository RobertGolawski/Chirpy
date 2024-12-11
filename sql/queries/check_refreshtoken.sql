-- name: GetTokenData :one
SELECT *
FROM refresh_tokens
WHERE token = $1;