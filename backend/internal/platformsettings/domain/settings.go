package domain

// PlatformDeviceTrustSettings holds platform-level MFA/device trust settings (from platform_settings table or defaults).
type PlatformDeviceTrustSettings struct {
	MFARequiredAlways   bool
	DefaultTrustTTLDays int
}
