package handlers

import (
	"database/sql"

	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
)

type AccountHandler struct {
	userpb.UnimplementedAccountServiceServer
	db  *sql.DB
	q   *db.Queries
	log *zap.Logger
}

func NewAccountHandler(sdb *sql.DB, log *zap.Logger) *AccountHandler {
	return &AccountHandler{db: sdb, q: db.New(sdb), log: log}
}
