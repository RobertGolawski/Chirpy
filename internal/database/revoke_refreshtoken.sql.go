// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: revoke_refreshtoken.sql

package database

import (
	"context"
)

const revokeRefreshToken = `-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW() 
WHERE token = $1
`

func (q *Queries) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx, revokeRefreshToken, token)
	return err
}
