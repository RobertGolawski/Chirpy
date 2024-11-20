-- name: UpdatePassword :exec
UPDATE users
SET hashed_password = $1
WHERE id = $2;
