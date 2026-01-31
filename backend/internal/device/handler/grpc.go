package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	devicev1 "zero-trust-control-plane/backend/api/generated/device/v1"
)

// Server implements DeviceService (proto server) for device trust and posture.
// Proto: device/device.proto → internal/device/handler.
type Server struct {
	devicev1.UnimplementedDeviceServiceServer
}

// NewServer returns a new Device gRPC server.
func NewServer() *Server {
	return &Server{}
}

// RegisterDevice registers a device. TODO: implement.
func (s *Server) RegisterDevice(ctx context.Context, req *devicev1.RegisterDeviceRequest) (*devicev1.RegisterDeviceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RegisterDevice not implemented")
}

// GetDevice returns a device by ID. TODO: implement.
func (s *Server) GetDevice(ctx context.Context, req *devicev1.GetDeviceRequest) (*devicev1.GetDeviceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetDevice not implemented")
}

// ListDevices returns a paginated list of devices. TODO: implement.
func (s *Server) ListDevices(ctx context.Context, req *devicev1.ListDevicesRequest) (*devicev1.ListDevicesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListDevices not implemented")
}
