package audit

import "strings"

// ActionResource holds action and resource derived from a gRPC full method name.
type ActionResource struct {
	Action   string
	Resource string
}

// Membership method overrides: audit as user_added, user_removed, role_changed on resource "user".
const (
	membershipAddMember    = "/ztcp.membership.v1.MembershipService/AddMember"
	membershipRemoveMember = "/ztcp.membership.v1.MembershipService/RemoveMember"
	membershipUpdateRole   = "/ztcp.membership.v1.MembershipService/UpdateRole"
)

// ParseFullMethod returns action and resource for a gRPC full method (e.g. /ztcp.user.v1.UserService/GetUser).
// Action is a verb: get, list, create, update, delete, or a lowercase method name for others.
// Resource is derived from the service name (e.g. UserService -> user).
// MembershipService AddMember/RemoveMember/UpdateRole are mapped to user_added, user_removed, role_changed on resource "user".
func ParseFullMethod(fullMethod string) ActionResource {
	switch fullMethod {
	case membershipAddMember:
		return ActionResource{Action: "user_added", Resource: "user"}
	case membershipRemoveMember:
		return ActionResource{Action: "user_removed", Resource: "user"}
	case membershipUpdateRole:
		return ActionResource{Action: "role_changed", Resource: "user"}
	}
	// fullMethod format: /ztcp.package.v1.ServiceName/MethodName
	slash := strings.LastIndex(fullMethod, "/")
	if slash < 0 {
		return ActionResource{Action: "unknown", Resource: "unknown"}
	}
	method := fullMethod[slash+1:]
	beforeSlash := fullMethod[:slash]
	dot := strings.LastIndex(beforeSlash, ".")
	if dot < 0 {
		return ActionResource{Action: strings.ToLower(method), Resource: "unknown"}
	}
	serviceName := beforeSlash[dot+1:]
	resource := serviceToResource(serviceName)
	action := methodToAction(method)
	return ActionResource{Action: action, Resource: resource}
}

func serviceToResource(serviceName string) string {
	// UserService -> user, OrganizationService -> organization
	s := strings.TrimSuffix(serviceName, "Service")
	if s == "" {
		return "unknown"
	}
	return strings.ToLower(s[0:1]) + s[1:]
}

func methodToAction(method string) string {
	switch {
	case strings.HasPrefix(method, "Get") && method != "Get":
		return "get"
	case strings.HasPrefix(method, "List"):
		return "list"
	case strings.HasPrefix(method, "Create"):
		return "create"
	case strings.HasPrefix(method, "Update"):
		return "update"
	case strings.HasPrefix(method, "Delete"):
		return "delete"
	case strings.HasPrefix(method, "Add"):
		return "add"
	case strings.HasPrefix(method, "Remove"):
		return "remove"
	case strings.HasPrefix(method, "Register"):
		return "register"
	case strings.HasPrefix(method, "Revoke"):
		return "revoke"
	case strings.HasPrefix(method, "Suspend"):
		return "suspend"
	case strings.HasPrefix(method, "Emit"), strings.HasPrefix(method, "Batch"):
		return "emit"
	default:
		return strings.ToLower(method)
	}
}
