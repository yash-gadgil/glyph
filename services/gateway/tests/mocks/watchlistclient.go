package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockWatchlistClient struct {
	mock.Mock
}

func (m *MockWatchlistClient) GetWatchlists(ctx context.Context, req *userpb.UserSpecifier, opts ...grpc.CallOption) (*userpb.WatchlistsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.WatchlistsResponse), args.Error(1)
}

func (m *MockWatchlistClient) GetWatchlist(ctx context.Context, req *userpb.WatchlistSpecifier, opts ...grpc.CallOption) (*userpb.Watchlist, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.Watchlist), args.Error(1)
}

func (m *MockWatchlistClient) CreateWatchlist(ctx context.Context, req *userpb.CreateWatchlistRequest, opts ...grpc.CallOption) (*userpb.WatchlistSpecifier, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.WatchlistSpecifier), args.Error(1)
}

func (m *MockWatchlistClient) ModifyWatchlist(ctx context.Context, req *userpb.ModifyWatchlistRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockWatchlistClient) DeleteWatchlist(ctx context.Context, req *userpb.WatchlistSpecifier, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockWatchlistClient) DeleteSymbolFromWatchlist(ctx context.Context, req *userpb.DeleteSymbolRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}
