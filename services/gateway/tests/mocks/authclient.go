package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Signup(ctx context.Context, req *authpb.SignupRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockAuthClient) Signin(ctx context.Context, req *authpb.SigninRequest, opts ...grpc.CallOption) (*authpb.TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) OAuthURL(ctx context.Context, req *authpb.OAuthURLRequest, opts ...grpc.CallOption) (*authpb.OAuthURLResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.OAuthURLResponse), args.Error(1)
}

func (m *MockAuthClient) OAuthCallback(ctx context.Context, req *authpb.OAuthCallbackRequest, opts ...grpc.CallOption) (*authpb.TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) VerifyEmail(ctx context.Context, req *authpb.VerificationRequest, opts ...grpc.CallOption) (*authpb.TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) VerifyToken(ctx context.Context, req *authpb.VerificationRequest, opts ...grpc.CallOption) (*authpb.VerificationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.VerificationResponse), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest, opts ...grpc.CallOption) (*authpb.TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) GetPublicKeys(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*authpb.GetPublicKeysResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.GetPublicKeysResponse), args.Error(1)
}

func (m *MockAuthClient) ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockAuthClient) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest, opts ...grpc.CallOption) (*authpb.TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.TokenResponse), args.Error(1)
}
