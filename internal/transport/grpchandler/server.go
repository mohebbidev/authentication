package grpchandler

import (
	"fmt"
	"log/slog"
	"net"

	authv1 "authpractice/proto/auth"

	"google.golang.org/grpc"
)

type Server struct {
	grpc *grpc.Server
	port string
}

func NewServer(handler *Handler, port string) *Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			RecoveryInterceptor,
			LoggingInterceptor,
		),
	)

	authv1.RegisterAuthServiceServer(srv, handler)

	return &Server{grpc: srv, port: port}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	slog.Info("gRPC server listening", "port", s.port)
	return s.grpc.Serve(lis)
}

func (s *Server) GracefulStop() {
	slog.Info("gRPC server shutting down")
	s.grpc.GracefulStop()
}