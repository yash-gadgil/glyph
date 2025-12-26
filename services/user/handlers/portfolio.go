package handlers

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *PortfolioHandler) GetPortfolio(ctx context.Context, req *userpb.UserSpecifier) (*userpb.PortfolioResponse, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("get_portfolio"))

	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	res, err := s.q.GetPortfolio(ctx, userUUID)
	if err != nil {
		log.Error("get_portfolio_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get portfolio")
	}

	return &userpb.PortfolioResponse{
		CashBalanceCents:  res.CashBalance,
		ReservedCashCents: res.ReservedCash,
		Currency:          res.Currency,
		Multiplier:        res.Multiplier,
		MarginUsedCents:   res.MarginUsed,
	}, nil
}
