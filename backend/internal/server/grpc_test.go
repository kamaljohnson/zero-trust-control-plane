package server

import (
	"testing"

	"google.golang.org/grpc"

	devv1 "zero-trust-control-plane/backend/api/generated/dev/v1"
	identityservice "zero-trust-control-plane/backend/internal/identity/service"
)

// mockServiceRegistrar implements grpc.ServiceRegistrar for testing.
type mockServiceRegistrar struct {
	callCount int
	services  []string
}

func (m *mockServiceRegistrar) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	m.callCount++
	if m.services == nil {
		m.services = make([]string, 0)
	}
	m.services = append(m.services, desc.ServiceName)
}

func TestRegisterServices_AllServicesRegistered(t *testing.T) {
	mockReg := &mockServiceRegistrar{}
	deps := Deps{
		Auth: &identityservice.AuthService{},
	}

	RegisterServices(mockReg, deps)

	// Should register 12 services (11 always + 0 DevService when nil)
	expectedCount := 12
	if mockReg.callCount != expectedCount {
		t.Errorf("RegisterService called %d times, want %d", mockReg.callCount, expectedCount)
	}
}

func TestRegisterServices_DevServiceNotRegisteredWhenNil(t *testing.T) {
	mockReg := &mockServiceRegistrar{}
	deps := Deps{
		Auth: &identityservice.AuthService{},
		// DevOTPHandler is nil
	}

	RegisterServices(mockReg, deps)

	// Should register 12 services (11 always + 0 DevService)
	expectedCount := 12
	if mockReg.callCount != expectedCount {
		t.Errorf("RegisterService called %d times, want %d (DevService should not be registered)", mockReg.callCount, expectedCount)
	}
}

func TestRegisterServices_DevServiceRegisteredWhenProvided(t *testing.T) {
	mockReg := &mockServiceRegistrar{}
	mockDevHandler := &mockDevService{}
	deps := Deps{
		Auth:          &identityservice.AuthService{},
		DevOTPHandler: mockDevHandler,
	}

	RegisterServices(mockReg, deps)

	// Should register 13 services (11 always + 1 DevService)
	expectedCount := 13
	if mockReg.callCount != expectedCount {
		t.Errorf("RegisterService called %d times, want %d (DevService should be registered)", mockReg.callCount, expectedCount)
	}
}

func TestRegisterServices_NilDependencies(t *testing.T) {
	mockReg := &mockServiceRegistrar{}
	deps := Deps{} // All dependencies are nil

	// Should not panic with nil dependencies
	RegisterServices(mockReg, deps)

	// Should still register all services (they handle nil dependencies internally)
	expectedCount := 12
	if mockReg.callCount != expectedCount {
		t.Errorf("RegisterService called %d times, want %d (services should be registered even with nil deps)", mockReg.callCount, expectedCount)
	}
}

// mockDevService implements devv1.DevServiceServer for testing.
type mockDevService struct {
	devv1.UnimplementedDevServiceServer
}
