package repositories

import (
	"authpractice/internal/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	UpdatePassword(ctx context.Context, id domain.UserID, newHash string) error
	SetActive(ctx context.Context, id domain.UserID, active bool) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error)
	Revoke(ctx context.Context, id domain.SessionID, now domain.RevokeTime) error
	RevokeAllForUser(ctx context.Context, userID domain.UserID, now domain.RevokeTime) error
}

type TokenRepository interface {
	CreateVerification(ctx context.Context, token *domain.VerificationToken) error
	GetVerification(ctx context.Context, hash domain.TokenHash) (*domain.VerificationToken, error)
	RevokeVerification(ctx context.Context, hash domain.TokenHash, now domain.RevokeTime) error

	CreatePasswordReset(ctx context.Context, token *domain.PasswordResetToken) error
	GetPasswordReset(ctx context.Context, hash domain.TokenHash) (*domain.PasswordResetToken, error)
	RevokePasswordReset(ctx context.Context, hash domain.TokenHash, now domain.RevokeTime) error
}

type PasswordHasher interface {
	Hash(pwd string) (string, error)
	Compare(hash, pwd string) error
}

type TokenGenerator interface {
	// GenerateAccessToken signs a short-lived JWT and returns the signed string.
	GenerateAccessToken(userID domain.UserID) (string, error)
	// GenerateOpaqueToken creates a cryptographically random token and returns
	// (rawToken, hash). Store the hash; send the raw token to the user.
	GenerateOpaqueToken() (raw string, hash domain.TokenHash, err error)
}