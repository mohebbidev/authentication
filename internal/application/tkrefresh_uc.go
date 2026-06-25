package application

import (
	"authpractice/internal/domain"
	"authpractice/internal/domain/repositories"
	"context"
	"time"
)

type RefreshTokenInput struct {
	SessionID domain.SessionID
}

type RefreshTokenOutput struct {
	AccessToken string
}

type RefreshTokenUseCase struct {
	sessionRepo repositories.SessionRepository
	tokenGen    repositories.TokenGenerator
}

func NewRefreshTokenUseCase(
	sessionRepo repositories.SessionRepository,
	tokenGen repositories.TokenGenerator,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{sessionRepo: sessionRepo, tokenGen: tokenGen}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error) {
	session, err := uc.sessionRepo.GetByID(ctx, input.SessionID)
	if err != nil {
		return nil, domain.ErrorsInstance.InvalidToken
	}

	now := time.Now().UTC()

	if session.IsExpired(now) {
		return nil, domain.ErrorsInstance.SessionExpired
	}
	if session.IsRevoked() {
		return nil, domain.ErrorsInstance.SessionRevoked
	}

	accessToken, err := uc.tokenGen.GenerateAccessToken(session.UserID)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenOutput{AccessToken: accessToken}, nil
}