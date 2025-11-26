package handlers

import (
	"database/sql"

	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
)

type PortfolioHandler struct {
	userpb.UnimplementedPortfolioServiceServer
	db  *sql.DB
	q   *db.Queries
	log *zap.Logger
}

func NewPortfolioHandler(sdb *sql.DB, log *zap.Logger) *PortfolioHandler {
	return &PortfolioHandler{db: sdb, q: db.New(sdb), log: log}
}
