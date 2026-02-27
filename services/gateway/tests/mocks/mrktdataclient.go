package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockMrktdataClient struct {
	mock.Mock
}

func (m *MockMrktdataClient) GetHistoricalStockData(ctx context.Context, req *mrktpb.HistoricalStockDataRequest, opts ...grpc.CallOption) (*mrktpb.HistoricalStockDataResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mrktpb.HistoricalStockDataResponse), args.Error(1)
}

func (m *MockMrktdataClient) WatchlistStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[mrktpb.WatchlistStreamRequest, mrktpb.MarketUpdate], error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(grpc.BidiStreamingClient[mrktpb.WatchlistStreamRequest, mrktpb.MarketUpdate]), args.Error(1)
}

func (m *MockMrktdataClient) GetAvailableSymbols(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*mrktpb.AvailableSymbolsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mrktpb.AvailableSymbolsResponse), args.Error(1)
}

func (m *MockMrktdataClient) GetLatestPrices(ctx context.Context, req *mrktpb.LatestPricesRequest, opts ...grpc.CallOption) (*mrktpb.LatestPricesResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mrktpb.LatestPricesResponse), args.Error(1)
}

func (m *MockMrktdataClient) GetNews(ctx context.Context, req *mrktpb.NewsRequest, opts ...grpc.CallOption) (*mrktpb.NewsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mrktpb.NewsResponse), args.Error(1)
}

func (m *MockMrktdataClient) GetMovers(ctx context.Context, req *mrktpb.MoversRequest, opts ...grpc.CallOption) (*mrktpb.MoversResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mrktpb.MoversResponse), args.Error(1)
}
