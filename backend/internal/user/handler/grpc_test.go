package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "zero-trust-control-plane/backend/api/generated/user/v1"
	"zero-trust-control-plane/backend/internal/user/domain"
)

// mockUserRepo implements userrepo.Repository for tests.
type mockUserRepo struct {
	usersByID    map[string]*domain.User
	usersByEmail map[string]*domain.User
	getByIDErr   error
	getByEmailErr error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.usersByID[id], nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	return m.usersByEmail[email], nil
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	return nil
}

func (m *mockUserRepo) SetPhoneVerified(ctx context.Context, userID, phone string) error {
	return nil
}

func TestGetUser_Success(t *testing.T) {
	now := time.Now().UTC()
	user := &domain.User{
		ID:        "user-1",
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    domain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo := &mockUserRepo{
		usersByID: map[string]*domain.User{"user-1": user},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.GetUser(ctx, &userv1.GetUserRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if resp == nil || resp.User == nil {
		t.Fatal("response or user is nil")
	}
	if resp.User.Id != "user-1" {
		t.Errorf("user id = %q, want %q", resp.User.Id, "user-1")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("user email = %q, want %q", resp.User.Email, "test@example.com")
	}
	if resp.User.Name != "Test User" {
		t.Errorf("user name = %q, want %q", resp.User.Name, "Test User")
	}
	if resp.User.Status != userv1.UserStatus_USER_STATUS_ACTIVE {
		t.Errorf("user status = %v, want %v", resp.User.Status, userv1.UserStatus_USER_STATUS_ACTIVE)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domain.User),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.GetUser(ctx, &userv1.GetUserRequest{UserId: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestGetUser_InvalidUserID(t *testing.T) {
	repo := &mockUserRepo{usersByID: make(map[string]*domain.User)}
	srv := NewServer(repo)
	ctx := context.Background()

	testCases := []struct {
		name    string
		userID  string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"only spaces", "\t\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.GetUser(ctx, &userv1.GetUserRequest{UserId: tc.userID})
			if err == nil {
				t.Fatal("expected error for invalid user_id")
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("error is not a gRPC status: %v", err)
			}
			if st.Code() != codes.InvalidArgument {
				t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
			}
		})
	}
}

func TestGetUser_RepositoryError(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:  make(map[string]*domain.User),
		getByIDErr: errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.GetUser(ctx, &userv1.GetUserRequest{UserId: "user-1"})
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

func TestGetUser_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.GetUser(ctx, &userv1.GetUserRequest{UserId: "user-1"})
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

func TestGetUserByEmail_Success(t *testing.T) {
	now := time.Now().UTC()
	user := &domain.User{
		ID:        "user-1",
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    domain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo := &mockUserRepo{
		usersByEmail: map[string]*domain.User{"test@example.com": user},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.GetUserByEmail(ctx, &userv1.GetUserByEmailRequest{Email: "test@example.com"})
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if resp == nil || resp.User == nil {
		t.Fatal("response or user is nil")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("user email = %q, want %q", resp.User.Email, "test@example.com")
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		usersByEmail: make(map[string]*domain.User),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.GetUserByEmail(ctx, &userv1.GetUserByEmailRequest{Email: "nonexistent@example.com"})
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestGetUserByEmail_InvalidEmail(t *testing.T) {
	repo := &mockUserRepo{usersByEmail: make(map[string]*domain.User)}
	srv := NewServer(repo)
	ctx := context.Background()

	testCases := []struct {
		name  string
		email string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"only spaces", "\t\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.GetUserByEmail(ctx, &userv1.GetUserByEmailRequest{Email: tc.email})
			if err == nil {
				t.Fatal("expected error for invalid email")
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("error is not a gRPC status: %v", err)
			}
			if st.Code() != codes.InvalidArgument {
				t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
			}
		})
	}
}

func TestGetUserByEmail_RepositoryError(t *testing.T) {
	repo := &mockUserRepo{
		usersByEmail: make(map[string]*domain.User),
		getByEmailErr: errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.GetUserByEmail(ctx, &userv1.GetUserByEmailRequest{Email: "test@example.com"})
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

func TestGetUserByEmail_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.GetUserByEmail(ctx, &userv1.GetUserByEmailRequest{Email: "test@example.com"})
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

func TestGetUser_DisabledStatus(t *testing.T) {
	now := time.Now().UTC()
	user := &domain.User{
		ID:        "user-1",
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    domain.UserStatusDisabled,
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo := &mockUserRepo{
		usersByID: map[string]*domain.User{"user-1": user},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.GetUser(ctx, &userv1.GetUserRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if resp.User.Status != userv1.UserStatus_USER_STATUS_DISABLED {
		t.Errorf("user status = %v, want %v", resp.User.Status, userv1.UserStatus_USER_STATUS_DISABLED)
	}
}

func TestListUsers_Unimplemented(t *testing.T) {
	repo := &mockUserRepo{usersByID: make(map[string]*domain.User)}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.ListUsers(ctx, &userv1.ListUsersRequest{})
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

func TestListUsers_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.ListUsers(ctx, &userv1.ListUsersRequest{})
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

func TestDisableUser_Unimplemented(t *testing.T) {
	repo := &mockUserRepo{usersByID: make(map[string]*domain.User)}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.DisableUser(ctx, &userv1.DisableUserRequest{UserId: "user-1"})
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

func TestDisableUser_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.DisableUser(ctx, &userv1.DisableUserRequest{UserId: "user-1"})
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

func TestEnableUser_Unimplemented(t *testing.T) {
	repo := &mockUserRepo{usersByID: make(map[string]*domain.User)}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.EnableUser(ctx, &userv1.EnableUserRequest{UserId: "user-1"})
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

func TestEnableUser_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.EnableUser(ctx, &userv1.EnableUserRequest{UserId: "user-1"})
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
