package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SignupsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "glyph_signups_total",
		Help: "Completed user signups.",
	})

	SigninsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "glyph_signins_total",
		Help: "Successful user signins.",
	})

	WSConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "glyph_ws_connections_active",
		Help: "Currently open market-data websocket connections.",
	})

	OrdersPlacedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "glyph_orders_placed_total",
		Help: "Orders accepted by the order service, by side.",
	}, []string{"side"})

	FillsAppliedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "glyph_fills_applied_total",
		Help: "Fills settled by the consumer, by side.",
	}, []string{"side"})

	FillApplyLagSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "glyph_fill_apply_lag_seconds",
		Help:    "Delay between a fill's execution time and the settlement consumer applying it.",
		Buckets: prometheus.DefBuckets,
	})

	StrategyTicksTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "glyph_strategy_ticks_total",
		Help: "Strategy-engine deployment evaluations.",
	})

	StrategyEngineOrdersTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "glyph_strategy_engine_orders_total",
		Help: "Orders placed by the strategy engine, by side.",
	}, []string{"side"})

	StrategyDeploymentsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "glyph_strategy_deployments_total",
		Help: "Strategy deployments started.",
	})

	BacktestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "glyph_backtests_total",
		Help: "Backtests run.",
	})
)

func SideLabel(side int32) string {
	if side == 1 {
		return "sell"
	}
	return "buy"
}
