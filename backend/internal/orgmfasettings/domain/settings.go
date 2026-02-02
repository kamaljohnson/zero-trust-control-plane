package domain

import "time"

// OrgMFASettings holds org-level MFA/device trust settings (one row per org).
type OrgMFASettings struct {
	OrgID                   string
	MFARequiredForNewDevice bool
	MFARequiredForUntrusted bool
	MFARequiredAlways       bool
	RegisterTrustAfterMFA   bool
	TrustTTLDays            int
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
