package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockStrategyClient struct {
	mock.Mock
}

func (m *MockStrategyClient) GetStrategies(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*userpb.StrategiesResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.StrategiesResponse), args.Error(1)
}

func (m *MockStrategyClient) CreateStrategy(ctx context.Context, req *userpb.CreateStrategyRequest, opts ...grpc.CallOption) (*userpb.Strategy, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.Strategy), args.Error(1)
}

func (m *MockStrategyClient) UpdateStrategy(ctx context.Context, req *userpb.UpdateStrategyRequest, opts ...grpc.CallOption) (*userpb.Strategy, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.Strategy), args.Error(1)
}

func (m *MockStrategyClient) DeleteStrategy(ctx context.Context, req *userpb.StrategySpecifier, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockStrategyClient) DeployStrategy(ctx context.Context, req *userpb.DeployStrategyRequest, opts ...grpc.CallOption) (*userpb.Deployment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.Deployment), args.Error(1)
}

func (m *MockStrategyClient) StopDeployment(ctx context.Context, req *userpb.DeploymentSpecifier, opts ...grpc.CallOption) (*userpb.Deployment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.Deployment), args.Error(1)
}

func (m *MockStrategyClient) DeleteDeployment(ctx context.Context, req *userpb.DeploymentSpecifier, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockStrategyClient) GetDeployments(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*userpb.DeploymentsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.DeploymentsResponse), args.Error(1)
}

func (m *MockStrategyClient) RunBacktest(ctx context.Context, req *userpb.BacktestRequest, opts ...grpc.CallOption) (*userpb.BacktestResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.BacktestResponse), args.Error(1)
}
