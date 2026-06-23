package application

import (
	"authpractice/internal/domain"
	"authpractice/internal/domain/repositories"
	"context"
	"time"
)

type VerifyEmailInput struct {
	RawToken string
}

type VerifyEmailUseCase struct {
	tokenRepo repositories.TokenRepository
	userRepo  repositories.UserRepository
}

func NewVerifyEmailUseCase(
	tokenRepo repositories.TokenRepository,
	userRepo repositories.UserRepository,
) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{tokenRepo: tokenRepo, userRepo: userRepo}
}

func (uc *VerifyEmailUseCase) Execute(ctx context.Context, input VerifyEmailInput) error {
	hash := hashToken(input.RawToken)

	token, err := uc.tokenRepo.GetVerification(ctx, hash)
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

	if err := uc.userRepo.SetActive(ctx, token.UserID, true); err != nil {
		return err
	}

	if err := uc.tokenRepo.RevokeVerification(ctx, hash, domain.RevokeTime(now)); err != nil {
		return err
	}

	return nil
}