package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	authv1 "zero-trust-control-plane/backend/api/generated/auth/v1"
	devv1 "zero-trust-control-plane/backend/api/generated/dev/v1"
	healthv1 "zero-trust-control-plane/backend/api/generated/health/v1"
	"zero-trust-control-plane/backend/internal/audit"
	auditrepo "zero-trust-control-plane/backend/internal/audit/repository"
	"zero-trust-control-plane/backend/internal/config"
	"zero-trust-control-plane/backend/internal/db"
	devicerepo "zero-trust-control-plane/backend/internal/device/repository"
	"zero-trust-control-plane/backend/internal/devotp"
	devotphandler "zero-trust-control-plane/backend/internal/devotp/handler"
	identityrepo "zero-trust-control-plane/backend/internal/identity/repository"
	identityservice "zero-trust-control-plane/backend/internal/identity/service"
	membershiprepo "zero-trust-control-plane/backend/internal/membership/repository"
	mfarepo "zero-trust-control-plane/backend/internal/mfa/repository"
	"zero-trust-control-plane/backend/internal/mfa/sms"
	mfaintentrepo "zero-trust-control-plane/backend/internal/mfaintent/repository"
	orgmfasettingsrepo "zero-trust-control-plane/backend/internal/orgmfasettings/repository"
	platformsettingsrepo "zero-trust-control-plane/backend/internal/platformsettings/repository"
	policyengine "zero-trust-control-plane/backend/internal/policy/engine"
	policyrepo "zero-trust-control-plane/backend/internal/policy/repository"
	"zero-trust-control-plane/backend/internal/security"
	"zero-trust-control-plane/backend/internal/server"
	"zero-trust-control-plane/backend/internal/server/interceptors"
	sessionrepo "zero-trust-control-plane/backend/internal/session/repository"
	telemetryproducer "zero-trust-control-plane/backend/internal/telemetry/producer"
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

	// Optional telemetry producer (Kafka). When brokers are set, emit telemetry after each RPC and from TelemetryService.
	var telemetryProducer telemetryproducer.Producer
	if brokers := cfg.TelemetryKafkaBrokersList(); len(brokers) > 0 {
		topic := cfg.TelemetryKafkaTopic
		if topic == "" {
			topic = "ztcp-telemetry"
		}
		if p, err := telemetryproducer.NewKafkaProducer(brokers, topic); err != nil {
			log.Printf("telemetry: kafka producer disabled: %v", err)
		} else if p != nil {
			telemetryProducer = p
			defer func() { _ = p.Close() }()
		}
	}
	deps.TelemetryProducer = telemetryProducer

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
		platformSettingsRepo := platformsettingsrepo.NewPostgresRepository(database)
		orgMFASettingsRepo := orgmfasettingsrepo.NewPostgresRepository(database)
		mfaChallengeRepo := mfarepo.NewPostgresRepository(database)
		mfaIntentRepo := mfaintentrepo.NewPostgresRepository(database)
		policyRepo := policyrepo.NewPostgresRepository(database)
		policyEvaluator := policyengine.NewOPAEvaluator(policyRepo)
		defaultTrustTTLDays := cfg.DefaultTrustTTLDays
		if defaultTrustTTLDays <= 0 {
			defaultTrustTTLDays = 30
		}
		var smsSender identityservice.OTPSender
		if cfg.SMSLocalAPIKey != "" {
			smsSender = sms.NewSMSLocalClient(cfg.SMSLocalAPIKey, cfg.SMSLocalBaseURL, cfg.SMSLocalSender)
		}
		var devOTPStore identityservice.DevOTPStore
		if cfg.OTPReturnToClient && cfg.Env != "production" {
			devStore := devotp.NewMemoryStore()
			devOTPStore = devStore
			deps.DevOTPHandler = devotphandler.NewServer(devStore)
		}
		auditRepo := auditrepo.NewPostgresRepository(database)
		deps.AuditRepo = auditRepo
		auditLogger := audit.NewLogger(auditRepo, interceptors.ClientIP)
		authService := identityservice.NewAuthService(
			userRepo,
			identityRepo,
			sessionRepo,
			deviceRepo,
			membershipRepo,
			platformSettingsRepo,
			orgMFASettingsRepo,
			mfaChallengeRepo,
			mfaIntentRepo,
			policyEvaluator,
			smsSender,
			hasher,
			tokens,
			cfg.AccessTTL(),
			cfg.RefreshTTL(),
			defaultTrustTTLDays,
			10*time.Minute,
			cfg.OTPReturnToClient,
			devOTPStore,
			auditLogger,
		)
		deps.Auth = authService
		deps.DeviceRepo = deviceRepo
		deps.PolicyRepo = policyRepo
		deps.HealthPinger = database
		deps.HealthPolicyChecker = policyEvaluator
	}

	if authEnabled {
		publicMethods := map[string]bool{
			authv1.AuthService_Register_FullMethodName:                 true,
			authv1.AuthService_Login_FullMethodName:                    true,
			authv1.AuthService_VerifyMFA_FullMethodName:                true,
			authv1.AuthService_SubmitPhoneAndRequestMFA_FullMethodName: true,
			authv1.AuthService_Refresh_FullMethodName:                  true,
			healthv1.HealthService_HealthCheck_FullMethodName:          true,
		}
		if deps.DevOTPHandler != nil {
			publicMethods[devv1.DevService_GetOTP_FullMethodName] = true
		}
		auditSkipMethods := map[string]bool{
			healthv1.HealthService_HealthCheck_FullMethodName: true,
		}
		telemetrySkipMethods := map[string]bool{
			healthv1.HealthService_HealthCheck_FullMethodName: true,
			devv1.DevService_GetOTP_FullMethodName:            true,
		}
		// tokens and deps.AuditRepo are in scope from authEnabled block
		s = grpc.NewServer(grpc.ChainUnaryInterceptor(
			interceptors.AuthUnary(tokens, publicMethods),
			interceptors.AuditUnary(deps.AuditRepo, auditSkipMethods),
			interceptors.TelemetryUnary(telemetryProducer, telemetrySkipMethods),
		))
	} else {
		telemetrySkipMethods := map[string]bool{
			healthv1.HealthService_HealthCheck_FullMethodName: true,
		}
		s = grpc.NewServer(grpc.ChainUnaryInterceptor(
			interceptors.TelemetryUnary(telemetryProducer, telemetrySkipMethods),
		))
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
