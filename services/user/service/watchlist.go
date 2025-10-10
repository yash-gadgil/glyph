package service

import (
	"context"
	"fmt"

	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	_ "github.com/yash-gadgil/glyph/services/user/types"
)

type WatchlistService struct {
}

func NewWatchlistService() *WatchlistService {
	return &WatchlistService{}
}

func (s *WatchlistService) GetWatchlists(ctx context.Context, req *userpb.WatchlistsRequest) error {

	fmt.Println("Gotten watchlists of user:", req.UserID)

	return nil
}

func (s *WatchlistService) GetWatchlist(ctx context.Context, req *userpb.WatchlistRequest) error {
	return nil
}
func (s *WatchlistService) CreateWatchlist(ctx context.Context, req *userpb.CreateWatchlistRequest) error {
	return nil
}
