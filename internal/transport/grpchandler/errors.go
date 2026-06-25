package grpchandler

import (
	"authpractice/internal/domain"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func domainToGRPC(err error) error {
	e := domain.ErrorsInstance

	switch {
	case errors.Is(err, e.InvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, e.UserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, e.WeakPassword):
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	case errors.Is(err, e.UserNotFound):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, e.InvalidToken):
		return status.Error(codes.Unauthenticated, "invalid or expired token")
	case errors.Is(err, e.TokenExpired):
		return status.Error(codes.Unauthenticated, "token has expired")
	case errors.Is(err, e.TokenRevoked):
		return status.Error(codes.Unauthenticated, "token has been revoked")
	case errors.Is(err, e.SessionExpired):
		return status.Error(codes.Unauthenticated, "session has expired")
	case errors.Is(err, e.SessionRevoked):
		return status.Error(codes.Unauthenticated, "session has been revoked")
	case errors.Is(err, e.AccountInactive):
		return status.Error(codes.PermissionDenied, "account is not active — verify your email first")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}