
package main

import (
	"authpractice/internal/application"
	"authpractice/internal/infrastructure/config"
	"authpractice/internal/infrastructure/database"
	"authpractice/internal/infrastructure/postgres"
	"authpractice/internal/infrastructure/security"
	grpchandler "authpractice/internal/transport/grpc"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()

	// ── Config ───────────────────────────────────────────────────────────────
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ── Database ─────────────────────────────────────────────────────────────
	dsn := database.BuildDSN(cfg.DB)
	db, err := database.ConnectDB(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(ctx, db, "file://migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo    := postgres.NewUserRepo(db)
	sessionRepo := postgres.NewSessionRepo(db)
	tokenRepo   := postgres.NewTokenRepo(db)

	// ── Infrastructure services ───────────────────────────────────────────────
	hasher   := security.NewBcryptHasher(12)
	tokenGen := security.NewJWTGenerator(cfg.JWT.Secret, time.Duration(cfg.JWT.AccessTokenTTLMinutes)*time.Minute)

	// ── Use cases ─────────────────────────────────────────────────────────────
	registerUC := application.NewRegisterUseCase(userRepo, hasher)
	loginUC    := application.NewLoginUseCase(userRepo, sessionRepo, hasher, tokenGen,
		time.Duration(cfg.JWT.SessionTTLDays)*24*time.Hour)
	refreshUC  := application.NewRefreshTokenUseCase(sessionRepo, tokenGen)
	logoutUC   := application.NewLogoutUseCase(sessionRepo)
	verifyUC   := application.NewVerifyEmailUseCase(tokenRepo, userRepo)
	reqResetUC := application.NewRequestPasswordResetUseCase(userRepo, tokenRepo, tokenGen,
		15*time.Minute)
	resetUC    := application.NewResetPasswordUseCase(tokenRepo, userRepo, sessionRepo, hasher)

	// ── gRPC server ───────────────────────────────────────────────────────────
	handler := grpchandler.NewHandler(
		registerUC,
		loginUC,
		refreshUC,
		logoutUC,
		verifyUC,
		reqResetUC,
		resetUC,
	)

	srv := grpchandler.NewServer(handler, cfg.Server.Port)

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Run(); err != nil {
			slog.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	srv.GracefulStop()
	slog.Info("shutdown complete")
}