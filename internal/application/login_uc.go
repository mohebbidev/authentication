package application

import (
	"authpractice/internal/domain"
	"authpractice/internal/domain/repositories"
	"context"
	"time"

	"github.com/google/uuid"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken  string
	SessionID    domain.SessionID
	SessionExpiry time.Time
}

type LoginUseCase struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	hasher      repositories.PasswordHasher
	tokenGen    repositories.TokenGenerator
	sessionTTL  time.Duration
}

func NewLoginUseCase(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	hasher repositories.PasswordHasher,
	tokenGen repositories.TokenGenerator,
	sessionTTL time.Duration,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		tokenGen:    tokenGen,
		sessionTTL:  sessionTTL,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	email, err := domain.ValidateEmail(input.Email)
	if err != nil {
		return nil, domain.ErrorsInstance.InvalidCredentials
	}

	user, err := uc.userRepo.GetByEmail(ctx, email.String())
	if err != nil {
		// don't leak whether the email exists
		return nil, domain.ErrorsInstance.InvalidCredentials
	}

	if !user.Active {
		return nil, domain.ErrorsInstance.AccountInactive
	}

	if err := uc.hasher.Compare(user.HashPassword, input.Password); err != nil {
		return nil, domain.ErrorsInstance.InvalidCredentials
	}

	accessToken, err := uc.tokenGen.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	session := domain.NewSession(
		domain.SessionID(uuid.NewString()),
		user.ID,
		now.Add(uc.sessionTTL),
	)

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &LoginOutput{
		AccessToken:   accessToken,
		SessionID:     session.ID,
		SessionExpiry: session.ExpiresAt,
	}, nil
}