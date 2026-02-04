package handler

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "zero-trust-control-plane/backend/api/generated/common/v1"
	sessionv1 "zero-trust-control-plane/backend/api/generated/session/v1"
	"zero-trust-control-plane/backend/internal/audit"
	membershiprepo "zero-trust-control-plane/backend/internal/membership/repository"
	"zero-trust-control-plane/backend/internal/platform/rbac"
	"zero-trust-control-plane/backend/internal/session/domain"
	sessionrepo "zero-trust-control-plane/backend/internal/session/repository"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// Server implements SessionService (proto server) for session lifecycle.
// Proto: session/session.proto → internal/session/handler.
type Server struct {
	sessionv1.UnimplementedSessionServiceServer
	sessionRepo    sessionrepo.Repository
	membershipRepo membershiprepo.Repository
	auditLogger    audit.AuditLogger
}

// NewServer returns a new Session gRPC server. If sessionRepo is nil, all RPCs return Unimplemented.
func NewServer(sessionRepo sessionrepo.Repository, membershipRepo membershiprepo.Repository, auditLogger audit.AuditLogger) *Server {
	return &Server{
		sessionRepo:    sessionRepo,
		membershipRepo: membershipRepo,
		auditLogger:    auditLogger,
	}
}

// RevokeSession revokes a session. Caller must be org admin or owner; session must belong to caller's org.
func (s *Server) RevokeSession(ctx context.Context, req *sessionv1.RevokeSessionRequest) (*sessionv1.RevokeSessionResponse, error) {
	if s.sessionRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method RevokeSession not implemented")
	}
	orgID, userID, err := rbac.RequireOrgAdmin(ctx, s.membershipRepo)
	if err != nil {
		return nil, err
	}
	sessionID := req.GetSessionId()
	if sessionID == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id required")
	}
	ses, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get session")
	}
	if ses == nil {
		return nil, status.Error(codes.NotFound, "session not found")
	}
	if ses.OrgID != orgID {
		return nil, status.Error(codes.PermissionDenied, "session does not belong to your organization")
	}
	if err := s.sessionRepo.Revoke(ctx, sessionID); err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke session")
	}
	if s.auditLogger != nil {
		s.auditLogger.LogEvent(ctx, orgID, userID, "revoke", "session", sessionID)
	}
	return &sessionv1.RevokeSessionResponse{}, nil
}

// ListSessions returns a paginated list of sessions for the org, optionally filtered by user. Caller must be org admin or owner.
func (s *Server) ListSessions(ctx context.Context, req *sessionv1.ListSessionsRequest) (*sessionv1.ListSessionsResponse, error) {
	if s.sessionRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method ListSessions not implemented")
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
	var userID *string
	if req.GetUserId() != "" {
		userID = &req.UserId
	}
	list, err := s.sessionRepo.ListByOrg(ctx, targetOrgID, userID, pageSize, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list sessions")
	}
	sessions := make([]*sessionv1.Session, len(list))
	for i := range list {
		sessions[i] = domainSessionToProto(list[i])
	}
	nextToken := ""
	if len(list) == int(pageSize) {
		nextToken = strconv.Itoa(int(offset + pageSize))
	}
	return &sessionv1.ListSessionsResponse{
		Sessions: sessions,
		Pagination: &commonv1.PaginationResult{
			NextPageToken: nextToken,
		},
	}, nil
}

// GetSession returns a session by ID. Caller must be org admin or owner; session must belong to caller's org.
func (s *Server) GetSession(ctx context.Context, req *sessionv1.GetSessionRequest) (*sessionv1.GetSessionResponse, error) {
	if s.sessionRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method GetSession not implemented")
	}
	orgID, _, err := rbac.RequireOrgAdmin(ctx, s.membershipRepo)
	if err != nil {
		return nil, err
	}
	sessionID := req.GetSessionId()
	if sessionID == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id required")
	}
	ses, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get session")
	}
	if ses == nil {
		return nil, status.Error(codes.NotFound, "session not found")
	}
	if ses.OrgID != orgID {
		return nil, status.Error(codes.PermissionDenied, "session does not belong to your organization")
	}
	return &sessionv1.GetSessionResponse{
		Session: domainSessionToProto(ses),
	}, nil
}

// RevokeAllSessionsForUser revokes all sessions for the given user in the org. Caller must be org admin or owner.
func (s *Server) RevokeAllSessionsForUser(ctx context.Context, req *sessionv1.RevokeAllSessionsForUserRequest) (*sessionv1.RevokeAllSessionsForUserResponse, error) {
	if s.sessionRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method RevokeAllSessionsForUser not implemented")
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
	if err := s.sessionRepo.RevokeAllSessionsByUserAndOrg(ctx, targetUserID, targetOrgID); err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke sessions")
	}
	if s.auditLogger != nil {
		s.auditLogger.LogEvent(ctx, targetOrgID, userID, "revoke", "session", "all:"+targetUserID)
	}
	return &sessionv1.RevokeAllSessionsForUserResponse{}, nil
}

func domainSessionToProto(s *domain.Session) *sessionv1.Session {
	if s == nil {
		return nil
	}
	var revokedAt, lastSeenAt *timestamppb.Timestamp
	if s.RevokedAt != nil {
		revokedAt = timestamppb.New(*s.RevokedAt)
	}
	if s.LastSeenAt != nil {
		lastSeenAt = timestamppb.New(*s.LastSeenAt)
	}
	return &sessionv1.Session{
		Id:         s.ID,
		UserId:     s.UserID,
		OrgId:      s.OrgID,
		DeviceId:   s.DeviceID,
		ExpiresAt:  timestamppb.New(s.ExpiresAt),
		RevokedAt:  revokedAt,
		LastSeenAt: lastSeenAt,
		IpAddress:  s.IPAddress,
		CreatedAt:  timestamppb.New(s.CreatedAt),
	}
}
