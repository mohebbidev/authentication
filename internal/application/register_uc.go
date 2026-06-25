package application

import (
	"authpractice/internal/domain"
	"context"
	"time"
)

type RegisterInput struct {
	FullName    string
	DateOfBirth time.Time
	Email       string
	Password    string
}

type RegisterOutput struct {
	UserID *domain.UserID
}

type RegisterUseCase struct {
	userRepo UserRepository
	hasher   PasswordHasher
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	email, err := domain.ValidateEmail(input.Email)

	if err != nil {
		return nil, err
	}

	if len(input.Password) < 8 {
		return nil, domain.ErrorsInstance.WeakPassword
	}

	existing, err := uc.userRepo.GetByEmail(ctx, email.String())
	if err != nil {
		_ = existing
	} else if existing != nil {
		return nil, domain.ErrorsInstance.UserAlreadyExists
	}

	hash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		FullName:     input.FullName,
		Email:        email.String(),
		HashPassword: hash,
		DateOfBirth:  input.DateOfBirth,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		UserID: (*domain.UserID)(&user.ID),
	}, nil
}
