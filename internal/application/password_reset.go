package application

import (
	"authpractice/internal/domain"
	"authpractice/internal/domain/repositories"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// hashToken is the single place we convert a raw opaque token → its stored hash.
// Both verify_email.go and this file import it from the same package.
func hashToken(raw string) domain.TokenHash {
	sum := sha256.Sum256([]byte(raw))
	return domain.TokenHash(hex.EncodeToString(sum[:]))
}

// ── Request password reset ───────────────────────────────────────────────────

type RequestPasswordResetInput struct {
	Email string
}

type RequestPasswordResetOutput struct {
	// RawToken is sent to the user via email.
	// The caller (HTTP handler / email service) is responsible for delivery.
	RawToken string
}

type RequestPasswordResetUseCase struct {
	userRepo  repositories.UserRepository
	tokenRepo repositories.TokenRepository
	tokenGen  repositories.TokenGenerator
	tokenTTL  time.Duration
}

func NewRequestPasswordResetUseCase(
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepository,
	tokenGen repositories.TokenGenerator,
	tokenTTL time.Duration,
) *RequestPasswordResetUseCase {
	return &RequestPasswordResetUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		tokenGen:  tokenGen,
		tokenTTL:  tokenTTL,
	}
}

func (uc *RequestPasswordResetUseCase) Execute(ctx context.Context, input RequestPasswordResetInput) (*RequestPasswordResetOutput, error) {
	email, err := domain.ValidateEmail(input.Email)
	if err != nil {
		// Return success anyway — never reveal whether the email exists.
		return &RequestPasswordResetOutput{}, nil
	}

	user, err := uc.userRepo.GetByEmail(ctx, email.String())
	if err != nil {
		// Same: silent success to prevent user enumeration.
		return &RequestPasswordResetOutput{}, nil
	}

	raw, hash, err := uc.tokenGen.GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}

	token := &domain.PasswordResetToken{
		Token:  *domain.NewToken(hash, time.Now().UTC().Add(uc.tokenTTL)),
		UserID: user.ID,
	}

	if err := uc.tokenRepo.CreatePasswordReset(ctx, token); err != nil {
		return nil, err
	}

	return &RequestPasswordResetOutput{RawToken: raw}, nil
}

// ── Reset password ───────────────────────────────────────────────────────────

type ResetPasswordInput struct {
	RawToken    string
	NewPassword string
}

type ResetPasswordUseCase struct {
	tokenRepo   repositories.TokenRepository
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	hasher      repositories.PasswordHasher
}

func NewResetPasswordUseCase(
	tokenRepo repositories.TokenRepository,
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	hasher repositories.PasswordHasher,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		tokenRepo:   tokenRepo,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
	}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, input ResetPasswordInput) error {
	if len(input.NewPassword) < 8 {
		return domain.ErrorsInstance.WeakPassword
	}

	hash := hashToken(input.RawToken)

	token, err := uc.tokenRepo.GetPasswordReset(ctx, hash)
	if err != nil {
		return domain.ErrorsInstance.InvalidToken
	}

	now := time.Now().UTC()

	if token.IsExpired(now) {
		return domain.ErrorsInstance.TokenExpired
	}
	if token.IsRevoked() {
		return domain.ErrorsInstance.TokenRevoked
	}

	newHash, err := uc.hasher.Hash(input.NewPassword)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePassword(ctx, token.UserID, newHash); err != nil {
		return err
	}

	// Revoke the reset token so it can't be reused.
	if err := uc.tokenRepo.RevokePasswordReset(ctx, hash, domain.RevokeTime(now)); err != nil {
		return err
	}

	// Invalidate all active sessions — password changed, force re-login everywhere.
	if err := uc.sessionRepo.RevokeAllForUser(ctx, token.UserID, domain.RevokeTime(now)); err != nil {
		return err
	}

	return nil
}