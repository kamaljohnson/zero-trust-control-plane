package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	devicev1 "zero-trust-control-plane/backend/api/generated/device/v1"
	"zero-trust-control-plane/backend/internal/device/domain"
)

// mockDeviceRepo implements repository.Repository for tests.
type mockDeviceRepo struct {
	devices   map[string]*domain.Device
	byOrg     map[string][]*domain.Device
	getByIDErr error
	listErr   error
	revokeErr error
}

func (m *mockDeviceRepo) GetByID(ctx context.Context, id string) (*domain.Device, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.devices[id], nil
}

func (m *mockDeviceRepo) GetByUserOrgAndFingerprint(ctx context.Context, userID, orgID, fingerprint string) (*domain.Device, error) {
	return nil, nil
}

func (m *mockDeviceRepo) ListByOrg(ctx context.Context, orgID string) ([]*domain.Device, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.byOrg[orgID], nil
}

func (m *mockDeviceRepo) Create(ctx context.Context, d *domain.Device) error {
	return nil
}

func (m *mockDeviceRepo) UpdateTrusted(ctx context.Context, id string, trusted bool) error {
	return nil
}

func (m *mockDeviceRepo) UpdateTrustedWithExpiry(ctx context.Context, id string, trusted bool, trustedUntil *time.Time) error {
	return nil
}

func (m *mockDeviceRepo) Revoke(ctx context.Context, id string) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	return nil
}

func (m *mockDeviceRepo) UpdateLastSeen(ctx context.Context, id string, at time.Time) error {
	return nil
}

func TestGetDevice_Success(t *testing.T) {
	now := time.Now().UTC()
	device := &domain.Device{
		ID:          "device-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		Fingerprint: "fp-123",
		Trusted:     true,
		CreatedAt:   now,
	}
	repo := &mockDeviceRepo{
		devices: map[string]*domain.Device{"device-1": device},
		byOrg:   make(map[string][]*domain.Device),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.GetDevice(ctx, &devicev1.GetDeviceRequest{DeviceId: "device-1"})
	if err != nil {
		t.Fatalf("GetDevice: %v", err)
	}
	if resp == nil || resp.Device == nil {
		t.Fatal("response or device is nil")
	}
	if resp.Device.Id != "device-1" {
		t.Errorf("device id = %q, want %q", resp.Device.Id, "device-1")
	}
	if resp.Device.Fingerprint != "fp-123" {
		t.Errorf("device fingerprint = %q, want %q", resp.Device.Fingerprint, "fp-123")
	}
	if !resp.Device.Trusted {
		t.Error("device trusted = false, want true")
	}
}

func TestGetDevice_NotFound(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: make(map[string]*domain.Device),
		byOrg:   make(map[string][]*domain.Device),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.GetDevice(ctx, &devicev1.GetDeviceRequest{DeviceId: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestGetDevice_RepositoryError(t *testing.T) {
	repo := &mockDeviceRepo{
		devices:     make(map[string]*domain.Device),
		byOrg:       make(map[string][]*domain.Device),
		getByIDErr: errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.GetDevice(ctx, &devicev1.GetDeviceRequest{DeviceId: "device-1"})
	if err == nil {
		t.Fatal("expected error for repository error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestGetDevice_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.GetDevice(ctx, &devicev1.GetDeviceRequest{DeviceId: "device-1"})
	if err == nil {
		t.Fatal("expected error for nil repo")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestListDevices_Success(t *testing.T) {
	now := time.Now().UTC()
	devices := []*domain.Device{
		{ID: "device-1", UserID: "user-1", OrgID: "org-1", Fingerprint: "fp-1", Trusted: true, CreatedAt: now},
		{ID: "device-2", UserID: "user-2", OrgID: "org-1", Fingerprint: "fp-2", Trusted: false, CreatedAt: now},
	}
	repo := &mockDeviceRepo{
		devices: make(map[string]*domain.Device),
		byOrg:   map[string][]*domain.Device{"org-1": devices},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.ListDevices(ctx, &devicev1.ListDevicesRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(resp.Devices) != 2 {
		t.Errorf("devices count = %d, want 2", len(resp.Devices))
	}
}

func TestListDevices_FilteredByUserID(t *testing.T) {
	now := time.Now().UTC()
	devices := []*domain.Device{
		{ID: "device-1", UserID: "user-1", OrgID: "org-1", Fingerprint: "fp-1", Trusted: true, CreatedAt: now},
		{ID: "device-2", UserID: "user-2", OrgID: "org-1", Fingerprint: "fp-2", Trusted: false, CreatedAt: now},
		{ID: "device-3", UserID: "user-1", OrgID: "org-1", Fingerprint: "fp-3", Trusted: true, CreatedAt: now},
	}
	repo := &mockDeviceRepo{
		devices: make(map[string]*domain.Device),
		byOrg:   map[string][]*domain.Device{"org-1": devices},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.ListDevices(ctx, &devicev1.ListDevicesRequest{
		OrgId:  "org-1",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(resp.Devices) != 2 {
		t.Errorf("devices count = %d, want 2", len(resp.Devices))
	}
	for _, d := range resp.Devices {
		if d.UserId != "user-1" {
			t.Errorf("device user_id = %q, want %q", d.UserId, "user-1")
		}
	}
}

func TestListDevices_EmptyList(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: make(map[string]*domain.Device),
		byOrg:   map[string][]*domain.Device{"org-1": {}},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.ListDevices(ctx, &devicev1.ListDevicesRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(resp.Devices) != 0 {
		t.Errorf("devices count = %d, want 0", len(resp.Devices))
	}
}

func TestListDevices_RepositoryError(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: make(map[string]*domain.Device),
		byOrg:   make(map[string][]*domain.Device),
		listErr: errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.ListDevices(ctx, &devicev1.ListDevicesRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for repository error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestListDevices_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.ListDevices(ctx, &devicev1.ListDevicesRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for nil repo")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestRevokeDevice_Success(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: make(map[string]*domain.Device),
		byOrg:   make(map[string][]*domain.Device),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.RevokeDevice(ctx, &devicev1.RevokeDeviceRequest{DeviceId: "device-1"})
	if err != nil {
		t.Fatalf("RevokeDevice: %v", err)
	}
}

func TestRevokeDevice_RepositoryError(t *testing.T) {
	repo := &mockDeviceRepo{
		devices:   make(map[string]*domain.Device),
		byOrg:     make(map[string][]*domain.Device),
		revokeErr: errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.RevokeDevice(ctx, &devicev1.RevokeDeviceRequest{DeviceId: "device-1"})
	if err == nil {
		t.Fatal("expected error for repository error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestRevokeDevice_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.RevokeDevice(ctx, &devicev1.RevokeDeviceRequest{DeviceId: "device-1"})
	if err == nil {
		t.Fatal("expected error for nil repo")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestGetDevice_WithTimestamps(t *testing.T) {
	now := time.Now().UTC()
	lastSeen := now.Add(-1 * time.Hour)
	trustedUntil := now.Add(24 * time.Hour)
	revokedAt := now.Add(-2 * time.Hour)
	device := &domain.Device{
		ID:          "device-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		Fingerprint: "fp-123",
		Trusted:     true,
		LastSeenAt:  &lastSeen,
		TrustedUntil: &trustedUntil,
		RevokedAt:   &revokedAt,
		CreatedAt:   now,
	}
	repo := &mockDeviceRepo{
		devices: map[string]*domain.Device{"device-1": device},
		byOrg:   make(map[string][]*domain.Device),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.GetDevice(ctx, &devicev1.GetDeviceRequest{DeviceId: "device-1"})
	if err != nil {
		t.Fatalf("GetDevice: %v", err)
	}
	if resp.Device.LastSeenAt == nil {
		t.Error("LastSeenAt should be set")
	}
	if resp.Device.TrustedUntil == nil {
		t.Error("TrustedUntil should be set")
	}
	if resp.Device.RevokedAt == nil {
		t.Error("RevokedAt should be set")
	}
}

func TestRegisterDevice_Unimplemented(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: make(map[string]*domain.Device),
		byOrg:   make(map[string][]*domain.Device),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.RegisterDevice(ctx, &devicev1.RegisterDeviceRequest{})
	if err == nil {
		t.Fatal("expected error for unimplemented method")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

// Tests for deviceToProto helper function

func TestDeviceToProto_NilDevice(t *testing.T) {
	proto := deviceToProto(nil)
	if proto != nil {
		t.Error("deviceToProto(nil) should return nil")
	}
}

func TestDeviceToProto_AllTimestampsNil(t *testing.T) {
	now := time.Now().UTC()
	device := &domain.Device{
		ID:          "device-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		Fingerprint: "fp-123",
		Trusted:     true,
		LastSeenAt:  nil,
		TrustedUntil: nil,
		RevokedAt:   nil,
		CreatedAt:   now,
	}
	proto := deviceToProto(device)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.LastSeenAt != nil {
		t.Error("LastSeenAt should be nil")
	}
	if proto.TrustedUntil != nil {
		t.Error("TrustedUntil should be nil")
	}
	if proto.RevokedAt != nil {
		t.Error("RevokedAt should be nil")
	}
	if proto.CreatedAt == nil {
		t.Error("CreatedAt should be set")
	}
}

func TestDeviceToProto_AllTimestampsPresent(t *testing.T) {
	now := time.Now().UTC()
	lastSeen := now.Add(-2 * time.Hour)
	trustedUntil := now.Add(24 * time.Hour)
	revokedAt := now.Add(-1 * time.Hour)
	device := &domain.Device{
		ID:          "device-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		Fingerprint: "fp-123",
		Trusted:     false,
		LastSeenAt:  &lastSeen,
		TrustedUntil: &trustedUntil,
		RevokedAt:   &revokedAt,
		CreatedAt:   now,
	}
	proto := deviceToProto(device)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.LastSeenAt == nil {
		t.Error("LastSeenAt should be set")
	}
	if !proto.LastSeenAt.AsTime().Equal(lastSeen) {
		t.Errorf("LastSeenAt = %v, want %v", proto.LastSeenAt.AsTime(), lastSeen)
	}
	if proto.TrustedUntil == nil {
		t.Error("TrustedUntil should be set")
	}
	if !proto.TrustedUntil.AsTime().Equal(trustedUntil) {
		t.Errorf("TrustedUntil = %v, want %v", proto.TrustedUntil.AsTime(), trustedUntil)
	}
	if proto.RevokedAt == nil {
		t.Error("RevokedAt should be set")
	}
	if !proto.RevokedAt.AsTime().Equal(revokedAt) {
		t.Errorf("RevokedAt = %v, want %v", proto.RevokedAt.AsTime(), revokedAt)
	}
}

func TestDeviceToProto_MixedTimestamps(t *testing.T) {
	now := time.Now().UTC()
	lastSeen := now.Add(-1 * time.Hour)
	device := &domain.Device{
		ID:          "device-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		Fingerprint: "fp-123",
		Trusted:     true,
		LastSeenAt:  &lastSeen,
		TrustedUntil: nil,
		RevokedAt:   nil,
		CreatedAt:   now,
	}
	proto := deviceToProto(device)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.LastSeenAt == nil {
		t.Error("LastSeenAt should be set")
	}
	if proto.TrustedUntil != nil {
		t.Error("TrustedUntil should be nil")
	}
	if proto.RevokedAt != nil {
		t.Error("RevokedAt should be nil")
	}
}

func TestDeviceToProto_AllFields(t *testing.T) {
	now := time.Now().UTC()
	device := &domain.Device{
		ID:          "device-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		Fingerprint: "fp-123",
		Trusted:     true,
		CreatedAt:   now,
	}
	proto := deviceToProto(device)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Id != "device-1" {
		t.Errorf("Id = %q, want %q", proto.Id, "device-1")
	}
	if proto.UserId != "user-1" {
		t.Errorf("UserId = %q, want %q", proto.UserId, "user-1")
	}
	if proto.OrgId != "org-1" {
		t.Errorf("OrgId = %q, want %q", proto.OrgId, "org-1")
	}
	if proto.Fingerprint != "fp-123" {
		t.Errorf("Fingerprint = %q, want %q", proto.Fingerprint, "fp-123")
	}
	if !proto.Trusted {
		t.Error("Trusted should be true")
	}
}
