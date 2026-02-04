package interceptors

import (
	"context"
	"errors"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	auditdomain "zero-trust-control-plane/backend/internal/audit/domain"
)

// mockAuditRepoForInterceptor implements auditrepo.Repository for interceptor tests.
type mockAuditRepoForInterceptor struct {
	entries []*auditdomain.AuditLog
	err     error
}

func (m *mockAuditRepoForInterceptor) GetByID(ctx context.Context, id string) (*auditdomain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditRepoForInterceptor) ListByOrg(ctx context.Context, orgID string, limit, offset int32) ([]*auditdomain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditRepoForInterceptor) ListByOrgFiltered(ctx context.Context, orgID string, limit, offset int32, userID, action, resource *string) ([]*auditdomain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditRepoForInterceptor) Create(ctx context.Context, a *auditdomain.AuditLog) error {
	if m.err != nil {
		return m.err
	}
	m.entries = append(m.entries, a)
	return nil
}

func TestAuditUnary_SkipMethod(t *testing.T) {
	repo := &mockAuditRepoForInterceptor{
		entries: make([]*auditdomain.AuditLog, 0),
	}
	skipMethods := map[string]bool{
		"/test.Service/HealthCheck": true,
	}
	interceptor := AuditUnary(repo, skipMethods)

	ctx := WithIdentity(context.Background(), "user-1", "org-1", "session-1")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/HealthCheck",
	}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "success" {
		t.Errorf("response = %v, want %q", resp, "success")
	}
	if len(repo.entries) != 0 {
		t.Errorf("audit entries = %d, want 0", len(repo.entries))
	}
}

func TestAuditUnary_AuthenticatedRequest(t *testing.T) {
	repo := &mockAuditRepoForInterceptor{
		entries: make([]*auditdomain.AuditLog, 0),
	}
	skipMethods := map[string]bool{}
	interceptor := AuditUnary(repo, skipMethods)

	ctx := WithIdentity(context.Background(), "user-1", "org-1", "session-1")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/SomeMethod",
	}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "success" {
		t.Errorf("response = %v, want %q", resp, "success")
	}
	if len(repo.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(repo.entries))
	}
	entry := repo.entries[0]
	if entry.OrgID != "org-1" {
		t.Errorf("entry org_id = %q, want %q", entry.OrgID, "org-1")
	}
	if entry.UserID != "user-1" {
		t.Errorf("entry user_id = %q, want %q", entry.UserID, "user-1")
	}
}

func TestAuditUnary_UnauthenticatedRequest(t *testing.T) {
	repo := &mockAuditRepoForInterceptor{
		entries: make([]*auditdomain.AuditLog, 0),
	}
	skipMethods := map[string]bool{}
	interceptor := AuditUnary(repo, skipMethods)

	ctx := context.Background()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/SomeMethod",
	}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "success" {
		t.Errorf("response = %v, want %q", resp, "success")
	}
	if len(repo.entries) != 0 {
		t.Errorf("audit entries = %d, want 0", len(repo.entries))
	}
}

func TestAuditUnary_RepositoryError(t *testing.T) {
	repo := &mockAuditRepoForInterceptor{
		entries: make([]*auditdomain.AuditLog, 0),
		err:     errors.New("database error"),
	}
	skipMethods := map[string]bool{}
	interceptor := AuditUnary(repo, skipMethods)

	ctx := WithIdentity(context.Background(), "user-1", "org-1", "session-1")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/SomeMethod",
	}, handler)
	if err != nil {
		t.Fatalf("interceptor should not fail on audit error: %v", err)
	}
	if resp != "success" {
		t.Errorf("response = %v, want %q", resp, "success")
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"x-forwarded-for": "192.168.1.1",
	}))
	ip := ClientIP(ctx)
	if ip != "192.168.1.1" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.1")
	}
}

func TestClientIP_XForwardedFor_WithComma(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"x-forwarded-for": "192.168.1.1, 10.0.0.1",
	}))
	ip := ClientIP(ctx)
	if ip != "192.168.1.1" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.1")
	}
}

func TestClientIP_XRealIP(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"x-real-ip": "192.168.1.2",
	}))
	ip := ClientIP(ctx)
	if ip != "192.168.1.2" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.2")
	}
}

func TestClientIP_XForwardedFor_Precedence(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"x-forwarded-for": "192.168.1.1",
		"x-real-ip":       "192.168.1.2",
	}))
	ip := ClientIP(ctx)
	if ip != "192.168.1.1" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.1")
	}
}

func TestClientIP_PeerAddress(t *testing.T) {
	addr := &net.TCPAddr{
		IP:   net.ParseIP("192.168.1.3"),
		Port: 12345,
	}
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: addr,
	})
	ip := ClientIP(ctx)
	if ip != "192.168.1.3" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.3")
	}
}

func TestClientIP_Unknown(t *testing.T) {
	ctx := context.Background()
	ip := ClientIP(ctx)
	if ip != "unknown" {
		t.Errorf("ip = %q, want %q", ip, "unknown")
	}
}

func TestClientIP_Whitespace(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"x-forwarded-for": "  192.168.1.1  ",
	}))
	ip := ClientIP(ctx)
	if ip != "192.168.1.1" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.1")
	}
}

func TestAuditUnary_ParseFullMethod(t *testing.T) {
	repo := &mockAuditRepoForInterceptor{
		entries: make([]*auditdomain.AuditLog, 0),
	}
	skipMethods := map[string]bool{}
	interceptor := AuditUnary(repo, skipMethods)

	ctx := WithIdentity(context.Background(), "user-1", "org-1", "session-1")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/SomeMethod",
	}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "success" {
		t.Errorf("response = %v, want %q", resp, "success")
	}
	if len(repo.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(repo.entries))
	}
	entry := repo.entries[0]
	if entry.Action == "" {
		t.Error("entry action should be set")
	}
	if entry.Resource == "" {
		t.Error("entry resource should be set")
	}
}

func TestAuditUnary_HandlerError(t *testing.T) {
	repo := &mockAuditRepoForInterceptor{
		entries: make([]*auditdomain.AuditLog, 0),
	}
	skipMethods := map[string]bool{}
	interceptor := AuditUnary(repo, skipMethods)

	ctx := WithIdentity(context.Background(), "user-1", "org-1", "session-1")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New("handler error")
	}

	_, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/SomeMethod",
	}, handler)
	if err == nil {
		t.Fatal("expected error from handler")
	}
	if len(repo.entries) != 1 {
		t.Errorf("audit entries = %d, want 1", len(repo.entries))
	}
}
