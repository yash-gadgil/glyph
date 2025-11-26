package handlers

import (
	"database/sql"

	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
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
