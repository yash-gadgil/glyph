package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	db "github.com/yash-gadgil/glyph/services/strategy/db/gen"
	"github.com/yash-gadgil/glyph/services/strategy/engine"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxStrategyConfigBytes = 64 * 1024

const minPositionSizeCents = 100

func deploymentStatus(status int16) string {
	if status == 0 {
		return "running"
	}
	return "stopped"
}

func deploymentToProto(id, strategyID, userID uuid.UUID, symbol string, sizeCents int64, status int16, inPosition bool, entryCents, qty int64, name string, createdAt, updatedAt time.Time) *strategypb.Deployment {
	return &strategypb.Deployment{
		Id:                id.String(),
		StrategyId:        strategyID.String(),
		UserId:            userID.String(),
		Symbol:            symbol,
		PositionSizeCents: sizeCents,
		Status:            deploymentStatus(status),
		InPosition:        inPosition,
		EntryPriceCents:   entryCents,
		Qty:               qty,
		StrategyName:      name,
		CreatedAt:         createdAt.Format(time.RFC3339),
		UpdatedAt:         updatedAt.Format(time.RFC3339),
	}
}

type StrategyHandler struct {
	strategypb.UnimplementedStrategyServiceServer
	db   *sql.DB
	q    *db.Queries
	mrkt mrktpb.MrktdataServiceClient
	log  *zap.Logger
}

func NewStrategyHandler(sdb *sql.DB, mrkt mrktpb.MrktdataServiceClient, log *zap.Logger) *StrategyHandler {
	return &StrategyHandler{db: sdb, q: db.New(sdb), mrkt: mrkt, log: log}
}

func validateStrategyInput(name, configJSON string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return status.Errorf(codes.InvalidArgument, "strategy name is required")
	}
	if len(name) > 80 {
		return status.Errorf(codes.InvalidArgument, "strategy name must be 80 characters or fewer")
	}
	if len(configJSON) > maxStrategyConfigBytes {
		return status.Errorf(codes.InvalidArgument, "strategy config too large")
	}
	if !json.Valid([]byte(configJSON)) {
		return status.Errorf(codes.InvalidArgument, "strategy config must be valid JSON")
	}
	return nil
}

func strategyToProto(s db.Strategy) *strategypb.Strategy {
	return &strategypb.Strategy{
		Id:         s.ID.String(),
		UserId:     s.UserID.String(),
		Name:       s.Name,
		ConfigJson: string(s.Config),
		CreatedAt:  s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  s.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *StrategyHandler) GetStrategies(ctx context.Context, req *strategypb.UserSpecifier) (*strategypb.StrategiesResponse, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	rows, err := s.q.GetStrategiesForUser(ctx, userUUID)
	if err != nil {
		s.log.Error("strategies_fetch_failed", logger.Action("get_strategies"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to load strategies")
	}

	resp := &strategypb.StrategiesResponse{}
	for _, row := range rows {
		resp.Strategies = append(resp.Strategies, strategyToProto(row))
	}
	return resp, nil
}

func (s *StrategyHandler) CreateStrategy(ctx context.Context, req *strategypb.CreateStrategyRequest) (*strategypb.Strategy, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	if err := validateStrategyInput(req.Name, req.ConfigJson); err != nil {
		return nil, err
	}

	row, err := s.q.CreateStrategy(ctx, db.CreateStrategyParams{
		UserID: userUUID,
		Name:   strings.TrimSpace(req.Name),
		Config: json.RawMessage(req.ConfigJson),
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, status.Errorf(codes.AlreadyExists, "a strategy named %q already exists", req.Name)
		}
		s.log.Error("strategy_create_failed", logger.Action("create_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create strategy")
	}

	return strategyToProto(row), nil
}

func (s *StrategyHandler) UpdateStrategy(ctx context.Context, req *strategypb.UpdateStrategyRequest) (*strategypb.Strategy, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	strategyUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid strategy ID")
	}
	if err := validateStrategyInput(req.Name, req.ConfigJson); err != nil {
		return nil, err
	}

	row, err := s.q.UpdateStrategy(ctx, db.UpdateStrategyParams{
		ID:     strategyUUID,
		UserID: userUUID,
		Name:   strings.TrimSpace(req.Name),
		Config: json.RawMessage(req.ConfigJson),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "strategy not found")
		}
		s.log.Error("strategy_update_failed", logger.Action("update_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update strategy")
	}

	return strategyToProto(row), nil
}

func (s *StrategyHandler) DeleteStrategy(ctx context.Context, req *strategypb.StrategySpecifier) (*emptypb.Empty, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	strategyUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid strategy ID")
	}

	if err := s.q.DeleteStrategy(ctx, db.DeleteStrategyParams{
		ID:     strategyUUID,
		UserID: userUUID,
	}); err != nil {
		s.log.Error("strategy_delete_failed", logger.Action("delete_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete strategy")
	}

	return &emptypb.Empty{}, nil
}

func (s *StrategyHandler) DeployStrategy(ctx context.Context, req *strategypb.DeployStrategyRequest) (*strategypb.Deployment, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	strategyUUID, err := uuid.Parse(req.StrategyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid strategy ID")
	}
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	if symbol == "" || len(symbol) > 10 {
		return nil, status.Errorf(codes.InvalidArgument, "a valid symbol is required")
	}
	if req.PositionSizeCents < minPositionSizeCents {
		return nil, status.Errorf(codes.InvalidArgument, "position size must be at least $1")
	}

	strat, err := s.q.GetStrategy(ctx, db.GetStrategyParams{ID: strategyUUID, UserID: userUUID})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "strategy not found")
	}
	if _, err := engine.ParseConfig(strat.Config); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "strategy has no runnable entry rules, edit it first")
	}

	existing, err := s.q.GetLatestDeploymentForSymbol(ctx, db.GetLatestDeploymentForSymbolParams{
		UserID:     userUUID,
		StrategyID: strategyUUID,
		Symbol:     symbol,
	})
	if err == nil {
		row := existing
		if existing.Status != 0 {
			row, err = s.q.ReactivateDeployment(ctx, db.ReactivateDeploymentParams{
				ID:                existing.ID,
				PositionSizeCents: req.PositionSizeCents,
			})
			if err != nil {
				s.log.Error("deployment_reactivate_failed", logger.Action("deploy_strategy"), zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to deploy strategy")
			}
			telemetry.StrategyDeploymentsTotal.Inc()
		}
		return deploymentToProto(row.ID, row.StrategyID, row.UserID, row.Symbol, row.PositionSizeCents,
			row.Status, row.InPosition, row.EntryPriceCents, row.Qty, strat.Name, row.CreatedAt, row.UpdatedAt), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		s.log.Error("deployment_lookup_failed", logger.Action("deploy_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to deploy strategy")
	}

	row, err := s.q.CreateDeployment(ctx, db.CreateDeploymentParams{
		UserID:            userUUID,
		StrategyID:        strategyUUID,
		Symbol:            symbol,
		PositionSizeCents: req.PositionSizeCents,
	})
	if err != nil {
		s.log.Error("deployment_create_failed", logger.Action("deploy_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to deploy strategy")
	}

	telemetry.StrategyDeploymentsTotal.Inc()

	return deploymentToProto(row.ID, row.StrategyID, row.UserID, row.Symbol, row.PositionSizeCents,
		row.Status, row.InPosition, row.EntryPriceCents, row.Qty, strat.Name, row.CreatedAt, row.UpdatedAt), nil
}

func (s *StrategyHandler) StopDeployment(ctx context.Context, req *strategypb.DeploymentSpecifier) (*strategypb.Deployment, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	depUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid deployment ID")
	}

	row, err := s.q.StopDeployment(ctx, db.StopDeploymentParams{ID: depUUID, UserID: userUUID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "deployment not found")
		}
		s.log.Error("deployment_stop_failed", logger.Action("stop_deployment"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to stop deployment")
	}

	return deploymentToProto(row.ID, row.StrategyID, row.UserID, row.Symbol, row.PositionSizeCents,
		row.Status, row.InPosition, row.EntryPriceCents, row.Qty, "", row.CreatedAt, row.UpdatedAt), nil
}

func (s *StrategyHandler) DeleteDeployment(ctx context.Context, req *strategypb.DeploymentSpecifier) (*emptypb.Empty, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	depUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid deployment ID")
	}

	rows, err := s.q.DeleteDeployment(ctx, db.DeleteDeploymentParams{ID: depUUID, UserID: userUUID})
	if err != nil {
		s.log.Error("deployment_delete_failed", logger.Action("delete_deployment"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete deployment")
	}
	if rows == 0 {
		return nil, status.Errorf(codes.NotFound, "deployment not found or still running, stop it first")
	}

	return &emptypb.Empty{}, nil
}

func (s *StrategyHandler) GetDeployments(ctx context.Context, req *strategypb.UserSpecifier) (*strategypb.DeploymentsResponse, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	rows, err := s.q.GetDeploymentsForUser(ctx, userUUID)
	if err != nil {
		s.log.Error("deployments_fetch_failed", logger.Action("get_deployments"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to load deployments")
	}

	resp := &strategypb.DeploymentsResponse{}
	for _, row := range rows {
		resp.Deployments = append(resp.Deployments, deploymentToProto(
			row.ID, row.StrategyID, row.UserID, row.Symbol, row.PositionSizeCents,
			row.Status, row.InPosition, row.EntryPriceCents, row.Qty, row.StrategyName,
			row.CreatedAt, row.UpdatedAt))
	}
	return resp, nil
}

func backtestTimeframe(tf string) (mrktpb.Timeframe, time.Duration, error) {
	switch strings.ToUpper(strings.TrimSpace(tf)) {
	case "", "DAY":
		return mrktpb.Timeframe_DAY, 2 * 365 * 24 * time.Hour, nil
	case "HOUR":
		return mrktpb.Timeframe_HOUR, 30 * 24 * time.Hour, nil
	case "MIN":
		return mrktpb.Timeframe_MIN, 7 * 24 * time.Hour, nil
	default:
		return 0, 0, status.Errorf(codes.InvalidArgument, "unknown timeframe %q (want DAY, HOUR, or MIN)", tf)
	}
}

func parseBacktestDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), true
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}

func dateProto(t time.Time) *mrktpb.Date {
	return &mrktpb.Date{
		Year:  int32(t.Year()),
		Month: int32(t.Month()),
		Day:   int32(t.Day()),
		Hour:  int32(t.Hour()),
		Min:   int32(t.Minute()),
	}
}

func warmupSpan(tf mrktpb.Timeframe, warmup int) time.Duration {
	n := time.Duration(warmup)
	switch tf {
	case mrktpb.Timeframe_MIN:
		return n*8*time.Minute + 5*24*time.Hour
	case mrktpb.Timeframe_HOUR:
		return n*8*time.Hour + 5*24*time.Hour
	default: // DAY
		return n*48*time.Hour + 10*24*time.Hour
	}
}

func trimToWindow(bars []engine.Bar, start time.Time, warmup int) []engine.Bar {
	startIdx := 0
	for startIdx < len(bars) && bars[startIdx].Time.Before(start) {
		startIdx++
	}
	lo := startIdx - warmup
	if lo < 0 {
		lo = 0
	}
	return bars[lo:]
}

func barsFromProto(resp *mrktpb.HistoricalStockDataResponse, symbol string) []engine.Bar {
	var src []*mrktpb.Bar
	for _, sb := range resp.GetSymbolBars() {
		if sb.Symbol == symbol {
			src = sb.Bars
			break
		}
	}
	if src == nil && len(resp.GetSymbolBars()) > 0 {
		src = resp.GetSymbolBars()[0].Bars
	}

	out := make([]engine.Bar, 0, len(src))
	for _, b := range src {
		t, _ := time.Parse(time.RFC3339, b.Time)
		out = append(out, engine.Bar{
			Time:   t,
			Open:   float64(b.Open),
			High:   float64(b.High),
			Low:    float64(b.Low),
			Close:  float64(b.Close),
			Volume: float64(b.Volume),
			VWAP:   float64(b.Vwap),
		})
	}
	return out
}

func backtestResultToProto(r *engine.BacktestResult) *strategypb.BacktestResponse {
	out := &strategypb.BacktestResponse{
		TotalReturnPct:   r.TotalReturnPct,
		MaxDrawdownPct:   r.MaxDrawdownPct,
		Sharpe:           r.Sharpe,
		WinRate:          r.WinRate,
		ProfitFactor:     r.ProfitFactor,
		NumTrades:        int32(r.NumTrades),
		AvgHoldBars:      r.AvgHoldBars,
		FinalEquityCents: r.FinalEquityCents,
		BarsUsed:         int32(r.BarsUsed),
		WarmupBars:       int32(r.WarmupBars),
	}
	for _, p := range r.EquityCurve {
		out.EquityCurve = append(out.EquityCurve, &strategypb.BacktestEquityPoint{
			TimeUnix:    p.Time.Unix(),
			EquityCents: p.EquityCents,
		})
	}
	for _, t := range r.Trades {
		out.Trades = append(out.Trades, &strategypb.BacktestTrade{
			EntryTimeUnix:   t.EntryTime.Unix(),
			ExitTimeUnix:    t.ExitTime.Unix(),
			EntryPriceCents: t.EntryPriceCents,
			ExitPriceCents:  t.ExitPriceCents,
			Qty:             t.Qty,
			PnlCents:        t.PnLCents,
			ReturnPct:       t.ReturnPct,
			HoldBars:        int32(t.HoldBars),
			ExitReason:      t.ExitReason,
		})
	}
	return out
}

func (s *StrategyHandler) RunBacktest(ctx context.Context, req *strategypb.BacktestRequest) (*strategypb.BacktestResponse, error) {
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	if symbol == "" {
		return nil, status.Errorf(codes.InvalidArgument, "symbol is required")
	}
	if req.InitialCapitalCents <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "initial_capital_cents must be positive")
	}
	if req.PositionSizeCents <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "position_size_cents must be positive")
	}

	cfg, err := engine.ParseConfig([]byte(req.ConfigJson))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid strategy config: %v", err)
	}

	tf, maxSpan, err := backtestTimeframe(req.Timeframe)
	if err != nil {
		return nil, err
	}

	end, ok := parseBacktestDate(req.End)
	if !ok {
		end = time.Now().UTC()
	}
	start, ok := parseBacktestDate(req.Start)
	if !ok {
		start = end.Add(-maxSpan)
	}
	if !start.Before(end) {
		return nil, status.Errorf(codes.InvalidArgument, "start must be before end")
	}
	if end.Sub(start) > maxSpan {
		return nil, status.Errorf(codes.InvalidArgument,
			"range too large for %s timeframe (max %d days)", tf.String(), int(maxSpan.Hours()/24))
	}

	if s.mrkt == nil {
		return nil, status.Errorf(codes.Unavailable, "market data service unavailable")
	}

	warmup := engine.MaxLookback(cfg)

	resp, err := s.mrkt.GetHistoricalStockData(ctx, &mrktpb.HistoricalStockDataRequest{
		Symbols:   []string{symbol},
		Timeframe: tf,
		Start:     dateProto(start.Add(-warmupSpan(tf, warmup))),
		End:       dateProto(end),
	})
	if err != nil {
		s.log.Error("backtest_bars_fetch_failed", logger.Action("run_backtest"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to load historical data")
	}

	bars := barsFromProto(resp, symbol)
	bars = trimToWindow(bars, start, warmup)
	if len(bars) <= warmup {
		return nil, status.Errorf(codes.InvalidArgument,
			"not enough bars (%d) for this strategy's warm-up (%d), widen the date range or use a longer timeframe",
			len(bars), warmup)
	}

	result, err := engine.RunBacktest(cfg, bars, engine.BacktestParams{
		InitialCapitalCents: req.InitialCapitalCents,
		PositionSizeCents:   req.PositionSizeCents,
		Timeframe:           tf.String(),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	telemetry.BacktestsTotal.Inc()

	return backtestResultToProto(result), nil
}
