package refresh

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefreshRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*Refresh, error)
	RefreshRotateTokens(ctx context.Context, oldRefresh *Refresh, newRefreshString string, expiresAt time.Time) error
	ByPrimary(ctx context.Context, userID uuid.UUID, tokenHash string) (*Refresh, error)
}

type PostgresRefreshRepository struct {
	db *sql.DB
}

func NewPostgresRefreshRepository(db *sql.DB) *PostgresRefreshRepository {
	return &PostgresRefreshRepository{
		db: db,
	}
}

func (r *PostgresRefreshRepository) RefreshRotateTokens(ctx context.Context, oldRefresh *Refresh, newRefreshTokenHash string, expiresAt time.Time) error {
	//tx
	//1. revoke token
	//2. add new token
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//revoke
	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE refresh_tokens
		SET revoked =  true
		WHERE token_hash = $1
		`,
		oldRefresh.TokenHash,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("refresh token is invalid or already revoked")
	}

	//add new
	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		`,
		oldRefresh.UserID,
		newRefreshTokenHash,
		expiresAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRefreshRepository) ByPrimary(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash string,
) (*Refresh, error) {
	refresh := &Refresh{}

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT
			user_id,
			token_hash,
			created_at,
			expires_at,
			revoked
		FROM refresh_tokens
		WHERE user_id = $1
		  AND token_hash = $2
		  AND expires_at > NOW()
		  AND revoked = false
		`,
		userID,
		tokenHash,
	).Scan(
		&refresh.UserID,
		&refresh.TokenHash,
		&refresh.CreatedAt,
		&refresh.ExpiresAt,
		&refresh.Revoked,
	)
	if err != nil {
		return nil, err
	}

	return refresh, nil
}

func (r *PostgresRefreshRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*Refresh, error) {
	refresh := &Refresh{}

	err := r.db.QueryRowContext(
		ctx,
		`
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING user_id, token_hash, created_at, expires_at, revoked
		`,
		userID, tokenHash, expiresAt,
	).Scan(
		&refresh.UserID,
		&refresh.TokenHash,
		&refresh.CreatedAt,
		&refresh.ExpiresAt,
		&refresh.Revoked,
	)

	if err != nil {
		return nil, err
	}

	return refresh, nil
}
