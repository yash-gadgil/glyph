package handlers

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WatchlistHandler struct {
	userpb.UnimplementedWatchlistServiceServer
	db  *sql.DB
	q   *db.Queries
	log *zap.Logger
}

func NewWatchlistHandler(sdb *sql.DB, log *zap.Logger) *WatchlistHandler {
	return &WatchlistHandler{db: sdb, q: db.New(sdb), log: log}
}

func parseSymbolArray(v interface{}) []string {
	b, ok := v.([]byte)
	if !ok {
		return []string{}
	}

	st := strings.TrimSuffix(strings.TrimPrefix(string(b), "{"), "}")
	if st == "" {
		return []string{}
	}

	return strings.Split(st, ",")
}

func (s *WatchlistHandler) GetWatchlists(ctx context.Context, req *userpb.UserSpecifier) (*userpb.WatchlistsResponse, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("get_watchlists"))

	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	metadata, err := s.q.GetWatchlistsMetadata(ctx, userUUID)
	if err != nil {
		log.Error("watchlist_metadata_fetch_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get watchlists")
	}

	wmeta := make([]*userpb.WatchlistMetadata, 0, len(metadata))
	for _, m := range metadata {
		wmeta = append(wmeta, &userpb.WatchlistMetadata{
			Id:   m.ID.String(),
			Name: m.WName,
		})
	}

	var first *userpb.Watchlist
	if len(metadata) > 0 {
		row, err := s.q.GetWatchlist(ctx, db.GetWatchlistParams{
			ID:     metadata[0].ID,
			UserID: userUUID,
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Error("first_watchlist_fetch_failed", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to load first watchlist")
		}
		if err == nil {
			first = &userpb.Watchlist{
				UserId:  req.UserId,
				Id:      metadata[0].ID.String(),
				Name:    metadata[0].WName,
				Symbols: parseSymbolArray(row.Symbols),
			}
		}
	}

	return &userpb.WatchlistsResponse{
		UserId:    req.UserId,
		WMetadata: wmeta,
		First:     first,
	}, nil
}
