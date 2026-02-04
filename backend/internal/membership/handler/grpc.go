package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "zero-trust-control-plane/backend/api/generated/common/v1"
	membershipv1 "zero-trust-control-plane/backend/api/generated/membership/v1"
	"zero-trust-control-plane/backend/internal/audit"
	"zero-trust-control-plane/backend/internal/membership/domain"
	membershiprepo "zero-trust-control-plane/backend/internal/membership/repository"
	"zero-trust-control-plane/backend/internal/platform/rbac"
	userrepo "zero-trust-control-plane/backend/internal/user/repository"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// Server implements MembershipService (proto server) for org membership and roles.
// Proto: membership/membership.proto → internal/membership/handler.
type Server struct {
	membershipv1.UnimplementedMembershipServiceServer
	membershipRepo membershiprepo.Repository
	userRepo       userrepo.Repository
	auditLogger    audit.AuditLogger
}

// NewServer returns a new Membership gRPC server. If membershipRepo is nil, all RPCs return Unimplemented.
func NewServer(membershipRepo membershiprepo.Repository, userRepo userrepo.Repository, auditLogger audit.AuditLogger) *Server {
	return &Server{
		membershipRepo: membershipRepo,
		userRepo:       userRepo,
		auditLogger:    auditLogger,
	}
}

// AddMember adds a member to an organization. Caller must be org admin or owner.
func (s *Server) AddMember(ctx context.Context, req *membershipv1.AddMemberRequest) (*membershipv1.AddMemberResponse, error) {
	if s.membershipRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method AddMember not implemented")
	}
	orgID, userID, err := rbac.RequireOrgAdmin(ctx, s.membershipRepo)
	if err != nil {
		return nil, err
	}
	if req.GetOrgId() != "" && req.GetOrgId() != orgID {
		return nil, status.Error(codes.PermissionDenied, "org_id does not match context")
	}
	targetOrgID := req.GetOrgId()
	if targetOrgID == "" {
		targetOrgID = orgID
	}
	targetUserID := req.GetUserId()
	if targetUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	role := protoRoleToDomain(req.GetRole())
	if role == "" {
		role = domain.RoleMember
	}
	if role != domain.RoleAdmin && role != domain.RoleMember {
		return nil, status.Error(codes.InvalidArgument, "role must be admin or member")
	}
	if s.userRepo != nil {
		u, err := s.userRepo.GetByID(ctx, targetUserID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to look up user")
		}
		if u == nil {
			return nil, status.Error(codes.NotFound, "user not found")
		}
	}
	existing, err := s.membershipRepo.GetMembershipByUserAndOrg(ctx, targetUserID, targetOrgID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check membership")
	}
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "user is already a member")
	}
	m := &domain.Membership{
		ID:        uuid.New().String(),
		UserID:    targetUserID,
		OrgID:     targetOrgID,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.membershipRepo.CreateMembership(ctx, m); err != nil {
		return nil, status.Error(codes.Internal, "failed to create membership")
	}
	if s.auditLogger != nil {
		s.auditLogger.LogEvent(ctx, targetOrgID, userID, "add", "membership", targetUserID)
	}
	return &membershipv1.AddMemberResponse{
		Member: domainMemberToProto(m),
	}, nil
}

// RemoveMember removes a member from an organization. Caller must be org admin or owner. Cannot remove the last owner.
func (s *Server) RemoveMember(ctx context.Context, req *membershipv1.RemoveMemberRequest) (*membershipv1.RemoveMemberResponse, error) {
	if s.membershipRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method RemoveMember not implemented")
	}
	orgID, userID, err := rbac.RequireOrgAdmin(ctx, s.membershipRepo)
	if err != nil {
		return nil, err
	}
	if req.GetOrgId() != "" && req.GetOrgId() != orgID {
		return nil, status.Error(codes.PermissionDenied, "org_id does not match context")
	}
	targetOrgID := req.GetOrgId()
	if targetOrgID == "" {
		targetOrgID = orgID
	}
	targetUserID := req.GetUserId()
	if targetUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	m, err := s.membershipRepo.GetMembershipByUserAndOrg(ctx, targetUserID, targetOrgID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to look up membership")
	}
	if m == nil {
		return nil, status.Error(codes.NotFound, "membership not found")
	}
	if m.Role == domain.RoleOwner {
		count, err := s.membershipRepo.CountOwnersByOrg(ctx, targetOrgID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to count owners")
		}
		if count <= 1 {
			return nil, status.Error(codes.FailedPrecondition, "cannot remove the last owner")
		}
	}
	if err := s.membershipRepo.DeleteByUserAndOrg(ctx, targetUserID, targetOrgID); err != nil {
		return nil, status.Error(codes.Internal, "failed to remove member")
	}
	if s.auditLogger != nil {
		s.auditLogger.LogEvent(ctx, targetOrgID, userID, "remove", "membership", targetUserID)
	}
	return &membershipv1.RemoveMemberResponse{}, nil
}

// UpdateRole updates a member's role. Caller must be org admin or owner. Cannot demote the last owner.
func (s *Server) UpdateRole(ctx context.Context, req *membershipv1.UpdateRoleRequest) (*membershipv1.UpdateRoleResponse, error) {
	if s.membershipRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method UpdateRole not implemented")
	}
	orgID, userID, err := rbac.RequireOrgAdmin(ctx, s.membershipRepo)
	if err != nil {
		return nil, err
	}
	if req.GetOrgId() != "" && req.GetOrgId() != orgID {
		return nil, status.Error(codes.PermissionDenied, "org_id does not match context")
	}
	targetOrgID := req.GetOrgId()
	if targetOrgID == "" {
		targetOrgID = orgID
	}
	targetUserID := req.GetUserId()
	if targetUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	newRole := protoRoleToDomain(req.GetRole())
	if newRole != domain.RoleOwner && newRole != domain.RoleAdmin && newRole != domain.RoleMember {
		return nil, status.Error(codes.InvalidArgument, "role must be owner, admin, or member")
	}
	m, err := s.membershipRepo.GetMembershipByUserAndOrg(ctx, targetUserID, targetOrgID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to look up membership")
	}
	if m == nil {
		return nil, status.Error(codes.NotFound, "membership not found")
	}
	if m.Role == domain.RoleOwner && newRole != domain.RoleOwner {
		count, err := s.membershipRepo.CountOwnersByOrg(ctx, targetOrgID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to count owners")
		}
		if count <= 1 {
			return nil, status.Error(codes.FailedPrecondition, "cannot demote the last owner")
		}
	}
	updated, err := s.membershipRepo.UpdateRole(ctx, targetUserID, targetOrgID, newRole)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update role")
	}
	if updated == nil {
		return nil, status.Error(codes.NotFound, "membership not found")
	}
	if s.auditLogger != nil {
		s.auditLogger.LogEvent(ctx, targetOrgID, userID, "update", "membership", targetUserID+":"+string(newRole))
	}
	return &membershipv1.UpdateRoleResponse{
		Member: domainMemberToProto(updated),
	}, nil
}

// ListMembers returns a paginated list of members for the org. Caller must be org admin or owner.
func (s *Server) ListMembers(ctx context.Context, req *membershipv1.ListMembersRequest) (*membershipv1.ListMembersResponse, error) {
	if s.membershipRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method ListMembers not implemented")
	}
	orgID, _, err := rbac.RequireOrgAdmin(ctx, s.membershipRepo)
	if err != nil {
		return nil, err
	}
	if req.GetOrgId() != "" && req.GetOrgId() != orgID {
		return nil, status.Error(codes.PermissionDenied, "org_id does not match context")
	}
	targetOrgID := req.GetOrgId()
	if targetOrgID == "" {
		targetOrgID = orgID
	}
	pageSize := int32(defaultPageSize)
	if pag := req.GetPagination(); pag != nil {
		if ps := pag.GetPageSize(); ps > 0 {
			pageSize = ps
		}
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset := int32(0)
	if pag := req.GetPagination(); pag != nil {
		if tok := pag.GetPageToken(); tok != "" {
			if n, err := strconv.ParseInt(tok, 10, 32); err == nil && n >= 0 {
				offset = int32(n)
			}
		}
	}
	all, err := s.membershipRepo.ListMembershipsByOrg(ctx, targetOrgID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list members")
	}
	total := int32(len(all))
	start := offset
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	page := all[int(start):int(end)]
	members := make([]*membershipv1.Member, len(page))
	for i := range page {
		members[i] = domainMemberToProto(page[i])
	}
	nextToken := ""
	if end < total {
		nextToken = strconv.Itoa(int(end))
	}
	return &membershipv1.ListMembersResponse{
		Members: members,
		Pagination: &commonv1.PaginationResult{
			NextPageToken: nextToken,
		},
	}, nil
}

func protoRoleToDomain(r membershipv1.Role) domain.Role {
	switch r {
	case membershipv1.Role_ROLE_OWNER:
		return domain.RoleOwner
	case membershipv1.Role_ROLE_ADMIN:
		return domain.RoleAdmin
	case membershipv1.Role_ROLE_MEMBER:
		return domain.RoleMember
	default:
		return ""
	}
}

func domainRoleToProto(r domain.Role) membershipv1.Role {
	switch r {
	case domain.RoleOwner:
		return membershipv1.Role_ROLE_OWNER
	case domain.RoleAdmin:
		return membershipv1.Role_ROLE_ADMIN
	case domain.RoleMember:
		return membershipv1.Role_ROLE_MEMBER
	default:
		return membershipv1.Role_ROLE_UNSPECIFIED
	}
}

func domainMemberToProto(m *domain.Membership) *membershipv1.Member {
	if m == nil {
		return nil
	}
	return &membershipv1.Member{
		Id:        m.ID,
		UserId:    m.UserID,
		OrgId:     m.OrgID,
		Role:      domainRoleToProto(m.Role),
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
}
