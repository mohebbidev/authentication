package security

import (
	"authpractice/internal/domain"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(pwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), h.cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

func (h *BcryptHasher) Compare(hash, pwd string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
	if err != nil {
		return domain.ErrorsInstance.InvalidCredentials
	}
	return nil
}