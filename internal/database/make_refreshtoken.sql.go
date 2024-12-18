// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: make_refreshtoken.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    NULL
)
`

type CreateRefreshTokenParams struct {
	Token     string
	UserID    uuid.NullUUID
	ExpiresAt time.Time
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) error {
	_, err := q.db.ExecContext(ctx, createRefreshToken, arg.Token, arg.UserID, arg.ExpiresAt)
	return err
}
