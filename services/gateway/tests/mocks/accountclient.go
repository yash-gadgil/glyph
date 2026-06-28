package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockAccountClient struct {
	mock.Mock
}

func (m *MockAccountClient) SignupUser(ctx context.Context, req *userpb.SignupUserInfo, opts ...grpc.CallOption) (*userpb.UserSpecifier, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserSpecifier), args.Error(1)
}

func (m *MockAccountClient) SigninUser(ctx context.Context, req *userpb.SigninUserInfo, opts ...grpc.CallOption) (*userpb.UserSpecifier, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserSpecifier), args.Error(1)
}

func (m *MockAccountClient) CheckEmailAvailability(ctx context.Context, req *userpb.CheckEmailRequest, opts ...grpc.CallOption) (*userpb.CheckEmailResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.CheckEmailResponse), args.Error(1)
}

func (m *MockAccountClient) UpdatePasswordByEmail(ctx context.Context, req *userpb.UpdatePasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockAccountClient) GetProfile(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*userpb.Profile, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.Profile), args.Error(1)
}

func (m *MockAccountClient) AddFunds(ctx context.Context, req *userpb.AddFundsRequest, opts ...grpc.CallOption) (*userpb.AddFundsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.AddFundsResponse), args.Error(1)
}

func (m *MockAccountClient) ResetAccount(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockAccountClient) DeleteAccount(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockAccountClient) ReserveForOrder(ctx context.Context, req *userpb.ReserveForOrderRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockAccountClient) ReleaseForOrder(ctx context.Context, req *userpb.ReleaseForOrderRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}
