package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"google.golang.org/grpc"
)

type MockPortfolioClient struct {
	mock.Mock
}

func (m *MockPortfolioClient) GetPortfolio(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*userpb.PortfolioResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.PortfolioResponse), args.Error(1)
}

func (m *MockPortfolioClient) GetHoldings(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*userpb.HoldingsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.HoldingsResponse), args.Error(1)
}

func (m *MockPortfolioClient) GetPositions(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*userpb.PositionsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.PositionsResponse), args.Error(1)
}

func (m *MockPortfolioClient) GetPortfolioHistory(ctx context.Context, req *userpb.PortfolioHistoryRequest, opts ...grpc.CallOption) (*userpb.PortfolioHistoryResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.PortfolioHistoryResponse), args.Error(1)
}
