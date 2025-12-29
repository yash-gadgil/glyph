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

func (s *PortfolioHandler) GetHoldings(ctx context.Context, req *userpb.UserSpecifier) (*userpb.HoldingsResponse, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("get_holdings"))

	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	positions, err := s.q.GetPositionsForUser(ctx, userUUID)
	if err != nil {
		log.Error("get_holdings_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to load positions")
	}

	resp := &userpb.HoldingsResponse{}
	for _, p := range positions {
		if p.Qty == 0 && p.RealizedPnl == 0 {
			continue
		}

		holding := &userpb.Holding{
			Symbol:           p.Symbol,
			Qty:              p.Qty,
			CostBasisCents:   p.CostBasis,
			RealizedPnlCents: p.RealizedPnl,
		}
		if p.Qty != 0 {
			holding.AvgPriceCents = p.CostBasis / p.Qty
		}

		holding.LastPriceCents = holding.AvgPriceCents
		holding.MarketValueCents = p.CostBasis
		holding.UnrealizedPnlCents = holding.MarketValueCents - p.CostBasis

		resp.Holdings = append(resp.Holdings, holding)
		resp.TotalMarketValueCents += holding.MarketValueCents
		resp.TotalCostBasisCents += p.CostBasis
		resp.TotalUnrealizedPnlCents += holding.UnrealizedPnlCents
		resp.TotalRealizedPnlCents += p.RealizedPnl
	}

	return resp, nil
}

func (s *PortfolioHandler) GetPositions(ctx context.Context, req *userpb.UserSpecifier) (*userpb.PositionsResponse, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("get_positions"))

	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	res, err := s.q.GetPositionsForUser(ctx, userUUID)
	if err != nil {
		log.Error("get_positions_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get positions")
	}

	positions := make([]*userpb.Position, 0, len(res))
	for _, pos := range res {
		p := &userpb.Position{
			Symbol:           pos.Symbol,
			Qty:              pos.Qty,
			ReservedQty:      pos.ReservedQty,
			RealizedPnlCents: pos.RealizedPnl,
			CostBasisCents:   pos.CostBasis,
		}
		if pos.Qty != 0 {
			p.AvgPriceCents = pos.CostBasis / pos.Qty
		}
		positions = append(positions, p)
	}

	return &userpb.PositionsResponse{Positions: positions}, nil
}
