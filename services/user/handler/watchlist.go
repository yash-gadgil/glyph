package handler

import (
	"context"

	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"github.com/yash-gadgil/glyph/services/user/types"
	"google.golang.org/grpc"
)

type WatchlistsHandler struct {
	watchlistService types.WatchlistService
	userpb.UnimplementedWatchlistServiceServer
}

func NewGrpcWatchlistService(grpc *grpc.Server, watchlistService types.WatchlistService) {

	handler := &WatchlistsHandler{
		watchlistService: watchlistService,
	}
	userpb.RegisterWatchlistServiceServer(grpc, handler)
}

func (h *WatchlistsHandler) GetWatchlists(ctx context.Context, req *userpb.WatchlistsRequest) (*userpb.WatchlistsResponse, error) {

	h.watchlistService.GetWatchlists(ctx, req)

	return &userpb.WatchlistsResponse{
		UserID: 634,
		WMetadata: []*userpb.WatchlistMetadata{{
			Id:   3,
			Name: "Watchlist 1",
		}},
		First: &userpb.Watchlist{
			UserID:  634,
			Id:      1,
			Name:    "Watchlist 1",
			Symbols: []string{"AAPL", "SPY"},
		},
	}, nil
}

func (h *WatchlistsHandler) GetWatchlist(ctx context.Context, req *userpb.WatchlistRequest) (*userpb.Watchlist, error) {
	return &userpb.Watchlist{}, nil
}
func (h *WatchlistsHandler) CreateWatchList(ctx context.Context, req *userpb.CreateWatchlistRequest) (*userpb.CreateWatchlistResponse, error) {
	return &userpb.CreateWatchlistResponse{}, nil
}
