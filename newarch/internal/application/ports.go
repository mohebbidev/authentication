package application

import (
	"authpractice/internal/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type PasswordHasher interface {
	Hash(pwd string) (string, error)
}