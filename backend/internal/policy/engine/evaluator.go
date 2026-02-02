package engine

import (
	"context"

	devicedomain "zero-trust-control-plane/backend/internal/device/domain"
	orgmfasettingsdomain "zero-trust-control-plane/backend/internal/orgmfasettings/domain"
	platformdomain "zero-trust-control-plane/backend/internal/platformsettings/domain"
	userdomain "zero-trust-control-plane/backend/internal/user/domain"
)

// MFAResult holds the result of device-trust/MFA policy evaluation.
type MFAResult struct {
	MFARequired           bool
	RegisterTrustAfterMFA bool
	TrustTTLDays          int
}

// Evaluator evaluates device-trust/MFA policies using OPA or other engines.
type Evaluator interface {
	// EvaluateMFA evaluates platform and org device-trust/MFA policy for the given device and context.
	// Returns whether MFA is required, whether to register device as trusted after successful MFA, and trust TTL in days.
	EvaluateMFA(
		ctx context.Context,
		platformSettings *platformdomain.PlatformDeviceTrustSettings,
		orgSettings *orgmfasettingsdomain.OrgMFASettings,
		device *devicedomain.Device,
		user *userdomain.User,
		isNewDevice bool,
	) (MFAResult, error)
}
