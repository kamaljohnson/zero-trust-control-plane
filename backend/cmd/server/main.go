package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	authv1 "zero-trust-control-plane/backend/api/generated/auth/v1"
	healthv1 "zero-trust-control-plane/backend/api/generated/health/v1"
	"zero-trust-control-plane/backend/internal/config"
	"zero-trust-control-plane/backend/internal/db"
	devicerepo "zero-trust-control-plane/backend/internal/device/repository"
	identityrepo "zero-trust-control-plane/backend/internal/identity/repository"
	identityservice "zero-trust-control-plane/backend/internal/identity/service"
	membershiprepo "zero-trust-control-plane/backend/internal/membership/repository"
	"zero-trust-control-plane/backend/internal/security"
	sessionrepo "zero-trust-control-plane/backend/internal/session/repository"
	"zero-trust-control-plane/backend/internal/server"
	"zero-trust-control-plane/backend/internal/server/interceptors"
	userrepo "zero-trust-control-plane/backend/internal/user/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer lis.Close()

	var s *grpc.Server
	var tokens *security.TokenProvider
	deps := server.Deps{}

	authEnabled := cfg.DatabaseURL != "" && cfg.JWTPrivateKey != "" && cfg.JWTPublicKey != ""
	if authEnabled {
		database, err := db.Open(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("db: %v", err)
		}
		defer database.Close()

		hasher := security.NewHasher(cfg.BcryptCost)
		signer, err := security.ParsePrivateKey(cfg.JWTPrivateKey)
		if err != nil {
			log.Fatalf("jwt private key: %v", err)
		}
		pub, err := security.ParsePublicKey(cfg.JWTPublicKey)
		if err != nil {
			log.Fatalf("jwt public key: %v", err)
		}
		tokens = security.NewTokenProvider(signer, pub, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL(), cfg.RefreshTTL())

		userRepo := userrepo.NewPostgresRepository(database)
		identityRepo := identityrepo.NewPostgresRepository(database)
		sessionRepo := sessionrepo.NewPostgresRepository(database)
		deviceRepo := devicerepo.NewPostgresRepository(database)
		membershipRepo := membershiprepo.NewPostgresRepository(database)

		authService := identityservice.NewAuthService(
			userRepo,
			identityRepo,
			sessionRepo,
			deviceRepo,
			membershipRepo,
			hasher,
			tokens,
			cfg.AccessTTL(),
			cfg.RefreshTTL(),
		)
		deps.Auth = authService
	}

	if authEnabled {
		publicMethods := map[string]bool{
			authv1.AuthService_Register_FullMethodName:     true,
			authv1.AuthService_Login_FullMethodName:        true,
			authv1.AuthService_Refresh_FullMethodName:      true,
			healthv1.HealthService_HealthCheck_FullMethodName: true,
		}
		// tokens is in scope from authEnabled block
		s = grpc.NewServer(grpc.UnaryInterceptor(interceptors.AuthUnary(tokens, publicMethods)))
	} else {
		s = grpc.NewServer()
	}

	server.RegisterServices(s, deps)

	go func() {
		log.Printf("gRPC server listening on %s", cfg.GRPCAddr)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gRPC server...")
	s.GracefulStop()
	log.Println("gRPC server stopped")
}
