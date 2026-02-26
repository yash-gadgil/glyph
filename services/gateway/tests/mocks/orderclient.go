package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	"google.golang.org/grpc"
)

type MockOrderClient struct {
	mock.Mock
}

func (m *MockOrderClient) PlaceOrder(ctx context.Context, req *ordrpb.PlaceOrderRequest, opts ...grpc.CallOption) (*ordrpb.PlaceOrderResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordrpb.PlaceOrderResponse), args.Error(1)
}

func (m *MockOrderClient) CancelOrder(ctx context.Context, req *ordrpb.CancelOrderRequest, opts ...grpc.CallOption) (*ordrpb.CancelOrderResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordrpb.CancelOrderResponse), args.Error(1)
}

func (m *MockOrderClient) GetOrders(ctx context.Context, req *ordrpb.GetOrdersRequest, opts ...grpc.CallOption) (*ordrpb.GetOrdersResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordrpb.GetOrdersResponse), args.Error(1)
}

func (m *MockOrderClient) GetOrder(ctx context.Context, req *ordrpb.GetOrderRequest, opts ...grpc.CallOption) (*ordrpb.Order, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordrpb.Order), args.Error(1)
}

func (m *MockOrderClient) UpdateOrderStatus(ctx context.Context, req *ordrpb.UpdateOrderStatusRequest, opts ...grpc.CallOption) (*ordrpb.UpdateOrderStatusResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordrpb.UpdateOrderStatusResponse), args.Error(1)
}

func (m *MockOrderClient) GetFills(ctx context.Context, req *ordrpb.GetFillsRequest, opts ...grpc.CallOption) (*ordrpb.GetFillsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordrpb.GetFillsResponse), args.Error(1)
}

func (m *MockOrderClient) GetStrategyFills(ctx context.Context, req *ordrpb.GetStrategyFillsRequest, opts ...grpc.CallOption) (*ordrpb.GetFillsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordrpb.GetFillsResponse), args.Error(1)
}
