package types

import (
	"context"

	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
)

type WatchlistService interface {
	GetWatchlists(ctx context.Context, req *userpb.WatchlistsRequest) error

	GetWatchlist(ctx context.Context, req *userpb.WatchlistRequest) error

	CreateWatchlist(ctx context.Context, req *userpb.CreateWatchlistRequest) error
}
