package handler

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	auditv1 "zero-trust-control-plane/backend/api/generated/audit/v1"
	commonv1 "zero-trust-control-plane/backend/api/generated/common/v1"
	"zero-trust-control-plane/backend/internal/audit/domain"
	"zero-trust-control-plane/backend/internal/platform/rbac"
	"zero-trust-control-plane/backend/internal/server/interceptors"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// Server implements AuditService (proto server) for audit logs.
// Proto: audit/audit.proto → internal/audit/handler.
type Server struct {
	auditv1.UnimplementedAuditServiceServer
	repo            Repository
	orgAdminChecker rbac.OrgMembershipGetter
}

// Repository is the minimal interface needed by the audit handler for listing logs.
type Repository interface {
	ListByOrgFiltered(ctx context.Context, orgID string, limit, offset int32, userID, action, resource *string) ([]*domain.AuditLog, error)
}

// NewServer returns a new Audit gRPC server that uses repo for listing audit logs.
// If orgAdminChecker is non-nil, ListAuditLogs requires the caller to be org admin or owner.
func NewServer(repo Repository, orgAdminChecker rbac.OrgMembershipGetter) *Server {
	return &Server{repo: repo, orgAdminChecker: orgAdminChecker}
}

// ListAuditLogs returns a paginated list of audit logs for the caller's org, with optional filters.
// Caller must be authenticated; if orgAdminChecker is set, caller must be org admin or owner.
func (s *Server) ListAuditLogs(ctx context.Context, req *auditv1.ListAuditLogsRequest) (*auditv1.ListAuditLogsResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method ListAuditLogs not implemented")
	}
	var orgID string
	if s.orgAdminChecker != nil {
		var err error
		orgID, _, err = rbac.RequireOrgAdmin(ctx, s.orgAdminChecker)
		if err != nil {
			return nil, err
		}
	} else {
		var ok bool
		orgID, ok = interceptors.GetOrgID(ctx)
		if !ok || orgID == "" {
			return nil, status.Error(codes.Unauthenticated, "org context required")
		}
	}
	if req.GetOrgId() != "" && req.GetOrgId() != orgID {
		return nil, status.Error(codes.PermissionDenied, "org_id does not match context")
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
	var userID, action, resource *string
	if req.GetUserId() != "" {
		userID = &req.UserId
	}
	if req.GetAction() != "" {
		action = &req.Action
	}
	if req.GetResource() != "" {
		resource = &req.Resource
	}
	logs, err := s.repo.ListByOrgFiltered(ctx, orgID, pageSize, offset, userID, action, resource)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list audit logs")
	}
	events := make([]*auditv1.AuditEvent, len(logs))
	for i, l := range logs {
		events[i] = auditLogToProto(l)
	}
	result := &auditv1.ListAuditLogsResponse{
		Logs: events,
		Pagination: &commonv1.PaginationResult{
			NextPageToken: "",
		},
	}
	if len(logs) == int(pageSize) {
		result.Pagination.NextPageToken = strconv.Itoa(int(offset + pageSize))
	}
	return result, nil
}

func auditLogToProto(l *domain.AuditLog) *auditv1.AuditEvent {
	if l == nil {
		return nil
	}
	return &auditv1.AuditEvent{
		Id:        l.ID,
		OrgId:     l.OrgID,
		UserId:    l.UserID,
		Action:    l.Action,
		Resource:  l.Resource,
		Ip:        l.IP,
		Metadata:  l.Metadata,
		CreatedAt: timestamppb.New(l.CreatedAt),
	}
}
