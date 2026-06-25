package application

import (
	"authpractice/internal/domain"
	"authpractice/internal/domain/repositories"
	"context"
)

type LogoutInput struct {
	SessionID domain.SessionID
}

type LogoutUseCase struct {
	sessionRepo repositories.SessionRepository
}

func NewLogoutUseCase(sessionRepo repositories.SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{sessionRepo: sessionRepo}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	return uc.sessionRepo.Revoke(ctx, input.SessionID, domain.Now())
}