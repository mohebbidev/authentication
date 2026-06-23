package postgres

import (
	"authpractice/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepo struct {
	db *pgxpool.Pool
}

func NewTokenRepo(db *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{db: db}
}

// ── Verification tokens ──────────────────────────────────────────────────────

func (r *TokenRepo) CreateVerification(ctx context.Context, token *domain.VerificationToken) error {
	query := `
		INSERT INTO verification_tokens (hash, user_id, expires_at, revoked_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query,
		token.Hash, token.UserID, token.ExpiresAt, token.RevokedAt,
	)
	if err != nil {
		return fmt.Errorf("create verification token: %w", domain.MapPostgresError(err))
	}
	return nil
}

func (r *TokenRepo) GetVerification(ctx context.Context, hash domain.TokenHash) (*domain.VerificationToken, error) {
	query := `
		SELECT hash, user_id, expires_at, revoked_at
		FROM verification_tokens
		WHERE hash = $1
	`
	var t domain.VerificationToken
	err := r.db.QueryRow(ctx, query, hash).Scan(
		&t.Hash, &t.UserID, &t.ExpiresAt, &t.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorsInstance.InvalidToken
		}
		return nil, fmt.Errorf("get verification token: %w", err)
	}
	return &t, nil
}

func (r *TokenRepo) RevokeVerification(ctx context.Context, hash domain.TokenHash, now domain.RevokeTime) error {
	query := `UPDATE verification_tokens SET revoked_at = $1 WHERE hash = $2 AND revoked_at IS NULL`
	_, err := r.db.Exec(ctx, query, time.Time(now), hash)
	if err != nil {
		return fmt.Errorf("revoke verification token: %w", err)
	}
	return nil
}

// ── Password reset tokens ────────────────────────────────────────────────────

func (r *TokenRepo) CreatePasswordReset(ctx context.Context, token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (hash, user_id, expires_at, revoked_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query,
		token.Hash, token.UserID, token.ExpiresAt, token.RevokedAt,
	)
	if err != nil {
		return fmt.Errorf("create password reset token: %w", domain.MapPostgresError(err))
	}
	return nil
}

func (r *TokenRepo) GetPasswordReset(ctx context.Context, hash domain.TokenHash) (*domain.PasswordResetToken, error) {
	query := `
		SELECT hash, user_id, expires_at, revoked_at
		FROM password_reset_tokens
		WHERE hash = $1
	`
	var t domain.PasswordResetToken
	err := r.db.QueryRow(ctx, query, hash).Scan(
		&t.Hash, &t.UserID, &t.ExpiresAt, &t.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorsInstance.InvalidToken
		}
		return nil, fmt.Errorf("get password reset token: %w", err)
	}
	return &t, nil
}

func (r *TokenRepo) RevokePasswordReset(ctx context.Context, hash domain.TokenHash, now domain.RevokeTime) error {
	query := `UPDATE password_reset_tokens SET revoked_at = $1 WHERE hash = $2 AND revoked_at IS NULL`
	_, err := r.db.Exec(ctx, query, time.Time(now), hash)
	if err != nil {
		return fmt.Errorf("revoke password reset token: %w", err)
	}
	return nil
}