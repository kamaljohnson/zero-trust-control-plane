package main

import (
	"context"
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
	organizationv1 "zero-trust-control-plane/backend/api/generated/organization/v1"
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
	organizationrepo "zero-trust-control-plane/backend/internal/organization/repository"
	orgmfasettingsrepo "zero-trust-control-plane/backend/internal/orgmfasettings/repository"
	orgpolicyconfigrepo "zero-trust-control-plane/backend/internal/orgpolicyconfig/repository"
	platformsettingsrepo "zero-trust-control-plane/backend/internal/platformsettings/repository"
	policyengine "zero-trust-control-plane/backend/internal/policy/engine"
	policyrepo "zero-trust-control-plane/backend/internal/policy/repository"
	"zero-trust-control-plane/backend/internal/security"
	"zero-trust-control-plane/backend/internal/server"
	"zero-trust-control-plane/backend/internal/server/interceptors"
	sessionrepo "zero-trust-control-plane/backend/internal/session/repository"
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
	if !authEnabled {
		var missing []string
		if cfg.DatabaseURL == "" {
			missing = append(missing, "DATABASE_URL")
		}
		if cfg.JWTPrivateKey == "" {
			missing = append(missing, "JWT_PRIVATE_KEY")
		}
		if cfg.JWTPublicKey == "" {
			missing = append(missing, "JWT_PUBLIC_KEY")
		}
		log.Printf("auth disabled: %v not set or empty; Register/Login/Refresh will return Unimplemented", missing)
	} else {
		log.Print("auth enabled: DATABASE_URL and JWT keys set; Register/Login/Refresh available")
	}
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
		orgRepo := organizationrepo.NewPostgresRepository(database)
		platformSettingsRepo := platformsettingsrepo.NewPostgresRepository(database)
		orgMFASettingsRepo := orgmfasettingsrepo.NewPostgresRepository(database)
		orgPolicyConfigRepo := orgpolicyconfigrepo.NewPostgresRepository(database)
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
		if cfg.OTPReturnToClient {
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
		deps.MembershipRepo = membershipRepo
		deps.SessionRepo = sessionRepo
		deps.UserRepo = userRepo
		deps.OrgRepo = orgRepo
		deps.AuditLogger = auditLogger
		deps.OrgPolicyConfigRepo = orgPolicyConfigRepo
		deps.OrgMFASettingsRepo = orgMFASettingsRepo
	}

	if authEnabled {
		publicMethods := map[string]bool{
			authv1.AuthService_Register_FullMethodName:                 true,
			authv1.AuthService_Login_FullMethodName:                    true,
			authv1.AuthService_VerifyMFA_FullMethodName:                true,
			authv1.AuthService_SubmitPhoneAndRequestMFA_FullMethodName: true,
			authv1.AuthService_Refresh_FullMethodName:                  true,
			authv1.AuthService_VerifyCredentials_FullMethodName:        true,
			healthv1.HealthService_HealthCheck_FullMethodName:          true,
			organizationv1.OrganizationService_CreateOrganization_FullMethodName: true,
		}
		if deps.DevOTPHandler != nil {
			publicMethods[devv1.DevService_GetOTP_FullMethodName] = true
		}
		auditSkipMethods := map[string]bool{
			healthv1.HealthService_HealthCheck_FullMethodName: true,
		}
		var sessionValidator interceptors.SessionValidator
		if deps.SessionRepo != nil {
			sessionValidator = func(ctx context.Context, sessionID string) (bool, error) {
				sess, err := deps.SessionRepo.GetByID(ctx, sessionID)
				if err != nil {
					return false, err
				}
				return sess != nil && sess.RevokedAt == nil, nil
			}
		}
		s = grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				interceptors.AuthUnary(tokens, publicMethods, sessionValidator),
				interceptors.AuditUnary(deps.AuditRepo, auditSkipMethods),
			),
		)
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
