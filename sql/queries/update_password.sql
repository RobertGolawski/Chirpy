-- name: UpdatePassword :exec
UPDATE users
SET hashed_password = $1, updated_at = NOW()
WHERE id = $2;
