-- name: UpdateEmail :exec
UPDATE users
SET email = $1, updated_at = NOW()
WHERE id = $2;
