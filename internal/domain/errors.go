package domain

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type Errors struct {
	InvalidCredentials error
	UserAlreadyExists  error
	WeakPassword       error
	UserNotFound       error
	InvalidToken       error
	TokenExpired       error
	TokenRevoked       error
	SessionExpired     error
	SessionRevoked     error
	AccountInactive    error
}

var ErrorsInstance = Errors{
	InvalidCredentials: errors.New("invalid credentials"),
	UserAlreadyExists:  errors.New("user already exists"),
	WeakPassword:       errors.New("password must be at least 8 characters"),
	UserNotFound:       errors.New("user not found"),
	InvalidToken:       errors.New("invalid token"),
	TokenExpired:       errors.New("token has expired"),
	TokenRevoked:       errors.New("token has been revoked"),
	SessionExpired:     errors.New("session has expired"),
	SessionRevoked:     errors.New("session has been revoked"),
	AccountInactive:    errors.New("account is not active"),
}

func MapPostgresError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23505":
			if pgError.ConstraintName == "users_email_unique" {
				return ErrorsInstance.UserAlreadyExists
			}
			return fmt.Errorf("unique constraint violation: %w", err)
		case "40001", "40P01":
			return fmt.Errorf("transaction conflict, retry: %w", err)
		}
	}
	return fmt.Errorf("db error: %w", err)
}