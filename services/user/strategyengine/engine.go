package strategyengine

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
)

const tickInterval = time.Minute

type Engine struct {
	q     *db.Queries
	mrkt  mrktpb.MrktdataServiceClient
	order ordrpb.OrderServiceClient
	log   *zap.Logger
}

func NewEngine(q *db.Queries, mrkt mrktpb.MrktdataServiceClient, order ordrpb.OrderServiceClient, log *zap.Logger) *Engine {
	return &Engine{q: q, mrkt: mrkt, order: order, log: log}
}

func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	e.log.Info("strategy_engine_started")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !marketOpen(time.Now()) {
				continue
			}
			e.tick(ctx)
		}
	}
}

func (e *Engine) tick(ctx context.Context) {
	tickCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	deployments, err := e.q.GetRunningDeployments(tickCtx)
	if err != nil {
		e.log.Error("engine_deployments_query_failed", zap.Error(err))
		return
	}
	if len(deployments) == 0 {
		return
	}

	symbols := uniqueSymbols(deployments)
	var (
		wg     sync.WaitGroup
		bars   map[string][]Bar
		prices map[string]int64
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		bars = e.fetchBars(tickCtx, symbols)
	}()
	go func() {
		defer wg.Done()
		prices = e.fetchLatestPrices(tickCtx, symbols)
	}()
	wg.Wait()

	for _, dep := range deployments {
		e.evaluateDeployment(tickCtx, dep, bars[dep.Symbol], prices[dep.Symbol])
	}
}

func (e *Engine) evaluateDeployment(ctx context.Context, dep db.GetRunningDeploymentsRow, bars []Bar, latestCents int64) {
	telemetry.StrategyTicksTotal.Inc()

	log := e.log.With(
		logger.Action("strategy_tick"),
		logger.KV("deployment_id", dep.ID.String()),
		logger.KV("symbol", dep.Symbol),
	)

	cfg, err := ParseConfig(dep.StrategyConfig)
	if err != nil {
		log.Error("strategy_config_unreadable", zap.Error(err))
		return
	}

	if latestCents <= 0 {
		log.Warn("no_latest_price_skipping")
		return
	}
	price := float64(latestCents) / 100

	window := append(append([]Bar{}, bars...), Bar{
		Open: price, High: price, Low: price, Close: price,
	})
	if len(window) < 2 {
		log.Warn("not_enough_bars_skipping")
		return
	}

	if !dep.InPosition {
		enter, err := Evaluate(cfg.Entry, window)
		if err != nil {
			log.Error("entry_evaluation_failed", zap.Error(err))
			return
		}
		if !enter {
			return
		}
		qty := dep.PositionSizeCents / latestCents
		if qty < 1 {
			log.Warn("position_size_below_one_share", zap.Int64("price_cents", latestCents))
			return
		}
		if e.placeOrder(ctx, dep, ordrpb.Side_BUY, qty, latestCents) {
			e.persistPosition(ctx, dep.ID, true, latestCents, qty)
			log.Info("strategy_entered", zap.Int64("qty", qty), zap.Int64("price_cents", latestCents))
		}
		return
	}

	entry := float64(dep.EntryPriceCents) / 100
	exitNow := false
	switch {
	case cfg.StopLossPct > 0 && price <= entry*(1-cfg.StopLossPct/100):
		exitNow = true
		log.Info("stop_loss_hit")
	case cfg.TakeProfitPct > 0 && price >= entry*(1+cfg.TakeProfitPct/100):
		exitNow = true
		log.Info("take_profit_hit")
	default:
		exitNow, err = Evaluate(cfg.Exit, window)
		if err != nil {
			log.Error("exit_evaluation_failed", zap.Error(err))
			return
		}
	}
	if !exitNow {
		return
	}

	if e.placeOrder(ctx, dep, ordrpb.Side_SELL, dep.Qty, latestCents) {
		e.persistPosition(ctx, dep.ID, false, 0, 0)
		log.Info("strategy_exited", zap.Int64("qty", dep.Qty), zap.Int64("price_cents", latestCents))
	}
}

func (e *Engine) placeOrder(ctx context.Context, dep db.GetRunningDeploymentsRow, side ordrpb.Side, qty, refPriceCents int64) bool {
	_, err := e.order.PlaceOrder(ctx, &ordrpb.PlaceOrderRequest{
		UserId:              dep.UserID.String(),
		Symbol:              dep.Symbol,
		Side:                side,
		OrderType:           ordrpb.OrderType_MARKET,
		TimeInForce:         ordrpb.TimeInForce_GTC,
		Qty:                 qty,
		ReferencePriceCents: refPriceCents,
		StrategyId:          dep.StrategyID.String(),
	})
	if err != nil {
		e.log.Error("strategy_order_failed",
			logger.KV("deployment_id", dep.ID.String()),
			logger.KV("symbol", dep.Symbol),
			zap.Error(err),
		)
		return false
	}
	telemetry.StrategyEngineOrdersTotal.WithLabelValues(telemetry.SideLabel(int32(side))).Inc()
	return true
}

func (e *Engine) persistPosition(ctx context.Context, depID uuid.UUID, inPosition bool, entryCents, qty int64) {
	if err := e.q.UpdateDeploymentPosition(ctx, db.UpdateDeploymentPositionParams{
		ID:              depID,
		InPosition:      inPosition,
		EntryPriceCents: entryCents,
		Qty:             qty,
	}); err != nil {
		e.log.Error("deployment_state_persist_failed", zap.Error(err))
	}
}

func (e *Engine) fetchBars(ctx context.Context, symbols []string) map[string][]Bar {
	out := make(map[string][]Bar, len(symbols))
	if len(symbols) == 0 {
		return out
	}

	start := time.Now().UTC().AddDate(0, 0, -2)
	resp, err := e.mrkt.GetHistoricalStockData(ctx, &mrktpb.HistoricalStockDataRequest{
		Symbols:   symbols,
		Timeframe: mrktpb.Timeframe_MIN,
		Start: &mrktpb.Date{
			Year:  int32(start.Year()),
			Month: int32(start.Month()),
			Day:   int32(start.Day()),
		},
	})
	if err != nil {
		e.log.Error("engine_bars_fetch_failed", zap.Error(err))
		return out
	}

	for _, sb := range resp.SymbolBars {
		bars := make([]Bar, 0, len(sb.Bars))
		for _, b := range sb.Bars {
			bars = append(bars, Bar{
				Open:   float64(b.Open),
				High:   float64(b.High),
				Low:    float64(b.Low),
				Close:  float64(b.Close),
				Volume: float64(b.Volume),
				VWAP:   float64(b.Vwap),
			})
		}
		out[sb.Symbol] = bars
	}
	return out
}

func (e *Engine) fetchLatestPrices(ctx context.Context, symbols []string) map[string]int64 {
	out := make(map[string]int64, len(symbols))
	if len(symbols) == 0 {
		return out
	}
	resp, err := e.mrkt.GetLatestPrices(ctx, &mrktpb.LatestPricesRequest{Symbols: symbols})
	if err != nil {
		e.log.Error("engine_prices_fetch_failed", zap.Error(err))
		return out
	}
	for _, p := range resp.Prices {
		out[p.Symbol] = p.PriceCents
	}
	return out
}

func uniqueSymbols(deps []db.GetRunningDeploymentsRow) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, d := range deps {
		if _, ok := seen[d.Symbol]; !ok {
			seen[d.Symbol] = struct{}{}
			out = append(out, d.Symbol)
		}
	}
	return out
}

func marketOpen(now time.Time) bool {
	if os.Getenv("GLYPH_TRADING_247") == "true" {
		return true
	}
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.FixedZone("EST", -5*60*60)
	}
	ny := now.In(loc)
	switch ny.Weekday() {
	case time.Saturday, time.Sunday:
		return false
	}
	minutes := ny.Hour()*60 + ny.Minute()
	return minutes >= 9*60+30 && minutes < 16*60
}
