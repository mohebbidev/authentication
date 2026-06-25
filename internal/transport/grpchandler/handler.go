package grpchandler

import (
	"authpractice/internal/application"
	"authpractice/internal/domain"
	auth "authpractice/proto/auth"
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler holds all use cases and implements auth.AuthServiceServer.
type Handler struct {
	auth.UnimplementedAuthServiceServer

	register             *application.RegisterUseCase
	login                *application.LoginUseCase
	refreshToken         *application.RefreshTokenUseCase
	logout               *application.LogoutUseCase
	verifyEmail          *application.VerifyEmailUseCase
	requestPasswordReset *application.RequestPasswordResetUseCase
	resetPassword        *application.ResetPasswordUseCase
}

func NewHandler(
	register *application.RegisterUseCase,
	login *application.LoginUseCase,
	refreshToken *application.RefreshTokenUseCase,
	logout *application.LogoutUseCase,
	verifyEmail *application.VerifyEmailUseCase,
	requestPasswordReset *application.RequestPasswordResetUseCase,
	resetPassword *application.ResetPasswordUseCase,
) *Handler {
	return &Handler{
		register:             register,
		login:                login,
		refreshToken:         refreshToken,
		logout:               logout,
		verifyEmail:          verifyEmail,
		requestPasswordReset: requestPasswordReset,
		resetPassword:        resetPassword,
	}
}

func (h *Handler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	if req.FullName == "" || req.Email == "" || req.Password == "" || req.DateOfBirth == "" {
		return nil, status.Error(codes.InvalidArgument, "full_name, email, password and date_of_birth are required")
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "date_of_birth must be ISO 8601: YYYY-MM-DD")
	}

	out, err := h.register.Execute(ctx, application.RegisterInput{
		FullName:    req.FullName,
		Email:       req.Email,
		Password:    req.Password,
		DateOfBirth: dob,
	})
	if err != nil {
		return nil, domainToGRPC(err)
	}

	return &auth.RegisterResponse{UserId: string(out.UserID)}, nil
}

func (h *Handler) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	out, err := h.login.Execute(ctx, application.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, domainToGRPC(err)
	}

	return &auth.LoginResponse{
		AccessToken:   out.AccessToken,
		SessionId:     string(out.SessionID),
		SessionExpiry: out.SessionExpiry.Unix(),
	}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	out, err := h.refreshToken.Execute(ctx, application.RefreshTokenInput{
		SessionID: domain.SessionID(req.SessionId),
	})
	if err != nil {
		return nil, domainToGRPC(err)
	}

	return &auth.RefreshTokenResponse{AccessToken: out.AccessToken}, nil
}

func (h *Handler) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	if err := h.logout.Execute(ctx, application.LogoutInput{
		SessionID: domain.SessionID(req.SessionId),
	}); err != nil {
		return nil, domainToGRPC(err)
	}

	return &auth.LogoutResponse{}, nil
}

func (h *Handler) VerifyEmail(ctx context.Context, req *auth.VerifyEmailRequest) (*auth.VerifyEmailResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if err := h.verifyEmail.Execute(ctx, application.VerifyEmailInput{
		RawToken: req.Token,
	}); err != nil {
		return nil, domainToGRPC(err)
	}

	return &auth.VerifyEmailResponse{}, nil
}

func (h *Handler) RequestPasswordReset(ctx context.Context, req *auth.RequestPasswordResetRequest) (*auth.RequestPasswordResetResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	out, err := h.requestPasswordReset.Execute(ctx, application.RequestPasswordResetInput{
		Email: req.Email,
	})
	if err != nil {
		return nil, domainToGRPC(err)
	}

	return &auth.RequestPasswordResetResponse{RawToken: out.RawToken}, nil
}

func (h *Handler) ResetPassword(ctx context.Context, req *auth.ResetPasswordRequest) (*auth.ResetPasswordResponse, error) {
	if req.Token == "" || req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "token and new_password are required")
	}

	if err := h.resetPassword.Execute(ctx, application.ResetPasswordInput{
		RawToken:    req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		return nil, domainToGRPC(err)
	}

	return &auth.ResetPasswordResponse{}, nil
}