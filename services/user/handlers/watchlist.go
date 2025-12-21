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
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *WatchlistHandler) GetWatchlist(ctx context.Context, req *userpb.WatchlistSpecifier) (*userpb.Watchlist, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	watchlistUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid watchlist ID")
	}

	row, err := s.q.GetWatchlist(ctx, db.GetWatchlistParams{
		ID:     watchlistUUID,
		UserID: userUUID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "watchlist not found")
		}
		s.log.Error("get_watchlist_failed", logger.Action("get_watchlist"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get watchlist")
	}

	return &userpb.Watchlist{
		Id:      req.Id,
		UserId:  row.UserID.String(),
		Name:    row.WName,
		Symbols: parseSymbolArray(row.Symbols),
	}, nil
}

func (s *WatchlistHandler) CreateWatchlist(ctx context.Context, req *userpb.CreateWatchlistRequest) (*userpb.WatchlistSpecifier, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "watchlist name is required")
	}

	name := strings.TrimSpace(*req.Name)
	watchlistID, err := s.q.CreateWatchlist(ctx, db.CreateWatchlistParams{
		WName:  name,
		UserID: userUUID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, status.Errorf(codes.AlreadyExists, "a watchlist named %q already exists", name)
		}
		s.log.Error("create_watchlist_failed", logger.Action("create_watchlist"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create watchlist")
	}

	return &userpb.WatchlistSpecifier{
		Id:     watchlistID.String(),
		UserId: req.UserId,
	}, nil
}

func (s *WatchlistHandler) ModifyWatchlist(ctx context.Context, req *userpb.ModifyWatchlistRequest) (*emptypb.Empty, error) {
	watchlistUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid watchlist ID")
	}
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	symbols := make([]string, 0, len(req.Symbols))
	for _, sym := range req.Symbols {
		sym = strings.ToUpper(strings.TrimSpace(sym))
		if sym == "" || len(sym) > 10 {
			return nil, status.Errorf(codes.InvalidArgument, "invalid symbol %q", sym)
		}
		symbols = append(symbols, sym)
	}
	if len(symbols) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no symbols provided")
	}

	switch req.Action {
	case userpb.ModifyWatchlistRequest_SUBSCRIBE:
		if _, err := s.q.AddSymbols(ctx, db.AddSymbolsParams{
			ID:      watchlistUUID,
			UserID:  userUUID,
			Column3: symbols,
		}); err != nil {
			s.log.Error("add_symbols_failed", logger.Action("modify_watchlist"), zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to add symbols")
		}
	case userpb.ModifyWatchlistRequest_UNSUBSCRIBE:
		if _, err := s.q.RemoveSymbols(ctx, db.RemoveSymbolsParams{
			ID:      watchlistUUID,
			UserID:  userUUID,
			Column3: symbols,
		}); err != nil {
			s.log.Error("remove_symbols_failed", logger.Action("modify_watchlist"), zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to remove symbols")
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported modify action")
	}

	return &emptypb.Empty{}, nil
}
