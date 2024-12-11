-- name: RefreshJWT :exec
UPDATE refresh_tokens
SET token = $1
WHERE user_id = $2;
