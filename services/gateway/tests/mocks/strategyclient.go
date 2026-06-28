package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockStrategyClient struct {
	mock.Mock
}

func (m *MockStrategyClient) GetStrategies(ctx context.Context, req *strategypb.UserSpecifier, opts ...grpc.CallOption) (*strategypb.StrategiesResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strategypb.StrategiesResponse), args.Error(1)
}

func (m *MockStrategyClient) CreateStrategy(ctx context.Context, req *strategypb.CreateStrategyRequest, opts ...grpc.CallOption) (*strategypb.Strategy, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strategypb.Strategy), args.Error(1)
}

func (m *MockStrategyClient) UpdateStrategy(ctx context.Context, req *strategypb.UpdateStrategyRequest, opts ...grpc.CallOption) (*strategypb.Strategy, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strategypb.Strategy), args.Error(1)
}

func (m *MockStrategyClient) DeleteStrategy(ctx context.Context, req *strategypb.StrategySpecifier, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockStrategyClient) DeployStrategy(ctx context.Context, req *strategypb.DeployStrategyRequest, opts ...grpc.CallOption) (*strategypb.Deployment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strategypb.Deployment), args.Error(1)
}

func (m *MockStrategyClient) StopDeployment(ctx context.Context, req *strategypb.DeploymentSpecifier, opts ...grpc.CallOption) (*strategypb.Deployment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strategypb.Deployment), args.Error(1)
}

func (m *MockStrategyClient) DeleteDeployment(ctx context.Context, req *strategypb.DeploymentSpecifier, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockStrategyClient) GetDeployments(ctx context.Context, req *strategypb.UserSpecifier, opts ...grpc.CallOption) (*strategypb.DeploymentsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strategypb.DeploymentsResponse), args.Error(1)
}

func (m *MockStrategyClient) RunBacktest(ctx context.Context, req *strategypb.BacktestRequest, opts ...grpc.CallOption) (*strategypb.BacktestResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strategypb.BacktestResponse), args.Error(1)
}
