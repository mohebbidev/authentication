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

type SessionRepo struct {
	db *pgxpool.Pool
}

func NewSessionRepo(db *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, expires_at, revoked_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.ExpiresAt,
		session.RevokedAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", domain.MapPostgresError(err))
	}
	return nil
}

func (r *SessionRepo) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, revoked_at
		FROM sessions
		WHERE id = $1
	`
	var s domain.Session
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.UserID,
		&s.ExpiresAt,
		&s.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorsInstance.InvalidToken
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

func (r *SessionRepo) Revoke(ctx context.Context, id domain.SessionID, now domain.RevokeTime) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE id = $2 AND revoked_at IS NULL`
	_, err := r.db.Exec(ctx, query, time.Time(now), id)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *SessionRepo) RevokeAllForUser(ctx context.Context, userID domain.UserID, now domain.RevokeTime) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.db.Exec(ctx, query, time.Time(now), userID)
	if err != nil {
		return fmt.Errorf("revoke all sessions: %w", err)
	}
	return nil
}