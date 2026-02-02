package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	devicev1 "zero-trust-control-plane/backend/api/generated/device/v1"
	"zero-trust-control-plane/backend/internal/device/domain"
	"zero-trust-control-plane/backend/internal/device/repository"
)

// Server implements DeviceService (proto server) for device trust and posture.
// Proto: device/device.proto → internal/device/handler.
type Server struct {
	devicev1.UnimplementedDeviceServiceServer
	repo repository.Repository
}

// NewServer returns a new Device gRPC server. Pass nil repo for stub (Unimplemented).
func NewServer(repo repository.Repository) *Server {
	return &Server{repo: repo}
}

// RegisterDevice registers a device. TODO: implement (auth creates device on login).
func (s *Server) RegisterDevice(ctx context.Context, req *devicev1.RegisterDeviceRequest) (*devicev1.RegisterDeviceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RegisterDevice not implemented")
}

// GetDevice returns a device by ID.
func (s *Server) GetDevice(ctx context.Context, req *devicev1.GetDeviceRequest) (*devicev1.GetDeviceResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method GetDevice not implemented")
	}
	dev, err := s.repo.GetByID(ctx, req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if dev == nil {
		return nil, status.Error(codes.NotFound, "device not found")
	}
	return &devicev1.GetDeviceResponse{Device: deviceToProto(dev)}, nil
}

// ListDevices returns a paginated list of devices for the org (and optional user filter).
func (s *Server) ListDevices(ctx context.Context, req *devicev1.ListDevicesRequest) (*devicev1.ListDevicesResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method ListDevices not implemented")
	}
	list, err := s.repo.ListByOrg(ctx, req.GetOrgId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	devices := make([]*devicev1.Device, 0, len(list))
	for _, d := range list {
		if req.GetUserId() != "" && d.UserID != req.GetUserId() {
			continue
		}
		devices = append(devices, deviceToProto(d))
	}
	return &devicev1.ListDevicesResponse{Devices: devices}, nil
}

// RevokeDevice revokes the device (sets revoked_at, clears trusted).
func (s *Server) RevokeDevice(ctx context.Context, req *devicev1.RevokeDeviceRequest) (*devicev1.RevokeDeviceResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method RevokeDevice not implemented")
	}
	if err := s.repo.Revoke(ctx, req.GetDeviceId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &devicev1.RevokeDeviceResponse{}, nil
}

func deviceToProto(d *domain.Device) *devicev1.Device {
	if d == nil {
		return nil
	}
	out := &devicev1.Device{
		Id:          d.ID,
		UserId:      d.UserID,
		OrgId:       d.OrgID,
		Fingerprint: d.Fingerprint,
		Trusted:     d.Trusted,
	}
	if d.LastSeenAt != nil {
		out.LastSeenAt = timestamppb.New(*d.LastSeenAt)
	}
	if d.TrustedUntil != nil {
		out.TrustedUntil = timestamppb.New(*d.TrustedUntil)
	}
	if d.RevokedAt != nil {
		out.RevokedAt = timestamppb.New(*d.RevokedAt)
	}
	out.CreatedAt = timestamppb.New(d.CreatedAt)
	return out
}
