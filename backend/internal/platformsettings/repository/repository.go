package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/platformsettings/domain"
)

// Repository defines read access to platform settings for device trust / MFA.
type Repository interface {
	// GetDeviceTrustSettings returns platform-level MFA/device trust settings.
	// Uses defaults when keys are missing (MFARequiredAlways false, DefaultTrustTTLDays from config).
	GetDeviceTrustSettings(ctx context.Context, defaultTrustTTLDays int) (*domain.PlatformDeviceTrustSettings, error)
}
