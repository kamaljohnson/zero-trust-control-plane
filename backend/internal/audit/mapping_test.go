package audit

import (
	"testing"
)

func TestParseFullMethod_GetUser(t *testing.T) {
	fullMethod := "/ztcp.user.v1.UserService/GetUser"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "get" {
		t.Errorf("action = %q, want %q", ar.Action, "get")
	}
	if ar.Resource != "user" {
		t.Errorf("resource = %q, want %q", ar.Resource, "user")
	}
}

func TestParseFullMethod_ListUsers(t *testing.T) {
	fullMethod := "/ztcp.user.v1.UserService/ListUsers"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "list" {
		t.Errorf("action = %q, want %q", ar.Action, "list")
	}
	if ar.Resource != "user" {
		t.Errorf("resource = %q, want %q", ar.Resource, "user")
	}
}

func TestParseFullMethod_CreatePolicy(t *testing.T) {
	fullMethod := "/ztcp.policy.v1.PolicyService/CreatePolicy"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "create" {
		t.Errorf("action = %q, want %q", ar.Action, "create")
	}
	if ar.Resource != "policy" {
		t.Errorf("resource = %q, want %q", ar.Resource, "policy")
	}
}

func TestParseFullMethod_UpdatePolicy(t *testing.T) {
	fullMethod := "/ztcp.policy.v1.PolicyService/UpdatePolicy"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "update" {
		t.Errorf("action = %q, want %q", ar.Action, "update")
	}
	if ar.Resource != "policy" {
		t.Errorf("resource = %q, want %q", ar.Resource, "policy")
	}
}

func TestParseFullMethod_DeletePolicy(t *testing.T) {
	fullMethod := "/ztcp.policy.v1.PolicyService/DeletePolicy"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "delete" {
		t.Errorf("action = %q, want %q", ar.Action, "delete")
	}
	if ar.Resource != "policy" {
		t.Errorf("resource = %q, want %q", ar.Resource, "policy")
	}
}

func TestParseFullMethod_AddMember(t *testing.T) {
	fullMethod := "/ztcp.membership.v1.MembershipService/AddMember"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "user_added" {
		t.Errorf("action = %q, want %q", ar.Action, "user_added")
	}
	if ar.Resource != "user" {
		t.Errorf("resource = %q, want %q", ar.Resource, "user")
	}
}

func TestParseFullMethod_RemoveMember(t *testing.T) {
	fullMethod := "/ztcp.membership.v1.MembershipService/RemoveMember"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "user_removed" {
		t.Errorf("action = %q, want %q", ar.Action, "user_removed")
	}
	if ar.Resource != "user" {
		t.Errorf("resource = %q, want %q", ar.Resource, "user")
	}
}

func TestParseFullMethod_UpdateRole(t *testing.T) {
	fullMethod := "/ztcp.membership.v1.MembershipService/UpdateRole"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "role_changed" {
		t.Errorf("action = %q, want %q", ar.Action, "role_changed")
	}
	if ar.Resource != "user" {
		t.Errorf("resource = %q, want %q", ar.Resource, "user")
	}
}

func TestParseFullMethod_RegisterDevice(t *testing.T) {
	fullMethod := "/ztcp.device.v1.DeviceService/RegisterDevice"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "register" {
		t.Errorf("action = %q, want %q", ar.Action, "register")
	}
	if ar.Resource != "device" {
		t.Errorf("resource = %q, want %q", ar.Resource, "device")
	}
}

func TestParseFullMethod_RevokeSession(t *testing.T) {
	fullMethod := "/ztcp.session.v1.SessionService/RevokeSession"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "revoke" {
		t.Errorf("action = %q, want %q", ar.Action, "revoke")
	}
	if ar.Resource != "session" {
		t.Errorf("resource = %q, want %q", ar.Resource, "session")
	}
}

func TestParseFullMethod_UnknownFormat(t *testing.T) {
	fullMethod := "invalid-format"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "unknown" {
		t.Errorf("action = %q, want %q", ar.Action, "unknown")
	}
	if ar.Resource != "unknown" {
		t.Errorf("resource = %q, want %q", ar.Resource, "unknown")
	}
}

func TestParseFullMethod_NoSlash(t *testing.T) {
	fullMethod := "SomeService/SomeMethod"
	ar := ParseFullMethod(fullMethod)

	// When there's no leading slash, strings.LastIndex finds the / and extracts method
	// but beforeSlash has no dot, so resource becomes "unknown"
	if ar.Action != "somemethod" {
		t.Errorf("action = %q, want %q", ar.Action, "somemethod")
	}
	if ar.Resource != "unknown" {
		t.Errorf("resource = %q, want %q", ar.Resource, "unknown")
	}
}

func TestParseFullMethod_OrganizationService(t *testing.T) {
	fullMethod := "/ztcp.organization.v1.OrganizationService/GetOrganization"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "get" {
		t.Errorf("action = %q, want %q", ar.Action, "get")
	}
	if ar.Resource != "organization" {
		t.Errorf("resource = %q, want %q", ar.Resource, "organization")
	}
}

func TestParseFullMethod_SuspendOrganization(t *testing.T) {
	fullMethod := "/ztcp.organization.v1.OrganizationService/SuspendOrganization"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "suspend" {
		t.Errorf("action = %q, want %q", ar.Action, "suspend")
	}
	if ar.Resource != "organization" {
		t.Errorf("resource = %q, want %q", ar.Resource, "organization")
	}
}

func TestParseFullMethod_UnknownMethod(t *testing.T) {
	fullMethod := "/ztcp.user.v1.UserService/UnknownMethod"
	ar := ParseFullMethod(fullMethod)

	if ar.Action != "unknownmethod" {
		t.Errorf("action = %q, want %q", ar.Action, "unknownmethod")
	}
	if ar.Resource != "user" {
		t.Errorf("resource = %q, want %q", ar.Resource, "user")
	}
}
